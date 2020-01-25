/**

    Plik upload server

The MIT License (MIT)

Copyright (c) <2015>
	- Mathieu Bodjikian <mathieu@bodjikian.fr>
	- Charles-Antoine Mathieu <skatkatt@root.gg>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
**/

package mongo

import (
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"github.com/root-gg/plik/server/metadata"
	"github.com/root-gg/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
 * User input is only safe in document field !!!
 * Keys with ( '.', '$', ... ) may be interpreted
 */

// Ensure Mongo Metadata Backend implements metadata.Backend interface
var _ metadata.Backend = (*Backend)(nil)

// Config object
type Config struct {
	URL            string
	Database       string
	Collection     string
	UserCollection string
	Username       string
	Password       string
	Ssl            bool
}

// NewConfig configures the backend
// from config passed as argument
func NewConfig(params map[string]interface{}) (c *Config) {
	c = new(Config)
	c.URL = "127.0.0.1:27017"
	c.Database = "plik"
	c.Collection = "meta"
	c.UserCollection = "tokens"
	utils.Assign(c, params)
	return
}

// Backend object
type Backend struct {
	config  *Config
	session *mgo.Session
}

// NewBackend instantiate a new MongoDB Metadata Backend
// from configuration passed as argument
func NewBackend(config *Config) (b *Backend, err error) {
	b = new(Backend)
	b.config = config

	// Open connection
	dialInfo := &mgo.DialInfo{}
	dialInfo.Addrs = []string{b.config.URL}
	dialInfo.Database = b.config.Database
	dialInfo.Timeout = 5 * time.Second
	if b.config.Username != "" && b.config.Password != "" {
		dialInfo.Username = b.config.Username
		dialInfo.Password = b.config.Password
	}
	if b.config.Ssl {
		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", addr.String(), &tls.Config{InsecureSkipVerify: true})
		}
	}

	// TODO use logger or move
	fmt.Printf("Connecting to mongodb @ %s/%s", b.config.URL, b.config.Database)

	b.session, err = mgo.DialWithInfo(dialInfo)
	if err != nil {
		return nil, fmt.Errorf("Unable to contact mongodb at %s : %s", b.config.URL, err.Error())
	}

	// TODO use logger or move
	fmt.Printf("Connected to mongodb @ %s/%s", b.config.URL, b.config.Database)

	// Ensure log.Infof(everything is persisted and replicated
	b.session.SetMode(mgo.Strong, false)
	b.session.SetSafe(&mgo.Safe{})

	return b, nil
}



// GetUserUploads implementation from MongoDB Metadata Backend
func (b *Backend) GetUserUploads(ctx *juliet.Context, user *common.User, token *common.Token) (ids []string, err error) {
	log := context.GetLogger(ctx)

	if user == nil {
		err = log.EWarning("Unable to get user uploads : Missing user")
		return
	}

	session := b.session.Copy()
	defer session.Close()
	collection := session.DB(b.config.Database).C(b.config.Collection)

	b := bson.M{"user": user.ID}
	if token != nil {
		b["token"] = token.Token
	}

	var uploads []*common.Upload
	err = collection.Find(b).Select(bson.M{"id": 1}).Sort("-uploadDate").All(&uploads)
	if err != nil {
		err = log.EWarningf("Unable to get user uploads : %s", err)
		return
	}

	// Get all ids
	for _, upload := range uploads {
		ids = append(ids, upload.ID)
	}

	return
}

// GetUploadsToRemove implementation from MongoDB Metadata Backend
func (b *Backend) GetUploadsToRemove(ctx *juliet.Context) (ids []string, err error) {
	log := context.GetLogger(ctx)

	session := b.session.Copy()
	defer session.Close()
	collection := session.DB(b.config.Database).C(b.config.Collection)

	// Look for expired uploads
	var uploads []*common.Upload
	b := bson.M{"$where": "this.ttl > 0 && " + strconv.Itoa(int(time.Now().Unix())) + " > this.uploadDate + this.ttl"}

	err = collection.Find(b).Select(bson.M{"id": 1}).All(&uploads)
	if err != nil {
		err = log.EWarningf("Unable to get uploads to remove : %s", err)
		return
	}

	// Get all ids
	for _, upload := range uploads {
		ids = append(ids, upload.ID)
	}

	return
}

