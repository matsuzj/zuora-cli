// Package alias implements the "zr alias" command group.
package alias

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"gopkg.in/yaml.v3"
)

// Store manages alias persistence.
type Store struct {
	path string
	mu   sync.Mutex
	data map[string]string
}

// NewStore creates a Store that reads/writes aliases from the given config directory.
func NewStore(configDir string) *Store {
	return &Store{
		path: filepath.Join(configDir, "aliases.yml"),
		data: make(map[string]string),
	}
}

// Load reads aliases from disk. If the file does not exist, an empty map is used.
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	raw, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.data = make(map[string]string)
			return nil
		}
		return fmt.Errorf("reading aliases.yml: %w", err)
	}
	m := make(map[string]string)
	if err := yaml.Unmarshal(raw, &m); err != nil {
		return fmt.Errorf("parsing aliases.yml: %w", err)
	}
	if m == nil {
		m = make(map[string]string)
	}
	s.data = m
	return nil
}

// Save writes aliases to disk.
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	raw, err := yaml.Marshal(s.data)
	if err != nil {
		return fmt.Errorf("encoding aliases: %w", err)
	}
	return os.WriteFile(s.path, raw, 0600)
}

// Set adds or updates an alias.
func (s *Store) Set(name, command string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[name] = command
}

// Delete removes an alias. Returns an error if it does not exist.
func (s *Store) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[name]; !ok {
		return fmt.Errorf("alias %q not found", name)
	}
	delete(s.data, name)
	return nil
}

// Get returns the command for an alias. Returns ("", false) if not found.
func (s *Store) Get(name string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cmd, ok := s.data[name]
	return cmd, ok
}

// All returns all aliases sorted by name.
func (s *Store) All() []AliasEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries := make([]AliasEntry, 0, len(s.data))
	for name, cmd := range s.data {
		entries = append(entries, AliasEntry{Name: name, Command: cmd})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	return entries
}

// Len returns the number of aliases.
func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.data)
}

// AliasEntry is a name-command pair.
type AliasEntry struct {
	Name    string
	Command string
}
