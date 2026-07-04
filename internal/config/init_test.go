package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteEnvTemplateCreatesFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	envPath := filepath.Join(dir, "gitwork", ".env")

	if err := WriteEnvTemplate(envPath); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != EnvTemplate() {
		t.Fatalf("unexpected template content:\n%s", data)
	}

	info, err := os.Stat(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected file mode: %o", info.Mode().Perm())
	}
}

func TestWriteEnvTemplateReturnsErrorWhenFileExists(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := WriteEnvTemplate(envPath)
	if !errors.Is(err, ErrEnvFileExists) {
		t.Fatalf("expected ErrEnvFileExists, got %v", err)
	}
}
