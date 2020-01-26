package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	"github.com/stretchr/testify/require"
)

func TestSourceIPInvalid(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	req, err := http.NewRequest("GET", "url", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")

	req.RemoteAddr = "invalid_ip_address"
	rr := httptest.NewRecorder()
	SourceIP(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestFail(t, rr, http.StatusInternalServerError, "Unable to parse source IP address")
}

func TestSourceIPInvalidFromHeader(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	context.GetConfig(ctx).SourceIPHeader = "IP"

	req, err := http.NewRequest("GET", "url", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")
	req.Header.Set("IP", "invalid_ip_address")

	rr := httptest.NewRecorder()
	SourceIP(ctx, common.DummyHandler).ServeHTTP(rr, req)

	context.TestFail(t, rr, http.StatusInternalServerError, "Unable to parse source IP address")
}

func TestSourceIP(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())

	req, err := http.NewRequest("GET", "url", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")
	req.RemoteAddr = "1.1.1.1:1111"

	rr := httptest.NewRecorder()
	SourceIP(ctx, common.DummyHandler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "invalid handler response status code")

	ip := context.GetSourceIP(ctx)
	require.Equal(t, "1.1.1.1", ip.String(), "invalid source ip from context")
}

func TestSourceIPFromHeader(t *testing.T) {
	ctx := newTestingContext(common.NewConfiguration())
	context.GetConfig(ctx).SourceIPHeader = "IP"

	req, err := http.NewRequest("GET", "url", &bytes.Buffer{})
	require.NoError(t, err, "unable to create new request")
	req.Header.Set("IP", "1.1.1.1")

	rr := httptest.NewRecorder()
	SourceIP(ctx, common.DummyHandler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "invalid handler response status code")

	ip := context.GetSourceIP(ctx)
	require.Equal(t, "1.1.1.1", ip.String(), "invalid source ip from context")
}
