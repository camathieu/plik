package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewInvite(t *testing.T) {
	issuer := &User{ID: "user"}
	invite, err := NewInvite(issuer, 30*24*time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, invite.ID)
	require.Equal(t, issuer.ID, *invite.Issuer)
	require.False(t, invite.HasExpired())
}

func TestNewInviteNoIssuer(t *testing.T) {
	invite, err := NewInvite(nil, 30*24*time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, invite.ID)
	require.Nil(t, invite.Issuer)
	require.False(t, invite.HasExpired())
}

func TestNewInviteNoTTL(t *testing.T) {
	issuer := &User{ID: "user"}
	invite, err := NewInvite(issuer, -1)
	require.NoError(t, err)
	require.NotEmpty(t, invite.ID)
	require.Equal(t, issuer.ID, *invite.Issuer)
	require.Nil(t, invite.ExpireAt)
	require.False(t, invite.HasExpired())
}

func TestInvite_HasExpired(t *testing.T) {
	issuer := &User{ID: "user"}
	invite, err := NewInvite(issuer, -1)
	require.NoError(t, err)
	require.False(t, invite.HasExpired())

	invite, err = NewInvite(issuer, 0)
	require.NoError(t, err)
	require.False(t, invite.HasExpired())

	invite, err = NewInvite(issuer, time.Hour)
	require.NoError(t, err)
	require.False(t, invite.HasExpired())

	deadline := time.Now().Add(-time.Hour)
	invite.ExpireAt = &deadline
	require.True(t, invite.HasExpired())
}
