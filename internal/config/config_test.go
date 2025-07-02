// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package config_test

import (
	"encoding/json"
	"github.com/playbymail/ottomap/internal/config"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Test non-existent file
	t.Run("non-existent file", func(t *testing.T) {
		cfg, err := config.Load("non-existent-file.json", false)
		if err != nil {
			t.Errorf("expected no error for non-existent file, got %v", err)
		}
		if cfg == nil {
			t.Errorf("expected non-nil config")
		}
		// Should return default config
		if cfg.Clan != "" {
			t.Errorf("expected empty clan, got %q", cfg.Clan)
		}
	})

	// Test directory instead of file
	t.Run("directory error", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, err := config.Load(tmpDir, false)
		if err == nil {
			t.Errorf("expected error for directory, got nil")
		}
	})

	// Test empty config file
	t.Run("empty config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.json")

		err := os.WriteFile(configFile, []byte("{}"), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		cfg, err := config.Load(configFile, false)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if cfg.Clan != "" {
			t.Errorf("expected empty clan, got %q", cfg.Clan)
		}
	})

	// Test partial config loading
	t.Run("partial config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.json")

		testConfig := config.Config{
			Clan: "TestClan",
			Experimental: config.Experimental_t{
				AllowConfig: true,
			},
		}

		data, err := json.Marshal(testConfig)
		if err != nil {
			t.Fatalf("failed to marshal test config: %v", err)
		}

		err = os.WriteFile(configFile, data, 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		cfg, err := config.Load(configFile, false)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if cfg.Clan != "TestClan" {
			t.Errorf("expected clan 'TestClan', got %q", cfg.Clan)
		}
		if !cfg.Experimental.AllowConfig {
			t.Errorf("expected AllowConfig to be true")
		}
		// Nested field should remain default
		if cfg.Worldographer.Map.Layers.MPCost {
			t.Errorf("expected MPCost to be false (default)")
		}
	})

	// Test full config loading
	t.Run("full config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.json")

		testConfig := config.Config{
			Clan:        "FullTestClan",
			AllowConfig: true,
			Worldographer: config.Worldographer_t{
				Map: config.Map_t{
					Layers: config.Layers_t{
						MPCost: true,
					},
				},
			},
		}

		data, err := json.Marshal(testConfig)
		if err != nil {
			t.Fatalf("failed to marshal test config: %v", err)
		}

		err = os.WriteFile(configFile, data, 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		cfg, err := config.Load(configFile, false)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if cfg.Clan != "FullTestClan" {
			t.Errorf("expected clan 'FullTestClan', got %q", cfg.Clan)
		}
		if !cfg.AllowConfig {
			t.Errorf("expected AllowConfig to be true")
		}
		if !cfg.Worldographer.Map.Layers.MPCost {
			t.Errorf("expected MPCost to be true")
		}
	})

	// Test invalid JSON
	t.Run("invalid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.json")

		err := os.WriteFile(configFile, []byte("invalid json"), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		cfg, err := config.Load(configFile, true)
		if err != nil {
			t.Errorf("expected no error for invalid JSON, got %v", err)
		}
		// Should return default config when JSON is invalid
		if cfg.Clan != "" {
			t.Errorf("expected empty clan for invalid JSON, got %q", cfg.Clan)
		}
	})
}

func TestCopyNonZeroFields(t *testing.T) {
	// We need to test the copyNonZeroFields function indirectly through Load
	// since it's not exported. This test ensures the field copying logic works.

	t.Run("copy only non-zero fields", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.json")

		// Create a config with only some fields set
		testConfig := config.Config{
			Clan: "TestClan",
			// Don't set Experimental.AllowConfig (should remain false)
			Worldographer: config.Worldographer_t{
				Map: config.Map_t{
					Layers: config.Layers_t{
						MPCost: true,
					},
				},
			},
		}

		data, err := json.Marshal(testConfig)
		if err != nil {
			t.Fatalf("failed to marshal test config: %v", err)
		}

		err = os.WriteFile(configFile, data, 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		cfg, err := config.Load(configFile, true)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Verify that non-zero fields were copied
		if cfg.Clan != "TestClan" {
			t.Errorf("expected clan 'TestClan', got %q", cfg.Clan)
		}
		if cfg.Worldographer.Map.Layers.MPCost != true {
			t.Errorf("expected MPCost to be true")
		}

		// Verify that zero fields remained at their defaults
		if cfg.Experimental.AllowConfig != false {
			t.Errorf("expected AllowConfig to remain false (default)")
		}
	})
}
