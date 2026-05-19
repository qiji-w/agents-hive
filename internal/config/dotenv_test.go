package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv_doesNotOverrideExistingEnv(t *testing.T) {
	t.Setenv("POSTGRES_PASSWORD", "from-shell")
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("POSTGRES_PASSWORD=from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	LoadDotEnv(envPath)
	if got := os.Getenv("POSTGRES_PASSWORD"); got != "from-shell" {
		t.Fatalf("expected from-shell, got %q", got)
	}
}

func TestLoadDotEnv_setsUnsetVars(t *testing.T) {
	const key = "HIVE_DOTENV_TEST_KEY"
	os.Unsetenv(key)
	t.Cleanup(func() { os.Unsetenv(key) })

	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte(key+"=hello\n# comment\n\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	LoadDotEnv(envPath)
	if got := os.Getenv(key); got != "hello" {
		t.Fatalf("expected hello, got %q", got)
	}
}

func TestEnsurePostgresPasswordFromDotEnv_overridesStaleShellExport(t *testing.T) {
	t.Setenv("POSTGRES_PASSWORD", "wrong-from-shell")
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("POSTGRES_PASSWORD=from-dotenv\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(dir, "config.json")
	cfg := &Config{Store: StoreConfig{Postgres: PostgresConfig{Password: "wrong-from-shell"}}}
	cfg.ensurePostgresPasswordFromDotEnv(cfgPath)
	if cfg.Store.Postgres.Password != "from-dotenv" {
		t.Fatalf("password = %q, want from-dotenv", cfg.Store.Postgres.Password)
	}
}

func TestLoadDotEnvForConfig_expandsInLoad(t *testing.T) {
	const key = "HIVE_DOTENV_PG_PASS"
	os.Unsetenv(key)
	t.Cleanup(func() { os.Unsetenv(key) })

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(key+"=secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(cfgPath, []byte(`{"store":{"type":"postgres","postgres":{"password":"${`+key+`}"}}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Store.Postgres.Password != "secret" {
		t.Fatalf("password = %q, want secret", cfg.Store.Postgres.Password)
	}
}
