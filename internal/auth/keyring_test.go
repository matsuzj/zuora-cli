package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockCredentialStore_SetAndGet(t *testing.T) {
	store := NewMockCredentialStore()
	require.NoError(t, store.Set("sandbox", "id-123", "secret-456"))

	id, secret, err := store.Get("sandbox")
	assert.NoError(t, err)
	assert.Equal(t, "id-123", id)
	assert.Equal(t, "secret-456", secret)
}

func TestMockCredentialStore_GetMissing(t *testing.T) {
	store := NewMockCredentialStore()

	_, _, err := store.Get("nonexistent")
	assert.Error(t, err)

	var authErr *AuthError
	assert.ErrorAs(t, err, &authErr)
}

func TestMockCredentialStore_Delete(t *testing.T) {
	store := NewMockCredentialStore()
	require.NoError(t, store.Set("sandbox", "id", "secret"))
	require.NoError(t, store.Delete("sandbox"))

	_, _, err := store.Get("sandbox")
	assert.Error(t, err)
}

func TestEnvVarStore(t *testing.T) {
	store := &envVarStore{clientID: "env-id", clientSecret: "env-secret"}

	id, secret, err := store.Get("any-env")
	assert.NoError(t, err)
	assert.Equal(t, "env-id", id)
	assert.Equal(t, "env-secret", secret)

	assert.Error(t, store.Set("any-env", "a", "b"))
	assert.Error(t, store.Delete("any-env"))
}

func TestNewCredentialStore_EnvVars(t *testing.T) {
	t.Setenv("ZR_CLIENT_ID", "env-id")
	t.Setenv("ZR_CLIENT_SECRET", "env-secret")

	store := NewCredentialStore()
	id, secret, err := store.Get("any")
	assert.NoError(t, err)
	assert.Equal(t, "env-id", id)
	assert.Equal(t, "env-secret", secret)
}
