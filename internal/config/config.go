package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const appDirName = "totonou"

type Config struct {
	BacklogSpaceURL     string
	BacklogAPIKey       string
	BacklogDoneStatusID int
	GitHubRepo          string
	DefaultBase         string
	BranchPattern       string
	ProjectKey          string
}

func Default() Config {
	return Config{
		DefaultBase:   "develop",
		BranchPattern: "feature/{issueKey}",
	}
}

func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appDirName), nil
}

func DefaultEnvPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".env"), nil
}

func DefaultTreePath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "tree.json"), nil
}

func Load() (Config, error) {
	if err := loadEnvFiles(); err != nil {
		return Config{}, err
	}

	cfg := Default()
	applyEnv(&cfg)
	return cfg, nil
}

func loadEnvFiles() error {
	paths := make([]string, 0, 2)

	if customPath := os.Getenv("TOTONOU_ENV_FILE"); customPath != "" {
		paths = append(paths, customPath)
	}

	defaultEnvPath, err := DefaultEnvPath()
	if err != nil {
		return err
	}
	paths = append(paths, defaultEnvPath)

	for _, path := range paths {
		if err := loadEnvFile(path); err != nil {
			return fmt.Errorf("read env file %s: %w", path, err)
		}
	}
	return nil
}

func loadEnvFile(path string) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}

	for lineNumber, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("invalid env format at line %d", lineNumber+1)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = unquoteEnvValue(value)

		if key == "" {
			return fmt.Errorf("empty env key at line %d", lineNumber+1)
		}
		if os.Getenv(key) != "" {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set env %s: %w", key, err)
		}
	}
	return nil
}

func unquoteEnvValue(value string) string {
	if len(value) < 2 {
		return value
	}
	if value[0] == '"' && value[len(value)-1] == '"' {
		return value[1 : len(value)-1]
	}
	if value[0] == '\'' && value[len(value)-1] == '\'' {
		return value[1 : len(value)-1]
	}
	return value
}

func (c Config) ValidateBacklog() error {
	var missing []string
	if c.BacklogSpaceURL == "" {
		missing = append(missing, "BACKLOG_SPACE_URL")
	}
	if c.BacklogAPIKey == "" {
		missing = append(missing, "BACKLOG_API_KEY")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing Backlog config in .env: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (c Config) ValidateDoneStatus() error {
	if err := c.ValidateBacklog(); err != nil {
		return err
	}
	if c.BacklogDoneStatusID == 0 {
		return errors.New("missing Backlog done status in .env: BACKLOG_DONE_STATUS_ID")
	}
	return nil
}

func applyEnv(cfg *Config) {
	if value := os.Getenv("BACKLOG_SPACE_URL"); value != "" {
		cfg.BacklogSpaceURL = value
	}
	if value := os.Getenv("BACKLOG_API_KEY"); value != "" {
		cfg.BacklogAPIKey = value
	}
	if value := os.Getenv("BACKLOG_DONE_STATUS_ID"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			cfg.BacklogDoneStatusID = parsed
		}
	}
	if value := os.Getenv("GITHUB_REPO"); value != "" {
		cfg.GitHubRepo = value
	}
	if value := os.Getenv("TOTONOU_DEFAULT_BASE"); value != "" {
		cfg.DefaultBase = value
	}
	if value := os.Getenv("TOTONOU_BRANCH_PATTERN"); value != "" {
		cfg.BranchPattern = value
	}
	if value := os.Getenv("TOTONOU_PROJECT_KEY"); value != "" {
		cfg.ProjectKey = value
	}
}
