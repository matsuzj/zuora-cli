package auth

import (
	"errors"
	"fmt"
	"os"

	"github.com/zalando/go-keyring"
)

const serviceName = "zuora-cli"

// CredentialStore provides access to client credentials.
type CredentialStore interface {
	// Set stores client credentials for an environment.
	Set(envName, clientID, clientSecret string) error
	// Get retrieves client credentials for an environment.
	Get(envName string) (clientID, clientSecret string, err error)
	// Delete removes client credentials for an environment.
	Delete(envName string) error
}

// EnvCredentials returns the ZR_CLIENT_ID/ZR_CLIENT_SECRET pair when BOTH
// variables are set — the single source of the both-or-nothing rule every
// consumer follows (store selection, auth status/logout displays, login
// prefill). A single variable alone is ignored everywhere; auth login falls
// back to flags or prompts instead of silently half-using the environment.
func EnvCredentials() (clientID, clientSecret string, ok bool) {
	id, secret := os.Getenv("ZR_CLIENT_ID"), os.Getenv("ZR_CLIENT_SECRET")
	if id != "" && secret != "" {
		return id, secret, true
	}
	return "", "", false
}

// NewCredentialStore creates a CredentialStore.
// Priority: env vars (both set) > OS keyring.
func NewCredentialStore() CredentialStore {
	if id, secret, ok := EnvCredentials(); ok {
		return &envVarStore{clientID: id, clientSecret: secret}
	}
	return &keyringStore{}
}

// KeyringStore returns a CredentialStore backed by the OS keychain.
func KeyringStore() CredentialStore {
	return &keyringStore{}
}

// keyringStore uses the OS keychain.
type keyringStore struct{}

func (s *keyringStore) Set(envName, clientID, clientSecret string) error {
	if err := keyring.Set(serviceName, envName+"/client_id", clientID); err != nil {
		return fmt.Errorf("storing client ID in keyring: %w", err)
	}
	if err := keyring.Set(serviceName, envName+"/client_secret", clientSecret); err != nil {
		return fmt.Errorf("storing client secret in keyring: %w", err)
	}
	return nil
}

func (s *keyringStore) Get(envName string) (string, string, error) {
	clientID, err := keyring.Get(serviceName, envName+"/client_id")
	if err != nil {
		hint := "Run 'zr auth login' to authenticate."
		// A keyring error other than "not found" usually means the OS keyring
		// is unavailable (e.g. headless Linux/CI without a secret service).
		if !errors.Is(err, keyring.ErrNotFound) {
			hint = "OS keyring unavailable. Set ZR_CLIENT_ID and ZR_CLIENT_SECRET environment variables instead."
		}
		return "", "", &AuthError{
			Message: fmt.Sprintf("no credentials found for environment %q", envName),
			Hint:    hint,
		}
	}
	clientSecret, err := keyring.Get(serviceName, envName+"/client_secret")
	if err != nil {
		return "", "", &AuthError{
			Message: fmt.Sprintf("client secret not found for environment %q", envName),
			Hint:    "Run 'zr auth login' to authenticate.",
		}
	}
	return clientID, clientSecret, nil
}

func (s *keyringStore) Delete(envName string) error {
	var firstErr error
	// Treat "not found" as success: deleting an already-absent secret is fine.
	if err := keyring.Delete(serviceName, envName+"/client_id"); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		firstErr = err
	}
	if err := keyring.Delete(serviceName, envName+"/client_secret"); err != nil && !errors.Is(err, keyring.ErrNotFound) && firstErr == nil {
		firstErr = err
	}
	return firstErr
}

// envVarStore reads credentials from environment variables.
type envVarStore struct {
	clientID     string
	clientSecret string
}

func (s *envVarStore) Get(_ string) (string, string, error) {
	return s.clientID, s.clientSecret, nil
}

func (s *envVarStore) Set(_ string, _, _ string) error {
	return fmt.Errorf("cannot store credentials when using environment variables (ZR_CLIENT_ID/ZR_CLIENT_SECRET)")
}

func (s *envVarStore) Delete(_ string) error {
	return fmt.Errorf("cannot delete credentials when using environment variables (ZR_CLIENT_ID/ZR_CLIENT_SECRET)")
}

// StaticCredentialStore is an in-memory CredentialStore. It is not just a
// test double: auth login uses it to validate freshly entered credentials
// against the OAuth endpoint BEFORE persisting them to the keyring.
type StaticCredentialStore struct {
	Creds map[string][2]string // envName -> [clientID, clientSecret]
}

func NewStaticCredentialStore() *StaticCredentialStore {
	return &StaticCredentialStore{Creds: make(map[string][2]string)}
}

// MockCredentialStore is the test-facing alias of StaticCredentialStore,
// kept so existing tests read naturally.
type MockCredentialStore = StaticCredentialStore

// NewMockCredentialStore is the test-facing alias of NewStaticCredentialStore.
func NewMockCredentialStore() *StaticCredentialStore { return NewStaticCredentialStore() }

func (m *StaticCredentialStore) Set(envName, clientID, clientSecret string) error {
	m.Creds[envName] = [2]string{clientID, clientSecret}
	return nil
}

func (m *StaticCredentialStore) Get(envName string) (string, string, error) {
	c, ok := m.Creds[envName]
	if !ok {
		return "", "", &AuthError{
			Message: fmt.Sprintf("no credentials found for environment %q", envName),
			Hint:    "Run 'zr auth login' to authenticate.",
		}
	}
	return c[0], c[1], nil
}

func (m *StaticCredentialStore) Delete(envName string) error {
	delete(m.Creds, envName)
	return nil
}
