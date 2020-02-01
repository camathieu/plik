package mongo

import (
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"gopkg.in/mgo.v2/bson"

	"github.com/root-gg/plik/server/common"
)

// CreateUpload create upload metadata in mongodb
func (b *Backend) CreateUpload(upload *common.Upload) (err error) {
	if upload == nil {
		return errors.New("missing upload")
	}

	ctx, cancel := b.newContext()
	defer cancel()

	_, err = b.uploadCollection.InsertOne(ctx, upload)

	return err
}

// GetUpload upload metadata from mongodb
func (b *Backend) GetUpload(ID string) (upload *common.Upload, err error) {
	if ID == "" {
		return nil, errors.New("missing upload id")
	}

	ctx, cancel := b.newContext()
	defer cancel()

	result := b.uploadCollection.FindOne(ctx, bson.M{"id": ID})
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, result.Err()
	}

	upload = &common.Upload{}
	err = result.Decode(&upload)
	if err != nil {
		return nil, err
	}

	return upload, nil
}

// UpdateUpload update upload metadata in mongodb
func (b *Backend) UpdateUpload(upload *common.Upload, uploadTx common.UploadTx) (u *common.Upload, err error) {
	ctx, cancel := b.newContext()
	defer cancel()

	// Prepare upload update transaction
	updateUploadTx := func(sctx mongo.SessionContext) error {
		err := sctx.StartTransaction(options.Transaction().
			SetReadConcern(readconcern.Snapshot()).
			SetWriteConcern(writeconcern.New(writeconcern.WMajority())),
		)
		if err != nil {
			return err
		}

		// Abort transaction
		defer func() { _ = sctx.AbortTransaction(sctx) }()

		upload = &common.Upload{}
		result := b.uploadCollection.FindOne(ctx, bson.M{"id": upload.ID})
		if result.Err() != nil {
			if result.Err() == mongo.ErrNoDocuments {
				// Upload not found ( maybe it has been removed in the mean time )
				// Let the upload tx set the (HTTP) error and forward it
				err = uploadTx(nil)
				if err != nil {
					return err
				}
				return fmt.Errorf("Upload tx without an upload shoud have returned an error")
			}
			return result.Err()
		}

		// Get upload
		u = &common.Upload{}
		err = result.Decode(&upload)
		if err != nil {
			return err
		}

		// Apply transaction ( mutate )
		err = uploadTx(u)
		if err != nil {
			return err
		}

		// Avoid the possibility to override an other upload by changing the upload.ID in the tx
		replaceResult, err := b.uploadCollection.ReplaceOne(sctx, bson.M{"id": upload.ID}, u)
		if err != nil {
			return err
		}
		if replaceResult.ModifiedCount != 1 {
			return fmt.Errorf("ReplaceOne should have updated exactly one mongodb document but has updated %d", replaceResult.ModifiedCount)
		}

		return commitWithRetry(sctx)
	}

	// Execute transaction with automatic retries and timeout
	err = b.client.UseSessionWithOptions(
		ctx, options.Session().SetDefaultReadPreference(readpref.Primary()),
		func(sctx mongo.SessionContext) error {
			return runTransactionWithRetry(sctx, updateUploadTx)
		},
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// RemoveUpload remove upload metadata in mongodb
func (b *Backend) RemoveUpload(upload *common.Upload) (err error) {
	if upload == nil {
		return errors.New("missing upload")
	}

	ctx, cancel := b.newContext()
	defer cancel()

	upload = &common.Upload{}
	collection := b.database.Collection(b.config.UploadCollection)
	_, err = collection.DeleteOne(ctx, bson.M{"id": upload.ID})

	return err
}
