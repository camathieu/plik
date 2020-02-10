package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFile(t *testing.T) {
	file := NewFile()
	require.NotNil(t, file, "invalid file")
	require.NotZero(t, file.ID, "invalid file id")
}

func TestFileGenerateID(t *testing.T) {
	file := &File{}
	file.GenerateID()
	require.NotEqual(t, "", file.ID, "missing file id")
}

func TestFileSanitize(t *testing.T) {
	file := &File{}
	file.BackendDetails = "value"
	file.Sanitize()
	require.Zero(t, file.BackendDetails, "invalid backend details")
}
