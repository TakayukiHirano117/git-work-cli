package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type Runner func(ctx context.Context, dir string, name string, args ...string) (string, error)

type Client struct {
	Dir string
	Run Runner
}

func New(dir string) Client {
	return Client{Dir: dir, Run: runCommand}
}

func (c Client) CurrentBranch(ctx context.Context) (string, error) {
	return c.output(ctx, "git", "branch", "--show-current")
}

func (c Client) RepoRoot(ctx context.Context) (string, error) {
	return c.output(ctx, "git", "rev-parse", "--show-toplevel")
}

func (c Client) CreateBranch(ctx context.Context, branch string) error {
	_, err := c.output(ctx, "git", "switch", "-c", branch)
	return err
}

func (c Client) PushCurrentBranch(ctx context.Context, branch string) error {
	_, err := c.output(ctx, "git", "push", "-u", "origin", branch)
	return err
}

func (c Client) GHAuthStatus(ctx context.Context) error {
	_, err := c.output(ctx, "gh", "auth", "status")
	return err
}

func (c Client) CreatePullRequest(ctx context.Context, repo string, title string, body string, base string, dryRun bool) (string, error) {
	args := []string{"pr", "create", "--title", title, "--body", body, "--base", base}
	if repo != "" {
		args = append(args, "--repo", repo)
	}
	if dryRun {
		args = append(args, "--dry-run")
	}
	return c.output(ctx, "gh", args...)
}

func (c Client) output(ctx context.Context, name string, args ...string) (string, error) {
	if c.Run == nil {
		c.Run = runCommand
	}
	return c.Run(ctx, c.Dir, name, args...)
}

func runCommand(ctx context.Context, dir string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if dir != "" {
		cmd.Dir = dir
	}

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return "", fmt.Errorf("%s %s: %s", name, strings.Join(args, " "), message)
	}
	return strings.TrimSpace(out.String()), nil
}
