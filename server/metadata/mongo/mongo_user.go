package mongo

import (
	"errors"
	"fmt"

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

	err = b.client.UseSession(ctx, func(sessionContext mongo.SessionContext) error {
		err = sessionContext.StartTransaction()
		if err != nil {
			return err
		}

		u := &common.User{}
		err := b.userCollection.FindOne(sessionContext, bson.M{"id": user.ID}).Decode(&u)
		if err != nil {
			err2 := sessionContext.AbortTransaction(sessionContext)
			if err2 != nil {
				return fmt.Errorf("%s : %s", err, err2)
			}
			return err
		}

		err = userTx(u)
		if err != nil {
			err2 := sessionContext.AbortTransaction(sessionContext)
			if err2 != nil {
				return fmt.Errorf("%s : %s", err, err2)
			}
			return err
		}

		// Avoid the possibility to override an other upload by changing the upload.ID in the tx
		_, err = b.userCollection.ReplaceOne(sessionContext, bson.M{"id": user.ID}, u)
		if err != nil {
			err2 := sessionContext.AbortTransaction(sessionContext)
			if err2 != nil {
				return fmt.Errorf("%s : %s", err, err2)
			}
			return err
		}

		return sessionContext.CommitTransaction(sessionContext)
	})

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
