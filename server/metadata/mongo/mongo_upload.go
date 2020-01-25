package mongo

import (
	"errors"



	"github.com/root-gg/plik/server/common"
	"gopkg.in/mgo.v2/bson"
)

// Create upload
func (b *Backend) CreateUpload(upload *common.Upload) (u *common.Upload, err error) {
	if id == "" {
		return nil, errors.New("Unable to get upload : Missing upload id")
	}

	session := b.session.Copy()
	defer session.Close()
	collection := session.DB(b.config.Database).C(b.config.Collection)
	u = &common.Upload{}
	err = collection.Insert(&upload)


	if err != nil {
		err = log.EWarningf("Unable to get metadata from mongodb : %s", err)
	}
	return
}

// Get implementation from MongoDB Metadata Backend
func (b *Backend) GetUpload(id string) (u *common.Upload, err error) {
	if id == "" {
		err = log.EWarning("Unable to `" +
			"" +
			"" +
			"" +
			"" +
			" upload : Missing upload id")
		return
	}

	session := b.session.Copy()
	defer session.Close()
	collection := session.DB(b.config.Database).C(b.config.Collection)
	u = &common.Upload{}
	err = collection.Find(bson.M{"id": id}).One(u)
	if err != nil {
		err = log.EWarningf("Unable to get metadata from mongodb : %s", err)
	}
	return
}

// Upsert implementation from MongoDB Metadata Backend
func (b *Backend) UpdateUpload(upload *common.Upload, uploadTx common.UploadTx) (err error) {
	if upload == nil {
		err = log.EWarning("Unable to save upload : Missing upload")
		return
	}

	session := b.session.Copy()
	defer session.Close()
	collection := session.DB(b.config.Database).C(b.config.Collection)
	_, err = collection.Upsert(bson.M{"id": upload.ID}, &upload)
	if err != nil {
		err = log.EWarningf("Unable to append metadata to mongodb : %s", err)
	}
	return
}

// Remove implementation from MongoDB Metadata Backend
func (b *Backend) RemoveUpload(upload *common.Upload) (err error) {
	if upload == nil {
		err = log.EWarning("Unable to remove upload : Missing upload")
		return
	}

	session := b.session.Copy()
	defer session.Close()
	collection := session.DB(b.config.Database).C(b.config.Collection)
	err = collection.Remove(bson.M{"id": upload.ID})
	if err != nil {
		err = log.EWarningf("Unable to remove upload from mongodb : %s", err)
	}
	return
}