package config

import "fmt"

// MockConfig is an in-memory Config implementation for testing.
type MockConfig struct {
	Cfg          configData
	Envs         map[string]*Environment
	Toks         map[string]*TokenEntry
	Dir          string
	SaveError    error
	SaveCallCount int
}

// NewMockConfig creates a MockConfig with defaults.
func NewMockConfig() *MockConfig {
	return &MockConfig{
		Cfg: configData{
			ActiveEnvironment: defaultActiveEnvironment,
			ZuoraVersion:      defaultZuoraVersion,
			DefaultOutput:     defaultOutput,
		},
		Envs: DefaultEnvironments(),
		Toks: make(map[string]*TokenEntry),
		Dir:  "/tmp/zr-test",
	}
}

func (m *MockConfig) ActiveEnvironment() string              { return m.Cfg.ActiveEnvironment }
func (m *MockConfig) ZuoraVersion() string                   { return m.Cfg.ZuoraVersion }
func (m *MockConfig) DefaultOutput() string                  { return m.Cfg.DefaultOutput }
func (m *MockConfig) ConfigDir() string                      { return m.Dir }
func (m *MockConfig) SetZuoraVersion(v string) error         { m.Cfg.ZuoraVersion = v; return nil }
func (m *MockConfig) Environments() map[string]*Environment  { return m.Envs }

func (m *MockConfig) SetActiveEnvironment(name string) error {
	if _, ok := m.Envs[name]; !ok {
		return fmt.Errorf("unknown environment: %s", name)
	}
	m.Cfg.ActiveEnvironment = name
	return nil
}

func (m *MockConfig) SetDefaultOutput(v string) error {
	switch v {
	case "table", "json":
		m.Cfg.DefaultOutput = v
		return nil
	default:
		return fmt.Errorf("invalid output format: %s", v)
	}
}

func (m *MockConfig) Environment(name string) (*Environment, error) {
	env, ok := m.Envs[name]
	if !ok {
		return nil, fmt.Errorf("unknown environment: %s", name)
	}
	return env, nil
}

func (m *MockConfig) AddEnvironment(name string, env *Environment) error {
	m.Envs[name] = env
	return nil
}

func (m *MockConfig) RemoveEnvironment(name string) error {
	delete(m.Envs, name)
	return nil
}

func (m *MockConfig) Token(envName string) (*TokenEntry, error) {
	t, ok := m.Toks[envName]
	if !ok {
		return nil, nil
	}
	return t, nil
}

func (m *MockConfig) SetToken(envName string, token *TokenEntry) error {
	m.Toks[envName] = token
	return nil
}

func (m *MockConfig) RemoveToken(envName string) error {
	delete(m.Toks, envName)
	return nil
}

func (m *MockConfig) Save() error {
	m.SaveCallCount++
	return m.SaveError
}
