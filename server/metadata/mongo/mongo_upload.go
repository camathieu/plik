package mongo

import (
	"errors"
	"fmt"
	"github.com/root-gg/plik/server/common"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
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

var errRetry = errors.New("retry tx")

func (b *Backend) AddOrUpdateFile(upload *common.Upload, file *common.File, status string) (err error) {
	if upload == nil {
		return errors.New("missing upload")
	}

	if file == nil {
		return errors.New("missing file")
	}

	ctx, cancel := b.newContext()
	defer cancel()

	if status == "" {

		// Insert file
		query := bson.M{"id": upload.ID, fmt.Sprintf("files.%s", file.ID): nil}
		update := bson.M{"$set": bson.M{fmt.Sprintf("files.%s", file.ID): file}}
		updateResult, err := b.uploadCollection.UpdateOne(ctx, query, update)
		if err != nil {
			return err
		}

		if updateResult.ModifiedCount != 1 {
			return fmt.Errorf("upload does not exist anymore or invalid file status")
		}

	} else {

		// Update file
		query := bson.M{"id": upload.ID, fmt.Sprintf("files.%s.status", file.ID): status}
		update := bson.M{"$set": bson.M{fmt.Sprintf("files.%s", file.ID): file}}
		updateResult, err := b.uploadCollection.UpdateOne(ctx, query, update)
		if err != nil {
			return err
		}

		if updateResult.ModifiedCount != 1 {
			return fmt.Errorf("upload does not exist anymore or invalid file status")
		}

	}

	return nil
}

// UpdateUpload update upload metadata in mongodb
//func (b *Backend) UpdateUpload(upload *common.Upload, uploadTx common.UploadTx) (u *common.Upload, err error) {
//	if upload == nil {
//		return nil, errors.New("missing upload")
//	}
//
//	// TODO Change method signature to take only uploadID
//	uploadId := upload.ID
//
//	tx := func() (err error){
//		ctx, cancel := b.newContext()
//		defer cancel()
//
//		// Fetch upload
//		result := b.uploadCollection.FindOne(ctx, bson.M{"id": uploadId})
//		if result.Err() != nil {
//			if result.Err() == mongo.ErrNoDocuments {
//				// Upload not found ( maybe it has been removed in the mean time )
//				// Let the upload tx set the (HTTP) error and forward it
//				err = uploadTx(nil)
//				if err != nil {
//					return err
//				}
//				return fmt.Errorf("upload tx without an upload should return an error")
//			}
//			return result.Err()
//		}
//
//		// Decode upload
//		u = &common.Upload{}
//		err = result.Decode(&u)
//		if err != nil {
//			return err
//		}
//
//		version := u.Version
//
//		// Apply transaction ( mutate )
//		err = uploadTx(u)
//		if err != nil {
//			return err
//		}
//
//		u.Version = version + 1
//
//		// Avoid the possibility to override an other upload by changing the upload.ID in the tx
//		replaceResult, err := b.uploadCollection.ReplaceOne(ctx, bson.M{"id": uploadId, "version": version}, u, options.Replace().SetUpsert(false))
//		if err != nil {
//			return err
//		}
//
//		if replaceResult.ModifiedCount == 1 {
//			return errRetry
//		}
//		return
//	}
//
//	for {
//		err = tx()
//
//		if err == nil {
//			return u, nil
//		}
//
//		if err == errRetry {
//			fmt.Printf("replaceOne should have updated exactly one mongodb document but has updated %d\n", replaceResult.ModifiedCount)
//			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
//			continue
//		}
//
//		return nil, err
//	}

// Prepare upload update transaction
//updateUploadTx := func(sctx mongo.SessionContext) error {
//	err := sctx.StartTransaction(options.Transaction().
//		SetReadPreference(readpref.Primary()).
//		SetReadConcern(readconcern.Snapshot()).
//		SetWriteConcern(writeconcern.New(writeconcern.WMajority())),
//	)
//	if err != nil {
//		return err
//	}
//
//	// Abort transaction
//	defer func() { _ = sctx.AbortTransaction(sctx) }()
//
//	// Acquire the lock
//	// https://www.mongodb.com/blog/post/how-to-select--for-update-inside-mongodb-transactions
//	uuid := bson.NewObjectId()
//	lock := bson.M{ "$set": bson.M{ "lock": uuid }}
//
//	// Fetch upload
//	result := b.uploadCollection.FindOneAndUpdate(ctx, bson.M{"id": upload.ID}, lock, options.FindOneAndUpdate().SetReturnDocument(options.After))
//	if result.Err() != nil {
//		if result.Err() == mongo.ErrNoDocuments {
//			// Upload not found ( maybe it has been removed in the mean time )
//			// Let the upload tx set the (HTTP) error and forward it
//			err = uploadTx(nil)
//			if err != nil {
//				return err
//			}
//			return fmt.Errorf("upload tx without an upload should return an error")
//		}
//		return result.Err()
//	}
//
//	// Decode upload
//	u = &common.Upload{}
//	err = result.Decode(&u)
//	if err != nil {
//		return err
//	}
//
//	// Apply transaction ( mutate )
//	err = uploadTx(u)
//	if err != nil {
//		return err
//	}
//
//	// Avoid the possibility to override an other upload by changing the upload.ID in the tx
//	replaceResult, err := b.uploadCollection.ReplaceOne(sctx, bson.M{"id": upload.ID, "lock": uuid}, u)
//	if err != nil {
//		return err
//	}
//	if replaceResult.ModifiedCount != 1 {
//		return fmt.Errorf("replaceOne should have updated exactly one mongodb document but has updated %d", replaceResult.ModifiedCount)
//	}
//
//	return commitWithRetry(sctx)
//}
//
//// Execute transaction with automatic retries and timeout
//err = b.client.UseSessionWithOptions(
//	ctx, options.Session().SetCausalConsistency(true),
//	func(sctx mongo.SessionContext) error {
//		return runTransactionWithRetry(sctx, updateUploadTx)
//	},
//)
//	if err != nil {
//		return nil, err
//	}
//	return u, nil
//}

// RemoveUpload remove upload metadata in mongodb
func (b *Backend) RemoveUpload(upload *common.Upload) (err error) {
	if upload == nil {
		return errors.New("missing upload")
	}

	ctx, cancel := b.newContext()
	defer cancel()

	_, err = b.uploadCollection.DeleteOne(ctx, bson.M{"id": upload.ID})

	return err
}
