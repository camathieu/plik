package handlers

//
//func addTestUser(ctx *context.Context, user *common.User) (err error) {
//	metadataBackend := ctx.GetMetadataBackend()
//	return metadataBackend.CreateUser(user)
//}
//
//func addTestUserAdmin(ctx *context.Context) (user *common.User, err error) {
//	user = common.NewUser()
//	user.ID = "admin"
//	user.Email = "admin@root.gg"
//	user.Login = "admin"
//	ctx.SetUser(user)
//	ctx.SetAdmin(true)
//	return user, addTestUser(ctx, user)
//}
//
//func TestGetUsers(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	_, err := addTestUserAdmin(ctx)
//	require.NoError(t, err, "unable to add user admin")
//
//	user1 := common.NewUser()
//	user1.ID = "user1"
//	user1.Email = "user1@root.gg"
//	user1.Login = "user1"
//
//	user2 := common.NewUser()
//	user2.ID = "user2"
//	user2.Email = "user2@root.gg"
//	user2.Login = "user2"
//
//	err = addTestUser(ctx, user1)
//	require.NoError(t, err, "unable to add user1")
//
//	err = addTestUser(ctx, user2)
//	require.NoError(t, err, "unable to add user2")
//
//	req, err := http.NewRequest("GET", "/admin/users", bytes.NewBuffer([]byte{}))
//	require.NoError(t, err, "unable to create new request")
//
//	rr := ctx.NewRecorder(req)
//	GetUsers(ctx, rr, req)
//
//	context.TestOK(t, rr)
//
//	respBody, err := ioutil.ReadAll(rr.Body)
//	require.NoError(t, err, "unable to read response body")
//
//	var users []*common.User
//	err = json.Unmarshal(respBody, &users)
//	require.NoError(t, err, "unable to unmarshal response body")
//
//	require.Equal(t, 3, len(users), "invalid user count")
//}
//
//func TestGetUsersNoUser(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	req, err := http.NewRequest("GET", "/admin/users", bytes.NewBuffer([]byte{}))
//	require.NoError(t, err, "unable to create new request")
//
//	rr := ctx.NewRecorder(req)
//	GetUsers(ctx, rr, req)
//
//	context.TestForbidden(t, rr, "you need administrator privileges")
//}
//
//func TestGetUsersNotAdmin(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	_, err := addTestUserAdmin(ctx)
//	require.NoError(t, err, "unable to add admin")
//	ctx.SetAdmin(false)
//
//	req, err := http.NewRequest("GET", "/admin/users", bytes.NewBuffer([]byte{}))
//	require.NoError(t, err, "unable to create new request")
//
//	rr := ctx.NewRecorder(req)
//	GetUsers(ctx, rr, req)
//
//	context.TestForbidden(t, rr, "you need administrator privileges")
//}
//
//func TestGetServerStatistics(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	_, err := addTestUserAdmin(ctx)
//	require.NoError(t, err, "unable to add user admin")
//
//	type pair struct {
//		typ   string
//		size  int64
//		count int
//	}
//
//	plan := []pair{
//		{"type1", 1, 1},
//		{"type2", 1000, 5},
//		{"type3", 1000 * 1000, 10},
//		{"type4", 1000 * 1000 * 1000, 15},
//	}
//
//	for _, item := range plan {
//		for i := 0; i < item.count; i++ {
//			upload := &common.Upload{}
//			upload.Create()
//			file := upload.NewFile()
//			file.Type = item.typ
//			file.CurrentSize = item.size
//
//			err := ctx.GetMetadataBackend().CreateUpload(upload)
//			require.NoError(t, err, "create error")
//		}
//	}
//
//	req, err := http.NewRequest("GET", "/admin/stats", bytes.NewBuffer([]byte{}))
//	require.NoError(t, err, "unable to create new request")
//
//	rr := ctx.NewRecorder(req)
//	GetServerStatistics(ctx, rr, req)
//
//	context.TestOK(t, rr)
//
//	respBody, err := ioutil.ReadAll(rr.Body)
//	require.NoError(t, err, "unable to read response body")
//
//	var stats *common.ServerStats
//	err = json.Unmarshal(respBody, &stats)
//	require.NoError(t, err, "unable to unmarshal response body")
//
//	require.NotNil(t, stats, "invalid server statistics")
//	require.Equal(t, 31, stats.Uploads, "invalid upload count")
//	require.Equal(t, 31, stats.Files, "invalid files count")
//	require.Equal(t, int64(15010005001), stats.TotalSize, "invalid total file size")
//	require.Equal(t, 31, stats.AnonymousUploads, "invalid anonymous upload count")
//	require.Equal(t, int64(15010005001), stats.AnonymousSize, "invalid anonymous total file size")
//	require.Equal(t, 10, len(stats.FileTypeByCount), "invalid file type by count length")
//	require.Equal(t, "type4", stats.FileTypeByCount[0].Type, "invalid file type by count type")
//	require.Equal(t, 10, len(stats.FileTypeBySize), "invalid file type by size length")
//	require.Equal(t, "type4", stats.FileTypeBySize[0].Type, "invalid file type by size type")
//
//}
//
//func TestGetServerStatisticsNoUser(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	req, err := http.NewRequest("GET", "/admin/users", bytes.NewBuffer([]byte{}))
//	require.NoError(t, err, "unable to create new request")
//	ctx.SetUser(nil)
//
//	rr := ctx.NewRecorder(req)
//	GetServerStatistics(ctx, rr, req)
//
//	context.TestForbidden(t, rr, "you need administrator privileges")
//}
//
//func TestGetServerStatisticsNotAdmin(t *testing.T) {
//	ctx := newTestingContext(common.NewConfiguration())
//
//	_, err := addTestUserAdmin(ctx)
//	require.NoError(t, err, "unable to add admin")
//	ctx.SetAdmin(false)
//
//	req, err := http.NewRequest("GET", "/admin/users", bytes.NewBuffer([]byte{}))
//	require.NoError(t, err, "unable to create new request")
//
//	rr := ctx.NewRecorder(req)
//	GetServerStatistics(ctx, rr, req)
//
//	context.TestForbidden(t, rr, "you need administrator privileges")
//}
