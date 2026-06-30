package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const appDirName = "gitwork"

type Config struct {
	BacklogSpaceURL     string `json:"backlogSpaceUrl"`
	BacklogAPIKey       string `json:"backlogApiKey"`
	BacklogDoneStatusID int    `json:"backlogDoneStatusId"`
	GitHubRepo          string `json:"githubRepo"`
	DefaultBase         string `json:"defaultBase"`
	BranchPattern       string `json:"branchPattern"`
	ProjectKey          string `json:"projectKey"`
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

func DefaultPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (Config, error) {
	path, err := DefaultPath()
	if err != nil {
		return Config{}, err
	}
	return LoadFile(path)
}

func LoadFile(path string) (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			applyEnv(&cfg)
			return cfg, nil
		}
		return Config{}, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	applyEnv(&cfg)
	return cfg, nil
}

func (c Config) ValidateBacklog() error {
	var missing []string
	if c.BacklogSpaceURL == "" {
		missing = append(missing, "backlogSpaceUrl or BACKLOG_SPACE_URL")
	}
	if c.BacklogAPIKey == "" {
		missing = append(missing, "backlogApiKey or BACKLOG_API_KEY")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing Backlog config: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (c Config) ValidateDoneStatus() error {
	if err := c.ValidateBacklog(); err != nil {
		return err
	}
	if c.BacklogDoneStatusID == 0 {
		return errors.New("missing Backlog done status: backlogDoneStatusId or BACKLOG_DONE_STATUS_ID")
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
	if value := os.Getenv("GITWORK_DEFAULT_BASE"); value != "" {
		cfg.DefaultBase = value
	}
	if value := os.Getenv("GITWORK_BRANCH_PATTERN"); value != "" {
		cfg.BranchPattern = value
	}
	if value := os.Getenv("GITWORK_PROJECT_KEY"); value != "" {
		cfg.ProjectKey = value
	}
}
