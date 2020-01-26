package mongo

import (
	"github.com/root-gg/plik/server/common"
)

// GetUsers get all users
func (b *Backend) GetUsers() (ids []string, err error) {
	panic("Not Yet Implemented")
}

// GetServerStatistics return server statistics
func (b *Backend) GetServerStatistics() (stats *common.ServerStats, err error) {
	panic("Not Yet Implemented")
}

// GetUploadsToRemove return upload ids that needs to be removed from the server
func (b *Backend) GetUploadsToRemove() (ids []string, err error) {
	panic("Not Yet Implemented")
}

//// GetUserUploads implementation from MongoDB Metadata Backend
//func (b *Backend) GetUserUploads(ctx *juliet.Context, user *common.User, token *common.Token) (ids []string, err error) {
//	log := context.GetLogger(ctx)
//
//	if user == nil {
//		err = log.EWarning("Unable to get user uploads : Missing user")
//		return
//	}
//
//	session := b.session.Copy()
//	defer session.Close()
//	collection := session.DB(b.config.Database).C(b.config.Collection)
//
//	b := bson.M{"user": user.ID}
//	if token != nil {
//		b["token"] = token.Token
//	}
//
//	var uploads []*common.Upload
//	err = collection.Find(b).Select(bson.M{"id": 1}).Sort("-uploadDate").All(&uploads)
//	if err != nil {
//		err = log.EWarningf("Unable to get user uploads : %s", err)
//		return
//	}
//
//	// Get all ids
//	for _, upload := range uploads {
//		ids = append(ids, upload.ID)
//	}
//
//	return
//}
//
//// GetUploadsToRemove implementation from MongoDB Metadata Backend
//func (b *Backend) GetUploadsToRemove(ctx *juliet.Context) (ids []string, err error) {
//	log := context.GetLogger(ctx)
//
//	session := b.session.Copy()
//	defer session.Close()
//	collection := session.DB(b.config.Database).C(b.config.Collection)
//
//	// Look for expired uploads
//	var uploads []*common.Upload
//	b := bson.M{"$where": "this.ttl > 0 && " + strconv.Itoa(int(time.Now().Unix())) + " > this.uploadDate + this.ttl"}
//
//	err = collection.Find(b).Select(bson.M{"id": 1}).All(&uploads)
//	if err != nil {
//		err = log.EWarningf("Unable to get uploads to remove : %s", err)
//		return
//	}
//
//	// Get all ids
//	for _, upload := range uploads {
//		ids = append(ids, upload.ID)
//	}
//
//	return
//}
//
//// GetUserStatistics implementation for MongoDB Metadata Backend
//func (b *Backend) GetUserStatistics(ctx *juliet.Context, user *common.User, token *common.Token) (stats *common.UserStats, err error) {
//	log := context.GetLogger(ctx)
//
//	if user == nil {
//		err = log.EWarning("Unable to get user uploads : Missing user")
//		return
//	}
//
//	session := b.session.Copy()
//	defer session.Close()
//	collection := session.DB(b.config.Database).C(b.config.Collection)
//
//	match := bson.M{"user": user.ID}
//	if token != nil {
//		match["token"] = token.Token
//	}
//
//	// db.plik_meta.aggregate([{$match: {user:"xxx", token:"xxx"}}, {$project: {"files": {$objectToArray: "$files"}}}, {$unwind: "$files"}, {$group: { _id: null, count: {$sum: 1}, total: {$sum: "$files.v.fileSize"}, uploads: {$addToSet: "$_id"}}}, {$project: {count: "$count", total: "$total", size: { $size: "$uploads"}}}]).pretty()
//	pipeline := []bson.M{
//		{"$match": match},
//		{"$project": bson.M{"file_count": bson.M{"$size": bson.M{"$objectToArray": "$files"}}}},
//		{"$unwind": "$files"},
//		{"$group": bson.M{"_id": nil, "files": bson.M{"$sum": 1}, "totalSize": bson.M{"$sum": "$files.v.fileSize"}, "uploads": bson.M{"$addToSet": "$_id"}}},
//		{"$project": bson.M{"Files": "$files", "TotalSize": "$totalSize", "Uploads": bson.M{"$size": "uploads"}}},
//	}
//
//	stats = new(common.UserStats)
//	err = collection.Pipe(pipeline).One(&stats)
//	if err != nil {
//		err = log.EWarningf("Unable to get file count from mongodb : %s", err)
//	}
//
//	return
//}
//
//// GetUsers implementation for MongoDB Metadata Backend
//func (b *Backend) GetUsers(ctx *juliet.Context) (ids []string, err error) {
//	log := context.GetLogger(ctx)
//
//	session := b.session.Copy()
//	defer session.Close()
//	collection := session.DB(b.config.Database).C(b.config.UserCollection)
//
//	var results []struct {
//		ID string `bson:"id"`
//	}
//	err = collection.Find(nil).Select(bson.M{"id": 1}).Sort("id").All(&results)
//	if err != nil {
//		err = log.EWarningf("Unable to get users from mongodb : %s", err)
//	}
//
//	for _, result := range results {
//		ids = append(ids, result.ID)
//	}
//
//	return
//}
//
//// GetServerStatistics implementation for MongoDB Metadata Backend
//func (b *Backend) GetServerStatistics(ctx *juliet.Context) (stats *common.ServerStats, err error) {
//	log := context.GetLogger(ctx)
//	config := context.GetConfig(ctx)
//
//	stats = new(common.ServerStats)
//	session := b.session.Copy()
//	defer session.Close()
//
//	uploadCollection := session.DB(b.config.Database).C(b.config.Collection)
//
//	// NUMBER OF UPLOADS
//	uploadCount, err := uploadCollection.Find(nil).Count()
//	if err != nil {
//		err = log.EWarningf("Unable to get upload count from mongodb : %s", err)
//	}
//	stats.Uploads = uploadCount
//
//	// NUMBER OF ANONYMOUS UPLOADS
//	anonymousUploadCount, err := uploadCollection.Find(bson.M{"user": ""}).Count()
//	if err != nil {
//		err = log.EWarningf("Unable to get anonymous upload count from mongodb : %s", err)
//	}
//	stats.AnonymousUploads = anonymousUploadCount
//
//	// NUMBER OF FILES
//	//db.plik_meta.aggregate([{$project: {"file_count": { $size: { $objectToArray : "$files" } }}}, { $group: { _id : null, total : { $sum : "$file_count" }}}]).pretty()
//
//	pipeline1 := []bson.M{
//		{"$project": bson.M{"file_count": bson.M{"$size": bson.M{"$objectToArray": "$files"}}}},
//		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": "$file_count"}}},
//	}
//
//	var result1 struct {
//		Total int `bson:"total"`
//	}
//
//	err = uploadCollection.Pipe(pipeline1).One(&result1)
//	if err != nil {
//		err = log.EWarningf("Unable to get file count from mongodb : %s", err)
//	}
//
//	stats.Files = result1.Total
//
//	// TOTAL SIZE OF ALL FILES
//	// db.plik_meta.aggregate([{$project: {"files": { $objectToArray : "$files" } }}, {$unwind: "$files"}, {$group : { _id: null, total : { $sum : "$files.v.fileSize"} }}]).pretty()
//
//	pipeline2 := []bson.M{
//		{"$project": bson.M{"files": bson.M{"$objectToArray": "$files"}}},
//		{"$unwind": "$files"},
//		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": "$files.v.fileSize"}}},
//	}
//
//	var result2 struct {
//		Total int64 `bson:"total"`
//	}
//
//	err = uploadCollection.Pipe(pipeline2).One(&result2)
//	if err != nil {
//		err = log.EWarningf("Unable to get total file size from mongodb : %s", err)
//	}
//
//	stats.TotalSize = result2.Total
//
//	if !config.NoAnonymousUploads {
//
//		// TOTAL SIZE OF ALL ANONYMOUS UPLOAD FILES
//		// db.plik_meta.aggregate([{$match: {user:""}},{$project: {"files": { $objectToArray : "$files" } }}, {$unwind: "$files"}, {$group : { _id: null, total : { $sum : "$files.v.fileSize"} }}]).pretty()
//
//		pipeline3 := []bson.M{
//			{"$match": bson.M{"user": ""}},
//			{"$project": bson.M{"files": bson.M{"$objectToArray": "$files"}}},
//			{"$unwind": "$files"},
//			{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": "$files.v.fileSize"}}},
//		}
//
//		var result3 struct {
//			Total int64 `bson:"total"`
//		}
//
//		err = uploadCollection.Pipe(pipeline3).One(&result3)
//		if err != nil {
//			err = log.EWarningf("Unable to get total file size from mongodb : %s", err)
//		}
//
//		stats.AnonymousSize = result3.Total
//	}
//
//	// TOTAL FILE SIZE BY FILE TYPE
//	// db.plik_meta.aggregate([{$project: {"files": { $objectToArray : "$files" } }}, {$unwind: "$files"}, {$group : { _id: "$files.v.fileType", total : { $sum : "$files.v.fileSize"} }},{ $sort : { total : -1 }},{ $limit : 5 }]).pretty()
//
//	pipeline4 := []bson.M{
//		{"$project": bson.M{"files": bson.M{"$objectToArray": "$files"}}},
//		{"$unwind": "$files"},
//		{"$group": bson.M{"_id": "$files.v.fileType", "total": bson.M{"$sum": 1}}},
//		{"$sort": bson.M{"total": -1}},
//		{"$limit": 10},
//	}
//
//	var result4 []common.FileTypeByCount
//
//	err = uploadCollection.Pipe(pipeline4).All(&result4)
//	if err != nil {
//		err = log.EWarningf("Unable to get total file size from mongodb : %s", err)
//	}
//
//	stats.FileTypeByCount = result4
//
//	// TOTAL FILE SIZE BY FILE TYPE
//	// db.plik_meta.aggregate([{$project: {"files": { $objectToArray : "$files" } }}, {$unwind: "$files"}, {$group : { _id: "$files.v.fileType", total : { $sum : 1} }},{ $sort : { total : -1 }},{ $limit : 5 }]).pretty()
//
//	pipeline5 := []bson.M{
//		{"$project": bson.M{"files": bson.M{"$objectToArray": "$files"}}},
//		{"$unwind": "$files"},
//		{"$group": bson.M{"_id": "$files.v.fileType", "total": bson.M{"$sum": "$files.v.fileSize"}}},
//		{"$sort": bson.M{"total": -1}},
//		{"$limit": 10},
//	}
//
//	var result5 []common.FileTypeBySize
//
//	err = uploadCollection.Pipe(pipeline5).All(&result5)
//	if err != nil {
//		err = log.EWarningf("Unable to get total file size from mongodb : %s", err)
//	}
//
//	stats.FileTypeBySize = result5
//
//	userCollection := session.DB(b.config.Database).C(b.config.UserCollection)
//
//	// NUMBER OF USERS
//	userCount, err := userCollection.Find(nil).Count()
//	if err != nil {
//		err = log.EWarningf("Unable to get user count from mongodb : %s", err)
//	}
//	stats.Users = userCount
//
//	return
//}