// GetUserStatistics implementation for MongoDB Metadata Backend
func (b *Backend) GetUserStatistics(ctx *juliet.Context, user *common.User, token *common.Token) (stats *common.UserStats, err error) {
	log := context.GetLogger(ctx)

	if user == nil {
		err = log.EWarning("Unable to get user uploads : Missing user")
		return
	}

	session := b.session.Copy()
	defer session.Close()
	collection := session.DB(b.config.Database).C(b.config.Collection)

	match := bson.M{"user": user.ID}
	if token != nil {
		match["token"] = token.Token
	}

	// db.plik_meta.aggregate([{$match: {user:"xxx", token:"xxx"}}, {$project: {"files": {$objectToArray: "$files"}}}, {$unwind: "$files"}, {$group: { _id: null, count: {$sum: 1}, total: {$sum: "$files.v.fileSize"}, uploads: {$addToSet: "$_id"}}}, {$project: {count: "$count", total: "$total", size: { $size: "$uploads"}}}]).pretty()
	pipeline := []bson.M{
		{"$match": match},
		{"$project": bson.M{"file_count": bson.M{"$size": bson.M{"$objectToArray": "$files"}}}},
		{"$unwind": "$files"},
		{"$group": bson.M{"_id": nil, "files": bson.M{"$sum": 1}, "totalSize": bson.M{"$sum": "$files.v.fileSize"}, "uploads": bson.M{"$addToSet": "$_id"}}},
		{"$project": bson.M{"Files": "$files", "TotalSize": "$totalSize", "Uploads": bson.M{"$size": "uploads"}}},
	}

	stats = new(common.UserStats)
	err = collection.Pipe(pipeline).One(&stats)
	if err != nil {
		err = log.EWarningf("Unable to get file count from mongodb : %s", err)
	}

	return
}

// GetUsers implementation for MongoDB Metadata Backend
func (b *Backend) GetUsers(ctx *juliet.Context) (ids []string, err error) {
	log := context.GetLogger(ctx)

	session := b.session.Copy()
	defer session.Close()
	collection := session.DB(b.config.Database).C(b.config.UserCollection)

	var results []struct {
		ID string `bson:"id"`
	}
	err = collection.Find(nil).Select(bson.M{"id": 1}).Sort("id").All(&results)
	if err != nil {
		err = log.EWarningf("Unable to get users from mongodb : %s", err)
	}

	for _, result := range results {
		ids = append(ids, result.ID)
	}

	return
}

