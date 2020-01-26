package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewResult(t *testing.T) {
	result := NewResult("message", "value")
	require.NotNil(t, result, "invalid result")
}

func TestResultToJSON(t *testing.T) {
	result := NewResult("foo", "bar")

	str := result.ToJSONString()

	require.Contains(t, str, "foo", "invalid result json string")
	require.Contains(t, str, "bar", "missing result json string")
}
