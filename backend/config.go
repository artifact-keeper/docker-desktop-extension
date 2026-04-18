package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

type ServiceConfig struct {
	Meilisearch     bool `json:"meilisearch"`
	Trivy           bool `json:"trivy"`
	OpenSCAP        bool `json:"openscap"`
	DependencyTrack bool `json:"dependencyTrack"`
	Jaeger          bool `json:"jaeger"`
}

type Config struct {
	Port           int           `json:"port"`
	BindAddress    string        `json:"bindAddress"`
	ExposePostgres bool          `json:"exposePostgres"`
	Services       ServiceConfig `json:"services"`
}

type Secrets struct {
	JWTSecret     string `json:"jwtSecret"`
	MeilisearchKey string `json:"meilisearchKey"`
	AdminPassword string `json:"adminPassword"`
}

func DefaultConfig() Config {
	return Config{
		Port:           8080,
		BindAddress:    "127.0.0.1",
		ExposePostgres: false,
		Services: ServiceConfig{
			Meilisearch:     true,
			Trivy:           false,
			OpenSCAP:        false,
			DependencyTrack: false,
			Jaeger:          false,
		},
	}
}

func LoadConfig(path string) (Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), err
	}

	return cfg, nil
}

func SaveConfig(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func LoadOrCreateSecrets(path string) (Secrets, bool, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		var s Secrets
		if err := json.Unmarshal(data, &s); err != nil {
			return Secrets{}, false, err
		}
		return s, false, nil
	}

	if !os.IsNotExist(err) {
		return Secrets{}, false, err
	}

	jwtSecret, err := randomHex(32)
	if err != nil {
		return Secrets{}, false, err
	}

	meiliKey, err := randomHex(16)
	if err != nil {
		return Secrets{}, false, err
	}

	adminPass, err := randomHex(12)
	if err != nil {
		return Secrets{}, false, err
	}

	s := Secrets{
		JWTSecret:      jwtSecret,
		MeilisearchKey: meiliKey,
		AdminPassword:  adminPass,
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return Secrets{}, false, err
	}

	secretData, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return Secrets{}, false, err
	}

	if err := os.WriteFile(path, secretData, 0644); err != nil {
		return Secrets{}, false, err
	}

	// Write the admin password as a plain-text sidecar so the backend's
	// entrypoint wrapper can read it without parsing JSON.
	passwordPath := filepath.Join(filepath.Dir(path), "admin_password")
	if err := os.WriteFile(passwordPath, []byte(adminPass), 0644); err != nil {
		return Secrets{}, false, err
	}

	return s, true, nil
}
