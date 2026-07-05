package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultTreePath(t *testing.T) {
	dir, err := ConfigDir()
	if err != nil {
		t.Fatal(err)
	}

	treePath, err := DefaultTreePath()
	if err != nil {
		t.Fatal(err)
	}
	if treePath != filepath.Join(dir, "tree.json") {
		t.Fatalf("unexpected tree path: %s", treePath)
	}
}

func TestLoadReadsSettingsFromEnvFile(t *testing.T) {
	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte(`BACKLOG_SPACE_URL=https://example.backlog.com
BACKLOG_API_KEY=secret-from-env
BACKLOG_DONE_STATUS_ID=5
GITHUB_REPO=owner/repo
TOTONOU_DEFAULT_BASE=main
TOTONOU_BRANCH_PATTERN=feature/test/{issueKey}
TOTONOU_PROJECT_KEY=TEST
`), 0o600); err != nil {
		t.Fatal(err)
	}

	clearConfigEnv(t)
	t.Setenv("TOTONOU_ENV_FILE", envPath)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultBase != "main" {
		t.Fatalf("unexpected default base: %s", cfg.DefaultBase)
	}
	if cfg.BranchPattern != "feature/test/{issueKey}" {
		t.Fatalf("unexpected branch pattern: %s", cfg.BranchPattern)
	}
	if cfg.ProjectKey != "TEST" {
		t.Fatalf("unexpected project key: %s", cfg.ProjectKey)
	}
	if cfg.BacklogAPIKey != "secret-from-env" {
		t.Fatalf("unexpected backlog api key: %s", cfg.BacklogAPIKey)
	}
}

func TestLoadUsesDefaultBranchPatternWhenEnvIsMissing(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("TOTONOU_ENV_FILE", filepath.Join(t.TempDir(), "missing.env"))

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BranchPattern != "feature/{issueKey}" {
		t.Fatalf("unexpected branch pattern: %s", cfg.BranchPattern)
	}
}

func TestLoadEnvFileSetsMissingEnvironmentVariables(t *testing.T) {
	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte(`BACKLOG_SPACE_URL=https://example.backlog.com
BACKLOG_API_KEY=secret-from-env
BACKLOG_DONE_STATUS_ID=5
GITHUB_REPO=owner/repo
`), 0o600); err != nil {
		t.Fatal(err)
	}

	clearConfigEnv(t)

	if err := loadEnvFile(envPath); err != nil {
		t.Fatal(err)
	}

	cfg := Default()
	if err := applyEnv(&cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.BacklogSpaceURL != "https://example.backlog.com" {
		t.Fatalf("unexpected backlog url: %s", cfg.BacklogSpaceURL)
	}
	if cfg.BacklogAPIKey != "secret-from-env" {
		t.Fatalf("unexpected backlog api key: %s", cfg.BacklogAPIKey)
	}
	if cfg.BacklogDoneStatusID != 5 {
		t.Fatalf("unexpected done status: %d", cfg.BacklogDoneStatusID)
	}
	if cfg.GitHubRepo != "owner/repo" {
		t.Fatalf("unexpected github repo: %s", cfg.GitHubRepo)
	}
}

func TestLoadEnvFileDoesNotOverrideExistingEnvironmentVariables(t *testing.T) {
	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte("BACKLOG_API_KEY=from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("BACKLOG_API_KEY", "from-shell")

	if err := loadEnvFile(envPath); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("BACKLOG_API_KEY") != "from-shell" {
		t.Fatalf("expected shell env to win, got %q", os.Getenv("BACKLOG_API_KEY"))
	}
}

func TestLoadRejectsInvalidDoneStatusEnvironment(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("TOTONOU_ENV_FILE", filepath.Join(t.TempDir(), "missing.env"))
	t.Setenv("BACKLOG_DONE_STATUS_ID", "done")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid BACKLOG_DONE_STATUS_ID")
	}
	if !strings.Contains(err.Error(), "BACKLOG_DONE_STATUS_ID") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadEnvFileRejectsInvalidFormatWithLineNumber(t *testing.T) {
	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte("BACKLOG_API_KEY=ok\ninvalid-line-without-equals\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := loadEnvFile(envPath)
	if err == nil {
		t.Fatal("expected error for invalid env format")
	}
	if !strings.Contains(err.Error(), "invalid env format at line 2") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadEnvFileRejectsEmptyKeyWithLineNumber(t *testing.T) {
	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte("=value-only\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := loadEnvFile(envPath)
	if err == nil {
		t.Fatal("expected error for empty env key")
	}
	if !strings.Contains(err.Error(), "empty env key at line 1") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadRejectsInvalidEnvFileWithLineNumber(t *testing.T) {
	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte("BACKLOG_API_KEY=ok\nnot-valid\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	clearConfigEnv(t)
	t.Setenv("GITWORK_ENV_FILE", envPath)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid env file")
	}
	if !strings.Contains(err.Error(), envPath) {
		t.Fatalf("expected env file path in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "invalid env format at line 2") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadRejectsInvalidDoneStatusInEnvFile(t *testing.T) {
	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte("BACKLOG_DONE_STATUS_ID=done\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	clearConfigEnv(t)
	t.Setenv("GITWORK_ENV_FILE", envPath)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid BACKLOG_DONE_STATUS_ID in env file")
	}
	if !strings.Contains(err.Error(), "BACKLOG_DONE_STATUS_ID") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateGitHubReportsMissingRepo(t *testing.T) {
	t.Parallel()

	cfg := Config{}
	err := cfg.ValidateGitHub()
	if err == nil {
		t.Fatal("expected error for missing GITHUB_REPO")
	}
	if !strings.Contains(err.Error(), "GITHUB_REPO") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateGitHubAcceptsConfiguredRepo(t *testing.T) {
	t.Parallel()

	cfg := Config{GitHubRepo: "owner/repo"}
	if err := cfg.ValidateGitHub(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func clearConfigEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"TOTONOU_ENV_FILE",
		"BACKLOG_SPACE_URL",
		"BACKLOG_API_KEY",
		"BACKLOG_DONE_STATUS_ID",
		"GITHUB_REPO",
		"TOTONOU_DEFAULT_BASE",
		"TOTONOU_BRANCH_PATTERN",
		"TOTONOU_PROJECT_KEY",
	} {
		t.Setenv(key, "")
	}
}
