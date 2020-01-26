package common

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStripPrefixNoPrefix(t *testing.T) {
	req, err := http.NewRequest("GET", "/prefix", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	StripPrefix("", DummyHandler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "invalid handler response status code")
	require.Equal(t, "/prefix", req.URL.Path, "invalid request url")
}

func TestStripPrefixNoExactPrefix(t *testing.T) {
	req, err := http.NewRequest("GET", "/prefix", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	StripPrefix("/prefix", DummyHandler).ServeHTTP(rr, req)

	require.Equal(t, 301, rr.Code, "invalid handler response status code")
	require.Equal(t, "/prefix/", rr.Result().Header.Get("Location"), "invalid location header")
}

func TestStripPrefix(t *testing.T) {
	req, err := http.NewRequest("GET", "/prefix/path", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	StripPrefix("/prefix", DummyHandler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "invalid handler response status code")
	require.Equal(t, "/path", req.URL.Path, "invalid location header")
}

func TestStripPrefixNotFound(t *testing.T) {
	req, err := http.NewRequest("GET", "/invalid/path", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	StripPrefix("/prefix", DummyHandler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code, "invalid handler response status code")
}

func TestStripPrefixRootSlash(t *testing.T) {
	req, err := http.NewRequest("GET", "/prefix/path", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	rr := httptest.NewRecorder()
	StripPrefix("/prefix/", DummyHandler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "invalid handler response status code")
	require.Equal(t, "/path", req.URL.Path, "invalid location header")
}