// GetServerStatistics implementation for MongoDB Metadata Backend
func (b *Backend) GetServerStatistics(ctx *juliet.Context) (stats *common.ServerStats, err error) {
	log := context.GetLogger(ctx)
	config := context.GetConfig(ctx)

	stats = new(common.ServerStats)
	session := b.session.Copy()
	defer session.Close()

	uploadCollection := session.DB(b.config.Database).C(b.config.Collection)

	// NUMBER OF UPLOADS
	uploadCount, err := uploadCollection.Find(nil).Count()
	if err != nil {
		err = log.EWarningf("Unable to get upload count from mongodb : %s", err)
	}
	stats.Uploads = uploadCount

	// NUMBER OF ANONYMOUS UPLOADS
	anonymousUploadCount, err := uploadCollection.Find(bson.M{"user": ""}).Count()
	if err != nil {
		err = log.EWarningf("Unable to get anonymous upload count from mongodb : %s", err)
	}
	stats.AnonymousUploads = anonymousUploadCount

	// NUMBER OF FILES
	//db.plik_meta.aggregate([{$project: {"file_count": { $size: { $objectToArray : "$files" } }}}, { $group: { _id : null, total : { $sum : "$file_count" }}}]).pretty()

	pipeline1 := []bson.M{
		{"$project": bson.M{"file_count": bson.M{"$size": bson.M{"$objectToArray": "$files"}}}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": "$file_count"}}},
	}

	var result1 struct {
		Total int `bson:"total"`
	}

	err = uploadCollection.Pipe(pipeline1).One(&result1)
	if err != nil {
		err = log.EWarningf("Unable to get file count from mongodb : %s", err)
	}

	stats.Files = result1.Total

	// TOTAL SIZE OF ALL FILES
	// db.plik_meta.aggregate([{$project: {"files": { $objectToArray : "$files" } }}, {$unwind: "$files"}, {$group : { _id: null, total : { $sum : "$files.v.fileSize"} }}]).pretty()

	pipeline2 := []bson.M{
		{"$project": bson.M{"files": bson.M{"$objectToArray": "$files"}}},
		{"$unwind": "$files"},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": "$files.v.fileSize"}}},
	}

	var result2 struct {
		Total int64 `bson:"total"`
	}

	err = uploadCollection.Pipe(pipeline2).One(&result2)
	if err != nil {
		err = log.EWarningf("Unable to get total file size from mongodb : %s", err)
	}

	stats.TotalSize = result2.Total

	if !config.NoAnonymousUploads {

		// TOTAL SIZE OF ALL ANONYMOUS UPLOAD FILES
		// db.plik_meta.aggregate([{$match: {user:""}},{$project: {"files": { $objectToArray : "$files" } }}, {$unwind: "$files"}, {$group : { _id: null, total : { $sum : "$files.v.fileSize"} }}]).pretty()

		pipeline3 := []bson.M{
			{"$match": bson.M{"user": ""}},
			{"$project": bson.M{"files": bson.M{"$objectToArray": "$files"}}},
			{"$unwind": "$files"},
			{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": "$files.v.fileSize"}}},
		}

		var result3 struct {
			Total int64 `bson:"total"`
		}

		err = uploadCollection.Pipe(pipeline3).One(&result3)
		if err != nil {
			err = log.EWarningf("Unable to get total file size from mongodb : %s", err)
		}

		stats.AnonymousSize = result3.Total
	}

	// TOTAL FILE SIZE BY FILE TYPE
	// db.plik_meta.aggregate([{$project: {"files": { $objectToArray : "$files" } }}, {$unwind: "$files"}, {$group : { _id: "$files.v.fileType", total : { $sum : "$files.v.fileSize"} }},{ $sort : { total : -1 }},{ $limit : 5 }]).pretty()

	pipeline4 := []bson.M{
		{"$project": bson.M{"files": bson.M{"$objectToArray": "$files"}}},
		{"$unwind": "$files"},
		{"$group": bson.M{"_id": "$files.v.fileType", "total": bson.M{"$sum": 1}}},
		{"$sort": bson.M{"total": -1}},
		{"$limit": 10},
	}

	var result4 []common.FileTypeByCount

	err = uploadCollection.Pipe(pipeline4).All(&result4)
	if err != nil {
		err = log.EWarningf("Unable to get total file size from mongodb : %s", err)
	}

	stats.FileTypeByCount = result4

	// TOTAL FILE SIZE BY FILE TYPE
	// db.plik_meta.aggregate([{$project: {"files": { $objectToArray : "$files" } }}, {$unwind: "$files"}, {$group : { _id: "$files.v.fileType", total : { $sum : 1} }},{ $sort : { total : -1 }},{ $limit : 5 }]).pretty()

	pipeline5 := []bson.M{
		{"$project": bson.M{"files": bson.M{"$objectToArray": "$files"}}},
		{"$unwind": "$files"},
		{"$group": bson.M{"_id": "$files.v.fileType", "total": bson.M{"$sum": "$files.v.fileSize"}}},
		{"$sort": bson.M{"total": -1}},
		{"$limit": 10},
	}

	var result5 []common.FileTypeBySize

	err = uploadCollection.Pipe(pipeline5).All(&result5)
	if err != nil {
		err = log.EWarningf("Unable to get total file size from mongodb : %s", err)
	}

	stats.FileTypeBySize = result5

	userCollection := session.DB(b.config.Database).C(b.config.UserCollection)

	// NUMBER OF USERS
	userCount, err := userCollection.Find(nil).Count()
	if err != nil {
		err = log.EWarningf("Unable to get user count from mongodb : %s", err)
	}
	stats.Users = userCount

	return
}
