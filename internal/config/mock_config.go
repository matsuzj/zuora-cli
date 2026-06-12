package config

// MockConfig is an in-memory Config implementation for testing. Every
// behavior DELEGATES to the real fileConfig — so validation (Zuora version
// format, base-URL checks, unknown-environment errors) and locking match
// production exactly instead of drifting — and only persistence is replaced:
// Save is a spy that never touches the filesystem.
type MockConfig struct {
	fc *fileConfig

	// Envs is the live environments map shared with the delegate; tests may
	// mutate it directly (cfg.Envs["sandbox"] = ...) as before.
	Envs map[string]*Environment
	// Toks is the live token map shared with the delegate.
	Toks map[string]*TokenEntry
	// Dir is the value ConfigDir reports (no directory is ever created).
	Dir string

	SaveError     error
	SaveCallCount int
}

// NewMockConfig creates a MockConfig with defaults.
func NewMockConfig() *MockConfig {
	fc := &fileConfig{
		dir: "/tmp/zr-test",
		cfg: configData{
			ActiveEnvironment: defaultActiveEnvironment,
			ZuoraVersion:      defaultZuoraVersion,
			DefaultOutput:     defaultOutput,
		},
		envs: environmentsData{Environments: DefaultEnvironments()},
		toks: tokensData{Tokens: make(map[string]*TokenEntry)},
	}
	return &MockConfig{
		fc:   fc,
		Envs: fc.envs.Environments,
		Toks: fc.toks.Tokens,
		Dir:  fc.dir,
	}
}

// sync re-points the delegate at the exported maps before every delegated
// call, so a test that REPLACES a whole map (cfg.Envs = map[...]) — not just
// mutates it — is still honored, exactly like the old field-backed mock
// (review finding on this PR).
func (m *MockConfig) sync() {
	m.fc.envs.Environments = m.Envs
	m.fc.toks.Tokens = m.Toks
}

func (m *MockConfig) ActiveEnvironment() string { return m.fc.ActiveEnvironment() }
func (m *MockConfig) SetActiveEnvironment(name string) error {
	m.sync()
	return m.fc.SetActiveEnvironment(name)
}
func (m *MockConfig) ZuoraVersion() string            { return m.fc.ZuoraVersion() }
func (m *MockConfig) SetZuoraVersion(v string) error  { return m.fc.SetZuoraVersion(v) }
func (m *MockConfig) DefaultOutput() string           { return m.fc.DefaultOutput() }
func (m *MockConfig) SetDefaultOutput(v string) error { return m.fc.SetDefaultOutput(v) }
func (m *MockConfig) ConfigDir() string               { return m.Dir }

func (m *MockConfig) Environment(name string) (*Environment, error) {
	m.sync()
	return m.fc.Environment(name)
}
func (m *MockConfig) Environments() map[string]*Environment {
	m.sync()
	return m.fc.Environments()
}
func (m *MockConfig) AddEnvironment(name string, env *Environment) error {
	m.sync()
	return m.fc.AddEnvironment(name, env)
}
func (m *MockConfig) RemoveEnvironment(name string) error {
	m.sync()
	return m.fc.RemoveEnvironment(name)
}

func (m *MockConfig) Token(envName string) (*TokenEntry, error) {
	m.sync()
	return m.fc.Token(envName)
}
func (m *MockConfig) SetToken(envName string, token *TokenEntry) error {
	m.sync()
	return m.fc.SetToken(envName, token)
}
func (m *MockConfig) RemoveToken(envName string) error {
	m.sync()
	return m.fc.RemoveToken(envName)
}

// Save records the call and returns the injected error without writing
// anything to disk.
func (m *MockConfig) Save() error {
	m.SaveCallCount++
	return m.SaveError
}
