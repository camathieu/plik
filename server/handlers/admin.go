package handlers

// GetUsers return users information ( name / email / tokens / ... )
//func GetUsers(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
//	log := ctx.GetLogger()
//
//	if !ctx.IsAdmin() {
//		ctx.Forbidden("you need administrator privileges")
//		return
//	}
//
//	ids, err := ctx.GetMetadataBackend().GetUsers()
//	if err != nil {
//		ctx.InternalServerError("unable to get user IDs", err)
//		return
//	}
//
//	// Get size from URL query parameter
//	size := 100
//	sizeStr := req.URL.Query().Get("size")
//	if sizeStr != "" {
//		size, err = strconv.Atoi(sizeStr)
//		if err != nil || size <= 0 || size > 1000 {
//			ctx.InvalidParameter("size. must be positive integer up to 1000")
//			return
//		}
//	}
//
//	// Get offset from URL query parameter
//	offset := 0
//	offsetStr := req.URL.Query().Get("offset")
//	if offsetStr != "" {
//		offset, err = strconv.Atoi(offsetStr)
//		if err != nil || offset < 0 {
//			ctx.InvalidParameter("offset. must be positive integer")
//			return
//		}
//	}
//
//	// Adjust offset
//	if offset > len(ids) {
//		offset = len(ids)
//	}
//
//	// Adjust size
//	if offset+size > len(ids) {
//		size = len(ids) - offset
//	}
//
//	var users []*common.User
//	for _, id := range ids[offset : offset+size] {
//		user, err := ctx.GetMetadataBackend().GetUser(id)
//		if err != nil {
//			log.Warningf("Unable to get user %s : %s", id, err)
//			continue
//		}
//
//		// Remove tokens
//		user.Tokens = nil
//
//		users = append(users, user)
//	}
//
//	// Print users in the json response.
//	var json []byte
//	if json, err = utils.ToJson(users); err != nil {
//		panic(fmt.Errorf("unable to serialize json response : %s", err))
//	}
//
//	_, _ = resp.Write(json)
//}
//
//// GetServerStatistics return the server statistics
//func GetServerStatistics(ctx *context.Context, resp http.ResponseWriter, req *http.Request) {
//	if !ctx.IsAdmin() {
//		ctx.Forbidden("you need administrator privileges")
//		return
//	}
//
//	// Get server statistics
//	stats, err := ctx.GetMetadataBackend().GetServerStatistics()
//	if err != nil {
//		ctx.InternalServerError("unable to get server statistics", err)
//		return
//	}
//
//	// Print stats in the json response.
//	var json []byte
//	if json, err = utils.ToJson(stats); err != nil {
//		panic(fmt.Errorf("unable to serialize json response : %s", err))
//	}
//
//	_, _ = resp.Write(json)
//}
