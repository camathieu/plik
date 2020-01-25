package mongo

import (
	"github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// SaveUser implementation from MongoDB Metadata Backend
func (b *Backend) SaveUser(user *common.User) (err error) {
	if user == nil {
		err = log.EWarning("Unable to save user : Missing user")
		return
	}

	session := b.session.Copy()
	defer session.Close()
	collection := session.DB(b.config.Database).C(b.config.UserCollection)

	_, err = collection.Upsert(bson.M{"id": user.ID}, &user)
	if err != nil {
		err = log.EWarningf("Unable to save user to mongodb : %s", err)
	}
	return
}

// GetUser implementation from MongoDB Metadata Backend
func (b *Backend) GetUser(id string, token string) (user *common.User, err error) {
	if id == "" && token == "" {
		err = log.EWarning("Unable to get user : Missing user id or token")
		return
	}

	session := b.session.Copy()
	defer session.Close()
	collection := session.DB(b.config.Database).C(b.config.UserCollection)

	user = &common.User{}
	if id != "" {
		err = collection.Find(bson.M{"id": id}).One(user)
		if err == mgo.ErrNotFound {
			return nil, nil
		} else if err != nil {
			err = log.EWarningf("Unable to get user from mongodb : %s", err)
		}
	} else if token != "" {
		err = collection.Find(bson.M{"tokens.token": token}).One(user)
		if err == mgo.ErrNotFound {
			return nil, nil
		} else if err != nil {
			err = log.EWarningf("Unable to get user from mongodb : %s", err)
		}
	} else {
		err = log.EWarning("Unable to get user from mongodb : Missing user id or token")
	}

	return
}

// RemoveUser implementation from MongoDB Metadata Backend
func (b *Backend) RemoveUser(user *common.User) (err error) {
	if user == nil {
		err = log.EWarning("Unable to remove user : Missing user")
		return
	}

	session := b.session.Copy()
	defer session.Close()

	collection := session.DB(b.config.Database).C(b.config.UserCollection)

	err = collection.Remove(bson.M{"id": user.ID})
	if err != nil {
		err = log.EWarningf("Unable to remove user from mongodb : %s", err)
	}

	return
}