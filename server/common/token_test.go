package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewToken(t *testing.T) {
	token := NewToken()
	require.NotNil(t, token, "invalid token")
}

func TestTokenCreate(t *testing.T) {
	token := NewToken()
	err := token.Create()
	require.NoError(t, err, "unable to create token")
	require.NotZero(t, token.Token, "missing token")
	require.NotZero(t, token.CreationDate, "missing creation date")
}
