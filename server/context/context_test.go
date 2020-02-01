package context

//
//import (
//	"bytes"
//	"encoding/json"
//	"net"
//	"net/http"
//	"net/http/httptest"
//	"testing"
//
//	"github.com/root-gg/juliet"
//	"github.com/root-gg/logger"
//
//	"github.com/root-gg/plik/server/common"
//	data_test "github.com/root-gg/plik/server/data/testing"
//	metadata_test "github.com/root-gg/plik/server/metadata/testing"
//	"github.com/stretchr/testify/require"
//)
//
//func TestConfig(t *testing.T) {
//	ctx := juliet.NewContext()
//	SetConfig(ctx, common.NewConfiguration())
//	require.NotNil(t, GetConfig(ctx), "invalid nil config")
//	ctx.Clear()
//	require.Nil(t, GetConfig(ctx), "invalid not nil config")
//}
//
//func TestLogger(t *testing.T) {
//	ctx := juliet.NewContext()
//	SetLogger(ctx, logger.NewLogger())
//	require.NotNil(t, GetLogger(ctx), "invalid nil logger")
//	ctx.Clear()
//	require.Nil(t, GetLogger(ctx), "invalid not nil logger")
//}
//
//func TestMetadataBackend(t *testing.T) {
//	ctx := juliet.NewContext()
//	SetMetadataBackend(ctx, metadata_test.NewBackend())
//	require.NotNil(t, GetMetadataBackend(ctx), "invalid nil metadata backend")
//	ctx.Clear()
//	require.Nil(t, GetMetadataBackend(ctx), "invalid not nil metadata backend")
//}
//
//func TestDataBackend(t *testing.T) {
//	ctx := juliet.NewContext()
//	SetDataBackend(ctx, data_test.NewBackend())
//	require.NotNil(t, GetDataBackend(ctx), "invalid nil data backend")
//	ctx.Clear()
//	require.Nil(t, GetDataBackend(ctx), "invalid not nil data backend")
//
//}
//
//func TestStreamBackend(t *testing.T) {
//	ctx := juliet.NewContext()
//	SetStreamBackend(ctx, data_test.NewBackend())
//	require.NotNil(t, GetStreamBackend(ctx), "invalid nil stream backend")
//	ctx.Clear()
//	require.Nil(t, GetStreamBackend(ctx), "invalid not nil stream backend")
//}
//
//func TestSourceIP(t *testing.T) {
//	ctx := juliet.NewContext()
//	SetSourceIP(ctx, net.ParseIP("1.1.1.1"))
//	require.NotNil(t, GetSourceIP(ctx), "invalid nil source ip")
//	ctx.Clear()
//	require.Nil(t, GetSourceIP(ctx), "invalid not nil source ip")
//}
//
//func TestIsWhitelistedAlreadyInContext(t *testing.T) {
//	ctx := juliet.NewContext()
//
//	SetWhitelisted(ctx, false)
//	require.False(t, IsWhitelisted(ctx), "invalid whitelisted status")
//
//	SetWhitelisted(ctx, true)
//	require.True(t, IsWhitelisted(ctx), "invalid whitelisted status")
//}
//
//func TestIsWhitelistedNoWhitelist(t *testing.T) {
//	config := common.NewConfiguration()
//	err := config.Initialize()
//	require.NoError(t, err, "unable to initialize config")
//
//	ctx := juliet.NewContext()
//	SetConfig(ctx, config)
//	SetSourceIP(ctx, net.ParseIP("1.1.1.1"))
//
//	require.True(t, IsWhitelisted(ctx), "invalid whitelisted status")
//}
//
//func TestIsWhitelistedNoIp(t *testing.T) {
//	config := common.NewConfiguration()
//	config.UploadWhitelist = append(config.UploadWhitelist, "1.1.1.1")
//	err := config.Initialize()
//	require.NoError(t, err, "unable to initialize config")
//
//	ctx := juliet.NewContext()
//	SetConfig(ctx, config)
//
//	require.False(t, IsWhitelisted(ctx), "invalid whitelisted status")
//}
//
//func TestIsWhitelisted(t *testing.T) {
//	config := common.NewConfiguration()
//	config.UploadWhitelist = append(config.UploadWhitelist, "1.1.1.1")
//	err := config.Initialize()
//	require.NoError(t, err, "unable to initialize config")
//
//	ctx := juliet.NewContext()
//	SetConfig(ctx, config)
//	SetSourceIP(ctx, net.ParseIP("1.1.1.1"))
//
//	require.True(t, IsWhitelisted(ctx), "invalid whitelisted status")
//}
//
//func TestGetUser(t *testing.T) {
//	ctx := juliet.NewContext()
//	SetUser(ctx, common.NewUser())
//	require.NotNil(t, GetUser(ctx), "invalid nil user")
//	ctx.Clear()
//	require.Nil(t, GetUser(ctx), "invalid not nil user")
//}
//
//func TestGetToken(t *testing.T) {
//	ctx := juliet.NewContext()
//	SetToken(ctx, common.NewToken())
//	require.NotNil(t, GetToken(ctx), "invalid nil token")
//	ctx.Clear()
//	require.Nil(t, GetToken(ctx), "invalid not nil token")
//}
//
//func TestGetFile(t *testing.T) {
//	ctx := juliet.NewContext()
//	SetFile(ctx, common.NewFile())
//	require.NotNil(t, GetFile(ctx), "invalid nil file")
//	ctx.Clear()
//	require.Nil(t, GetFile(ctx), "invalid not nil file")
//}
//
//func TestGetUpload(t *testing.T) {
//	ctx := juliet.NewContext()
//	SetUpload(ctx, common.NewUpload())
//	require.NotNil(t, GetUpload(ctx), "invalid nil upload")
//	ctx.Clear()
//	require.Nil(t, GetUpload(ctx), "invalid not nil upload")
//}
//
//func TestIsRedirectOnFailure(t *testing.T) {
//	ctx := juliet.NewContext()
//	SetRedirectOnFailure(ctx, true)
//	require.True(t, IsRedirectOnFailure(ctx), "invalid redirect status")
//	ctx.Clear()
//	require.False(t, IsRedirectOnFailure(ctx), "invalid redirect status")
//}
//
//func TestFailNoRedirect(t *testing.T) {
//	ctx := juliet.NewContext()
//
//	req, err := http.NewRequest("GET", "/path", bytes.NewBuffer([]byte{}))
//	require.NoError(t, err, "unable to create new request")
//
//	rr := httptest.NewRecorder()
//	Fail(ctx, req, rr, "error", http.StatusInternalServerError)
//	TestFail(t, rr, http.StatusInternalServerError, "error")
//}
//
//func TestFailWebRedirect(t *testing.T) {
//	ctx := juliet.NewContext()
//	SetConfig(ctx, common.NewConfiguration())
//
//	SetRedirectOnFailure(ctx, true)
//	GetConfig(ctx).Path = "/root"
//
//	req, err := http.NewRequest("GET", "/path", bytes.NewBuffer([]byte{}))
//	require.NoError(t, err, "unable to create new request")
//
//	req.RequestURI = "/path"
//
//	rr := httptest.NewRecorder()
//	Fail(ctx, req, rr, "error", http.StatusInternalServerError)
//
//	require.Equal(t, http.StatusMovedPermanently, rr.Code, "invalid http response status code")
//	require.Contains(t, rr.Result().Header.Get("Location"), "/root", "invalid redirect root")
//	require.Contains(t, rr.Result().Header.Get("Location"), "err=error", "invalid redirect message")
//	require.Contains(t, rr.Result().Header.Get("Location"), "errcode=500", "invalid redirect code")
//	require.Contains(t, rr.Result().Header.Get("Location"), "uri=/path", "invalid redirect path")
//}
//
//func TestFailCliNoRedirect(t *testing.T) {
//	ctx := juliet.NewContext()
//	SetRedirectOnFailure(ctx, true)
//
//	req, err := http.NewRequest("GET", "/path", bytes.NewBuffer([]byte{}))
//	require.NoError(t, err, "unable to create new request")
//
//	req.Header.Set("User-Agent", "wget")
//
//	rr := httptest.NewRecorder()
//	Fail(ctx, req, rr, "error", http.StatusInternalServerError)
//	TestFail(t, rr, http.StatusInternalServerError, "error")
//}
//
//func TestTestFail(t *testing.T) {
//	result := common.NewResult("error", nil)
//
//	bytes, err := json.Marshal(result)
//	require.NoError(t, err, "unable to marshal result")
//
//	rr := httptest.NewRecorder()
//	rr.WriteHeader(http.StatusInternalServerError)
//	_, err = rr.Write(bytes)
//
//	require.NoError(t, err, "unable to write response")
//	TestFail(t, rr, http.StatusInternalServerError, "error")
//}
//
//func TestNewTestingContext(t *testing.T) {
//	ctx := juliet.NewContext()
//	require.NotNil(t, ctx, "invalid nil context")
//}
