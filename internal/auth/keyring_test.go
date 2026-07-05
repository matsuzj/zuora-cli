package auth

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
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

// keyringStore against the real zalando/go-keyring API, backed by its
// in-memory mock — the store had 0% coverage before (hollow-coverage audit).
func TestKeyringStore_SetGetDelete(t *testing.T) {
	keyring.MockInit()
	s := &keyringStore{}

	require.NoError(t, s.Set("sandbox", "id-1", "secret-1"))
	id, secret, err := s.Get("sandbox")
	require.NoError(t, err)
	assert.Equal(t, "id-1", id)
	assert.Equal(t, "secret-1", secret)

	require.NoError(t, s.Delete("sandbox"))
	_, _, err = s.Get("sandbox")
	require.Error(t, err, "deleted credentials must not resolve")
}

func TestKeyringStore_GetMissing(t *testing.T) {
	keyring.MockInit()
	s := &keyringStore{}
	_, _, err := s.Get("never-set")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no credentials found")
}

// TestKeyringStore_SetError_WrapsStoringClientID pins the Set failure wrap:
// a keyring backend error surfaces with the "storing client ID" context and
// stays reachable via errors.Is.
func TestKeyringStore_SetError_WrapsStoringClientID(t *testing.T) {
	sentinel := errors.New("keyring backend down")
	keyring.MockInitWithError(sentinel)
	t.Cleanup(keyring.MockInit) // don't leak the failing provider into later tests

	err := (&keyringStore{}).Set("sandbox", "id", "secret")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "storing client ID in keyring")
	assert.ErrorIs(t, err, sentinel)
}

// TestKeyringStore_DeleteError_Surfaced pins that a non-NotFound delete
// failure is returned, not swallowed like ErrNotFound is.
func TestKeyringStore_DeleteError_Surfaced(t *testing.T) {
	sentinel := errors.New("keyring backend down")
	keyring.MockInitWithError(sentinel)
	t.Cleanup(keyring.MockInit)

	err := (&keyringStore{}).Delete("sandbox")
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel, "a real keyring failure must surface from Delete")
}

// TestKeyringStore_Delete_AttemptsSecondAfterFirstNotFound pins the other half
// of the Delete contract: the client_id delete returning ErrNotFound is
// tolerated AND the client_secret delete still runs (observed by the secret
// actually disappearing from the store).
func TestKeyringStore_Delete_AttemptsSecondAfterFirstNotFound(t *testing.T) {
	keyring.MockInit()
	// Only the client_secret entry exists; the client_id delete hits ErrNotFound.
	require.NoError(t, keyring.Set(serviceName, "sandbox/client_secret", "s3cr3t"))

	require.NoError(t, (&keyringStore{}).Delete("sandbox"), "ErrNotFound on the first key is not an error")

	_, err := keyring.Get(serviceName, "sandbox/client_secret")
	assert.ErrorIs(t, err, keyring.ErrNotFound,
		"the second delete must still run after the first returned ErrNotFound")
}

func TestNewCredentialStore_EnvVarsWinOverKeyring(t *testing.T) {
	// The env-var priority lives at store SELECTION time (NewCredentialStore),
	// not inside keyringStore.Get.
	keyring.MockInit()
	t.Setenv("ZR_CLIENT_ID", "env-id")
	t.Setenv("ZR_CLIENT_SECRET", "env-secret")

	s := NewCredentialStore()
	id, secret, err := s.Get("any-env")
	require.NoError(t, err)
	assert.Equal(t, "env-id", id, "both env vars set selects the env-var store")
	assert.Equal(t, "env-secret", secret)
}

func TestNewCredentialStore_PartialEnvFallsBackToKeyring(t *testing.T) {
	// Both-or-nothing: a single env var must NOT select the env store.
	keyring.MockInit()
	ks := &keyringStore{}
	require.NoError(t, ks.Set("sandbox", "keyring-id", "keyring-secret"))
	t.Setenv("ZR_CLIENT_ID", "env-id")
	t.Setenv("ZR_CLIENT_SECRET", "")

	s := NewCredentialStore()
	id, _, err := s.Get("sandbox")
	require.NoError(t, err)
	assert.Equal(t, "keyring-id", id, "one env var alone falls back to the keyring")
}

// EnvCredentials is the single source of the both-or-nothing rule.
func TestEnvCredentials_BothOrNothing(t *testing.T) {
	cases := []struct {
		name, id, secret string
		ok               bool
	}{
		{"both set", "id", "sec", true},
		{"only id", "id", "", false},
		{"only secret", "", "sec", false},
		{"neither", "", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv("ZR_CLIENT_ID", c.id)
			t.Setenv("ZR_CLIENT_SECRET", c.secret)
			id, secret, ok := EnvCredentials()
			assert.Equal(t, c.ok, ok)
			if c.ok {
				assert.Equal(t, c.id, id)
				assert.Equal(t, c.secret, secret)
			} else {
				assert.Empty(t, id)
				assert.Empty(t, secret)
			}
		})
	}
}
