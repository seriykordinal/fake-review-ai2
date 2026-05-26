package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_FileNotFound(t *testing.T) {
	// Сохраняем текущую директорию и переходим во временную
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when settings.json not found")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.WriteFile(filepath.Join(tmpDir, "settings.json"), []byte("not json"), 0644)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoad_ValidJSON(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	settings := `{
		"database": {"host":"localhost","port":5432,"user":"testuser","password":"testpass","dbname":"testdb"},
		"go_server": {"port": 8080},
		"python_server": {"host": "localhost", "port": 8000},
		"jwt": {"secret": "test-secret"},
		"email": {"from":"test@test.com","password":"pass","smtp_host":"smtp.test.com","smtp_port":"587"},
		"email_verification_enabled": true,
		"debug": true
	}`
	os.WriteFile(filepath.Join(tmpDir, "settings.json"), []byte(settings), 0644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Database.Host != "localhost" {
		t.Errorf("Database.Host = %q, want localhost", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("Database.Port = %d, want 5432", cfg.Database.Port)
	}
	if cfg.GoServer.Port != 8080 {
		t.Errorf("GoServer.Port = %d, want 8080", cfg.GoServer.Port)
	}
	if cfg.PythonServer.Host != "localhost" {
		t.Errorf("PythonServer.Host = %q", cfg.PythonServer.Host)
	}
	if cfg.PythonServer.Port != 8000 {
		t.Errorf("PythonServer.Port = %d", cfg.PythonServer.Port)
	}
	if cfg.JWT.Secret != "test-secret" {
		t.Errorf("JWT.Secret = %q", cfg.JWT.Secret)
	}
	if !cfg.EmailVerificationEnabled {
		t.Error("EmailVerificationEnabled should be true")
	}
	if !cfg.Debug {
		t.Error("Debug should be true")
	}

	// Global должен быть установлен
	if Global == nil {
		t.Fatal("Global config not set")
	}
	if Global.JWT.Secret != "test-secret" {
		t.Errorf("Global.JWT.Secret = %q", Global.JWT.Secret)
	}
}

func TestLoad_DebugFalse_DiscardsLog(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	settings := `{
		"database": {"host":"localhost","port":5432,"user":"u","password":"p","dbname":"d"},
		"go_server": {"port": 8080},
		"python_server": {"host": "localhost", "port": 8000},
		"jwt": {"secret": "s"},
		"email": {"from":"f","password":"p","smtp_host":"h","smtp_port":"587"},
		"debug": false
	}`
	os.WriteFile(filepath.Join(tmpDir, "settings.json"), []byte(settings), 0644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Debug {
		t.Error("Debug should be false")
	}
}
