package mongo

import (
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"

	"github.com/root-gg/plik/server/common"
)

// CreateUser create user metadata in mongodb
func (b *Backend) CreateUser(user *common.User) (err error) {
	if user == nil {
		return errors.New("missing user")
	}

	ctx, cancel := b.newContext()
	defer cancel()

	_, err = b.userCollection.InsertOne(ctx, user)

	return err
}

// GetUser get user metadata from mongodb
func (b *Backend) GetUser(userID string) (user *common.User, err error) {
	if userID == "" {
		return nil, errors.New("missing user id")
	}

	ctx, cancel := b.newContext()
	defer cancel()

	user = &common.User{}
	err = b.userCollection.FindOne(ctx, bson.M{"id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserFromToken get user metadata from mongodb
func (b *Backend) GetUserFromToken(token string) (user *common.User, err error) {
	if token == "" {
		return nil, errors.New("missing user token")
	}

	ctx, cancel := b.newContext()
	defer cancel()

	user = &common.User{}
	err = b.userCollection.FindOne(ctx, bson.M{"tokens.token": token}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser update user metadata in mongodb
func (b *Backend) UpdateUser(user *common.User, userTx common.UserTx) (u *common.User, err error) {
	if user == nil {
		return nil, errors.New("missing user")
	}

	ctx, cancel := b.newContext()
	defer cancel()

	// Prepare upload update transaction
	updateUsetTx := func(sctx mongo.SessionContext) error {
		err := sctx.StartTransaction(options.Transaction().
			SetReadConcern(readconcern.Snapshot()).
			SetWriteConcern(writeconcern.New(writeconcern.WMajority())),
		)
		if err != nil {
			return err
		}

		// Abort transaction
		defer func() { _ = sctx.AbortTransaction(sctx) }()

		// Fetch user
		u := &common.User{}
		result := b.uploadCollection.FindOne(ctx, bson.M{"id": user.ID})
		if result.Err() != nil {
			if result.Err() == mongo.ErrNoDocuments {
				// User not found ( maybe it has been removed in the mean time )
				// Let the user tx set the (HTTP) error and forward it
				err = userTx(nil)
				if err != nil {
					return err
				}
				return fmt.Errorf("user tx without an user should return an error")
			}
			return result.Err()
		}

		// Decode user
		u = &common.User{}
		err = result.Decode(&user)
		if err != nil {
			return err
		}

		// Apply transaction ( mutate )
		err = userTx(u)
		if err != nil {
			return err
		}

		// Avoid the possibility to override an other user by changing the user.ID in the tx
		replaceResult, err := b.userCollection.ReplaceOne(sctx, bson.M{"id": user.ID}, u)
		if err != nil {
			return err
		}
		if replaceResult.ModifiedCount != 1 {
			return fmt.Errorf("replaceOne should have updated exactly one mongodb document but has updated %d", replaceResult.ModifiedCount)
		}

		return commitWithRetry(sctx)
	}

	// Execute transaction with automatic retries and timeout
	err = b.client.UseSessionWithOptions(
		ctx, options.Session().SetDefaultReadPreference(readpref.Primary()),
		func(sctx mongo.SessionContext) error {
			return runTransactionWithRetry(sctx, updateUsetTx)
		},
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// RemoveUser remove user metadata from mongodb
func (b *Backend) RemoveUser(user *common.User) (err error) {
	if user == nil {
		return errors.New("missing user")
	}

	ctx, cancel := b.newContext()
	defer cancel()

	user = &common.User{}
	collection := b.database.Collection(b.config.UserCollection)
	_, err = collection.DeleteOne(ctx, bson.M{"id": user.ID})

	return err
}

// GetUserUploads remove user metadata from mongodb
func (b *Backend) GetUserUploads(user *common.User, token *common.Token) (ids []string, err error) {
	panic("Not Yet Implemented")
}

// GetUserStatistics get user/token statistics from mongodb
func (b *Backend) GetUserStatistics(user *common.User, token *common.Token) (stats *common.UserStats, err error) {
	panic("Not Yet Implemented")
}
