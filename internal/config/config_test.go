package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsSettingsFromEnvFile(t *testing.T) {
	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte(`BACKLOG_SPACE_URL=https://example.backlog.com
BACKLOG_API_KEY=secret-from-env
BACKLOG_DONE_STATUS_ID=5
GITHUB_REPO=owner/repo
GITWORK_DEFAULT_BASE=main
GITWORK_BRANCH_PATTERN=feature/test/{issueKey}
GITWORK_PROJECT_KEY=TEST
`), 0o600); err != nil {
		t.Fatal(err)
	}

	clearConfigEnv(t)
	t.Setenv("GITWORK_ENV_FILE", envPath)

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
	t.Setenv("GITWORK_ENV_FILE", filepath.Join(t.TempDir(), "missing.env"))

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
	applyEnv(&cfg)

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

func clearConfigEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"GITWORK_ENV_FILE",
		"BACKLOG_SPACE_URL",
		"BACKLOG_API_KEY",
		"BACKLOG_DONE_STATUS_ID",
		"GITHUB_REPO",
		"GITWORK_DEFAULT_BASE",
		"GITWORK_BRANCH_PATTERN",
		"GITWORK_PROJECT_KEY",
	} {
		t.Setenv(key, "")
	}
}
