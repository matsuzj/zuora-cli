package auth

import (
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
