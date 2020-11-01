package common

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func getVersionRegex() string {
	return `^\d+\.\d+((\.\d+)?|(\-RC\d+))$`
}

func validateVersion(t *testing.T, version string, ok bool) {
	matched, err := regexp.Match(getVersionRegex(), []byte(version))
	require.NoError(t, err, "invalid version regex")

	if ok {
		require.True(t, matched, "invalid version regex match")
	} else {
		require.False(t, matched, "invalid version regex match")
	}
}

func TestValidateVersionRegex(t *testing.T) {
	validateVersion(t, "1.1", true)
	validateVersion(t, "1.1.1", true)
	validateVersion(t, "1.1-RC1", true)
	validateVersion(t, "1.1.1-RC1", false)
	validateVersion(t, "1.1-rc1", false)
}

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	require.NotZero(t, version, "missing version")
	validateVersion(t, version, true)
}

func TestGetBuildInfo(t *testing.T) {
	buildInfo := GetBuildInfo()
	require.NotNil(t, buildInfo, "missing build info")
	require.NotZero(t, buildInfo.Version, "missing build info version")
}

func TestGetBuildInfoString(t *testing.T) {
	buildInfo := GetBuildInfo()
	buildInfo.GitShortRevision = "foobar"
	buildInfo.IsRelease = true
	buildInfo.IsMint = true
	buildInfo.Date = time.Now().Unix()
	buildInfo.GoVersion = runtime.Version()
	v := buildInfo.String()
	require.NotZero(t, v, "empty build string")
	require.True(t, strings.Contains(v, "foobar"), "invalid version string")
	require.True(t, strings.Contains(v, "mint"), "not mint")
	require.True(t, strings.Contains(v, "release"), "not release")
	require.True(t, strings.Contains(v, runtime.Version()), "not go version")
}

func TestGetBuildInfoStringSanitize(t *testing.T) {
	buildInfo := GetBuildInfo()
	buildInfo.Sanitize()
	v := buildInfo.String()
	require.Equal(t, fmt.Sprintf("v%s", GetVersion()), v, "invalid build string")
}
