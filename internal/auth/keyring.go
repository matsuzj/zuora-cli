package auth

import (
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

// NewCredentialStore creates a CredentialStore.
// Priority: env vars > OS keyring.
func NewCredentialStore() CredentialStore {
	if id, secret := os.Getenv("ZR_CLIENT_ID"), os.Getenv("ZR_CLIENT_SECRET"); id != "" && secret != "" {
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
		return "", "", &AuthError{
			Message: fmt.Sprintf("no credentials found for environment %q", envName),
			Hint:    "Run 'zr auth login' to authenticate.",
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
	if err := keyring.Delete(serviceName, envName+"/client_id"); err != nil {
		firstErr = err
	}
	if err := keyring.Delete(serviceName, envName+"/client_secret"); err != nil && firstErr == nil {
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

// MockCredentialStore is an in-memory credential store for testing.
type MockCredentialStore struct {
	Creds map[string][2]string // envName -> [clientID, clientSecret]
}

func NewMockCredentialStore() *MockCredentialStore {
	return &MockCredentialStore{Creds: make(map[string][2]string)}
}

func (m *MockCredentialStore) Set(envName, clientID, clientSecret string) error {
	m.Creds[envName] = [2]string{clientID, clientSecret}
	return nil
}

func (m *MockCredentialStore) Get(envName string) (string, string, error) {
	c, ok := m.Creds[envName]
	if !ok {
		return "", "", &AuthError{
			Message: fmt.Sprintf("no credentials found for environment %q", envName),
			Hint:    "Run 'zr auth login' to authenticate.",
		}
	}
	return c[0], c[1], nil
}

func (m *MockCredentialStore) Delete(envName string) error {
	delete(m.Creds, envName)
	return nil
}
