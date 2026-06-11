package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Config struct {
	NtfyServer      string   `json:"ntfyServer"`
	NtfyTopic       string   `json:"ntfyTopic"`
	NtfyToken       string   `json:"ntfyToken"`
	NotifyUnhealthy bool     `json:"notifyUnhealthy"`
	NotifyDown      bool     `json:"notifyDown"`
	NotifyRecovered bool     `json:"notifyRecovered"`
	Ignore          []string `json:"ignore"`
}

func defaultConfig() Config {
	return Config{
		NtfyServer:      "https://ntfy.sh",
		NotifyUnhealthy: true,
		NotifyDown:      true,
		NotifyRecovered: true,
		Ignore:          []string{},
	}
}

// ConfigStore is a thread-safe view of the config, persisted as JSON on disk.
type ConfigStore struct {
	mu   sync.RWMutex
	path string
	cfg  Config
}

func NewConfigStore(path string) (*ConfigStore, error) {
	s := &ConfigStore{path: path, cfg: defaultConfig()}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &s.cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if s.cfg.Ignore == nil {
		s.cfg.Ignore = []string{}
	}
	return s, nil
}

func (s *ConfigStore) Get() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cfg := s.cfg
	cfg.Ignore = append([]string{}, s.cfg.Ignore...)
	return cfg
}

func (s *ConfigStore) Set(cfg Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return err
	}
	s.cfg = cfg
	return nil
}
