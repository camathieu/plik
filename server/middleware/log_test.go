package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/root-gg/logger"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"github.com/stretchr/testify/require"
)

func TestLogInfo(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	log := context.GetLogger(ctx)
	log.SetMinLevel(logger.INFO)

	buffer := &bytes.Buffer{}
	log.SetOutput(buffer)

	req, err := http.NewRequest("GET", "url", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	req.RequestURI = "path"

	rr := httptest.NewRecorder()
	Log(ctx, common.DummyHandler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "invalid handler response status code")
	require.Contains(t, string(buffer.Bytes()), "GET path", "invalid log message")
}

func TestLogDebug(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	log := context.GetLogger(ctx)
	log.SetMinLevel(logger.DEBUG)

	buffer := &bytes.Buffer{}
	log.SetOutput(buffer)

	req, err := http.NewRequest("GET", "/version", bytes.NewBuffer([]byte("request body")))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	Log(ctx, common.DummyHandler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "invalid handler response status code")
	require.Contains(t, string(buffer.Bytes()), "GET /version HTTP/1.1", "invalid log message")
	require.Contains(t, string(buffer.Bytes()), "request body", "invalid log message")
}

func TestLogDebugNoBody(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	log := context.GetLogger(ctx)
	log.SetMinLevel(logger.DEBUG)

	buffer := &bytes.Buffer{}
	log.SetOutput(buffer)

	req, err := http.NewRequest("POST", "/file", bytes.NewBuffer([]byte("request body")))
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	Log(ctx, common.DummyHandler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "invalid handler response status code")
	require.Contains(t, string(buffer.Bytes()), "POST /file HTTP/1.1", "invalid log message")
	require.NotContains(t, string(buffer.Bytes()), "request body", "invalid log message")
}
