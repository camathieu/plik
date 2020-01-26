package mongo

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"

	"github.com/root-gg/plik/server/common"
)

// CreateUpload create upload metadata in mongodb
func (b *Backend) CreateUpload(upload *common.Upload) (err error) {
	if upload == nil {
		return errors.New("missing upload")
	}

	ctx, cancel := newContext()
	defer cancel()

	_, err = b.uploadCollection.InsertOne(ctx, upload)

	return err
}

// GetUpload upload metadata from mongodb
func (b *Backend) GetUpload(ID string) (upload *common.Upload, err error) {
	if ID == "" {
		return nil, errors.New("missing upload id")
	}

	ctx, cancel := newContext()
	defer cancel()

	upload = &common.Upload{}
	err = b.uploadCollection.FindOne(ctx, bson.M{"id": ID}).Decode(&upload)
	if err != nil {
		return nil, err
	}

	return upload, nil
}

// UpdateUpload update upload metadata in mongodb
func (b *Backend) UpdateUpload(upload *common.Upload, uploadTx common.UploadTx) (u *common.Upload, err error) {
	if upload == nil {
		return nil, errors.New("missing upload")
	}

	ctx, cancel := newContext()
	defer cancel()

	err = b.client.UseSession(ctx, func(sessionContext mongo.SessionContext) error {
		err = sessionContext.StartTransaction()
		if err != nil {
			return err
		}

		u = &common.Upload{}
		err := b.uploadCollection.FindOne(sessionContext, bson.M{"id": upload.ID}).Decode(&u)
		if err != nil {
			err2 := sessionContext.AbortTransaction(sessionContext)
			if err2 != nil {
				return fmt.Errorf("%s : %s", err, err2)
			}
			return err
		}

		err = uploadTx(u)
		if err != nil {
			err2 := sessionContext.AbortTransaction(sessionContext)
			if err2 != nil {
				return fmt.Errorf("%s : %s", err, err2)
			}
			return err
		}

		// Avoid the possibility to override an other upload by changing the upload.ID in the tx
		_, err = b.uploadCollection.ReplaceOne(sessionContext, bson.M{"id": upload.ID}, u)
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

// RemoveUpload remove upload metadata in mongodb
func (b *Backend) RemoveUpload(upload *common.Upload) (err error) {
	if upload == nil {
		return errors.New("missing upload")
	}

	ctx, cancel := newContext()
	defer cancel()

	upload = &common.Upload{}
	collection := b.database.Collection(b.config.UploadCollection)
	_, err = collection.DeleteOne(ctx, bson.M{"id": upload.ID})

	return err
}
