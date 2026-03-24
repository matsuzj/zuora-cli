package config

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// configData represents config.yml.
type configData struct {
	ActiveEnvironment string `yaml:"active_environment"`
	ZuoraVersion      string `yaml:"zuora_version"`
	DefaultOutput     string `yaml:"default_output"`
}

// environmentsData represents environments.yml.
type environmentsData struct {
	Environments map[string]*Environment `yaml:"environments"`
}

// tokensData represents tokens.yml.
type tokensData struct {
	Tokens map[string]*TokenEntry `yaml:"tokens"`
}

// fileConfig is the file-based implementation of Config.
type fileConfig struct {
	dir  string
	mu   sync.Mutex
	cfg  configData
	envs environmentsData
	toks tokensData
}

// Load reads configuration from the given directory.
// If the directory or files don't exist, defaults are used.
func Load(dir string) (Config, error) {
	fc := &fileConfig{dir: dir}
	if err := fc.load(); err != nil {
		return nil, err
	}
	return fc, nil
}

// LoadDefault reads configuration from the default config directory.
func LoadDefault() (Config, error) {
	return Load(configDir())
}

func (fc *fileConfig) load() error {
	fc.cfg = configData{
		ActiveEnvironment: defaultActiveEnvironment,
		ZuoraVersion:      defaultZuoraVersion,
		DefaultOutput:     defaultOutput,
	}
	fc.envs = environmentsData{
		Environments: DefaultEnvironments(),
	}
	fc.toks = tokensData{
		Tokens: make(map[string]*TokenEntry),
	}

	if err := readYAML(configFilePath(fc.dir), &fc.cfg); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading config.yml: %w", err)
	}
	if err := readYAML(environmentsFilePath(fc.dir), &fc.envs); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading environments.yml: %w", err)
	}
	if err := readYAML(tokensFilePath(fc.dir), &fc.toks); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading tokens.yml: %w", err)
	}

	if fc.envs.Environments == nil {
		fc.envs.Environments = DefaultEnvironments()
	}
	if fc.toks.Tokens == nil {
		fc.toks.Tokens = make(map[string]*TokenEntry)
	}
	return nil
}

func (fc *fileConfig) ActiveEnvironment() string { return fc.cfg.ActiveEnvironment }
func (fc *fileConfig) ZuoraVersion() string      { return fc.cfg.ZuoraVersion }
func (fc *fileConfig) DefaultOutput() string      { return fc.cfg.DefaultOutput }
func (fc *fileConfig) ConfigDir() string          { return fc.dir }

func (fc *fileConfig) SetActiveEnvironment(name string) error {
	if _, ok := fc.envs.Environments[name]; !ok {
		return fmt.Errorf("unknown environment: %s", name)
	}
	fc.cfg.ActiveEnvironment = name
	return nil
}

func (fc *fileConfig) SetZuoraVersion(v string) error {
	fc.cfg.ZuoraVersion = v
	return nil
}

func (fc *fileConfig) SetDefaultOutput(v string) error {
	switch v {
	case "table", "json":
		fc.cfg.DefaultOutput = v
		return nil
	default:
		return fmt.Errorf("invalid output format: %s (must be table or json)", v)
	}
}

func (fc *fileConfig) Environment(name string) (*Environment, error) {
	env, ok := fc.envs.Environments[name]
	if !ok {
		return nil, fmt.Errorf("unknown environment: %s", name)
	}
	return env, nil
}

func (fc *fileConfig) Environments() map[string]*Environment {
	return fc.envs.Environments
}

func (fc *fileConfig) AddEnvironment(name string, env *Environment) error {
	fc.envs.Environments[name] = env
	return nil
}

func (fc *fileConfig) RemoveEnvironment(name string) error {
	if _, ok := fc.envs.Environments[name]; !ok {
		return fmt.Errorf("unknown environment: %s", name)
	}
	delete(fc.envs.Environments, name)
	return nil
}

func (fc *fileConfig) Token(envName string) (*TokenEntry, error) {
	t, ok := fc.toks.Tokens[envName]
	if !ok {
		return nil, nil
	}
	return t, nil
}

func (fc *fileConfig) SetToken(envName string, token *TokenEntry) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.toks.Tokens[envName] = token
	return nil
}

func (fc *fileConfig) RemoveToken(envName string) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	delete(fc.toks.Tokens, envName)
	return nil
}

func (fc *fileConfig) Save() error {
	if err := os.MkdirAll(fc.dir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	if err := writeYAML(configFilePath(fc.dir), &fc.cfg); err != nil {
		return fmt.Errorf("writing config.yml: %w", err)
	}
	if err := writeYAML(environmentsFilePath(fc.dir), &fc.envs); err != nil {
		return fmt.Errorf("writing environments.yml: %w", err)
	}
	if err := writeYAML(tokensFilePath(fc.dir), &fc.toks); err != nil {
		return fmt.Errorf("writing tokens.yml: %w", err)
	}
	return nil
}

func readYAML(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, v)
}

func writeYAML(path string, v interface{}) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
