package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"git-cli/internal/backlog"
	"git-cli/internal/config"
	gitcmd "git-cli/internal/git"
	"git-cli/internal/store"
)

func TestWorkCreatesBranchAndRecordsParent(t *testing.T) {
	t.Parallel()

	st := store.New(filepath.Join(t.TempDir(), "tree.json"))
	var commands []string
	app := App{
		Stdin:  strings.NewReader(""),
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Store:  st,
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			commands = append(commands, command)
			switch command {
			case "git branch --show-current":
				return "feature/member/backend/COMMUNITY-101", nil
			case "git rev-parse --show-toplevel":
				return "/repo", nil
			case "git switch -c feature/member/backend/COMMUNITY-102":
				return "", nil
			default:
				t.Fatalf("unexpected command: %s", command)
				return "", nil
			}
		}},
	}

	if err := app.Run(context.Background(), []string{"work", "community-102", "--team", "member", "--layer", "backend"}); err != nil {
		t.Fatal(err)
	}

	tree, err := st.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(tree.Records))
	}
	record := tree.Records[0]
	if record.ParentBranch != "feature/member/backend/COMMUNITY-101" {
		t.Fatalf("unexpected parent branch: %s", record.ParentBranch)
	}
	if record.ChildBranch != "feature/member/backend/COMMUNITY-102" {
		t.Fatalf("unexpected child branch: %s", record.ChildBranch)
	}
	if len(commands) != 3 {
		t.Fatalf("expected 3 git commands, got %d", len(commands))
	}
}

func TestWorkPromptsForTeamAndLayer(t *testing.T) {
	t.Parallel()

	old := stdinIsTTY
	stdinIsTTY = func(io.Reader) bool { return true }
	t.Cleanup(func() { stdinIsTTY = old })

	st := store.New(filepath.Join(t.TempDir(), "tree.json"))
	app := App{
		Stdin:  strings.NewReader("2\n1\n"),
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Store:  st,
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			switch command {
			case "git branch --show-current":
				return "develop", nil
			case "git rev-parse --show-toplevel":
				return "/repo", nil
			case "git switch -c feature/admin/frontend/COMMUNITY-200":
				return "", nil
			default:
				t.Fatalf("unexpected command: %s", command)
				return "", nil
			}
		}},
	}

	if err := app.Run(context.Background(), []string{"work", "COMMUNITY-200"}); err != nil {
		t.Fatal(err)
	}

	tree, err := st.Load()
	if err != nil {
		t.Fatal(err)
	}
	if tree.Records[0].ChildBranch != "feature/admin/frontend/COMMUNITY-200" {
		t.Fatalf("unexpected child branch: %s", tree.Records[0].ChildBranch)
	}
}

func TestWorkRejectsMissingFlagsOnNonInteractiveStdin(t *testing.T) {
	t.Parallel()

	st := store.New(filepath.Join(t.TempDir(), "tree.json"))
	var commands []string
	app := App{
		Stdin:  strings.NewReader(""),
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Store:  st,
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			commands = append(commands, command)
			switch command {
			case "git branch --show-current":
				return "develop", nil
			case "git rev-parse --show-toplevel":
				return "/repo", nil
			default:
				t.Fatalf("unexpected command: %s", command)
				return "", nil
			}
		}},
	}

	err := app.Run(context.Background(), []string{"work", "COMMUNITY-200"})
	if err == nil {
		t.Fatal("expected non-interactive error")
	}
	if !strings.Contains(err.Error(), "non-interactive stdin") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "--team") || !strings.Contains(err.Error(), "--layer") {
		t.Fatalf("expected missing flag hints, got %q", err.Error())
	}
	if len(commands) != 2 {
		t.Fatalf("expected 2 git commands before flag check, got %d: %v", len(commands), commands)
	}

	tree, err := st.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.Records) != 0 {
		t.Fatalf("expected no records, got %d", len(tree.Records))
	}
}

func TestWorkRejectsInvalidIssueKey(t *testing.T) {
	t.Parallel()

	st := store.New(filepath.Join(t.TempDir(), "tree.json"))
	var commands []string
	app := App{
		Stdin:  strings.NewReader(""),
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Store:  st,
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			commands = append(commands, command)
			switch command {
			case "git branch --show-current":
				return "develop", nil
			case "git rev-parse --show-toplevel":
				return "/repo", nil
			default:
				t.Fatalf("unexpected command: %s", command)
				return "", nil
			}
		}},
	}

	err := app.Run(context.Background(), []string{"work", "invalid-key", "--team", "member", "--layer", "backend"})
	if err == nil {
		t.Fatal("expected invalid issue key error")
	}
	if !strings.Contains(err.Error(), "invalid issue key") {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(commands) != 0 {
		t.Fatalf("expected no git commands, got %d: %v", len(commands), commands)
	}

	tree, err := st.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.Records) != 0 {
		t.Fatalf("expected no records, got %d", len(tree.Records))
	}
}

func TestWorkShowsParentWhenBranchAlreadyRecorded(t *testing.T) {
	t.Parallel()

	st := store.New(filepath.Join(t.TempDir(), "tree.json"))
	existing := store.Record{
		RepoRoot:     "/repo",
		ParentBranch: "feature/member/backend/COMMUNITY-100",
		ChildBranch:  "feature/member/backend/COMMUNITY-102",
		IssueKey:     "COMMUNITY-102",
	}
	if err := st.Add(existing); err != nil {
		t.Fatal(err)
	}

	var commands []string
	app := App{
		Stdin:  strings.NewReader(""),
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Store:  st,
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			commands = append(commands, command)
			switch command {
			case "git branch --show-current":
				return "develop", nil
			case "git rev-parse --show-toplevel":
				return "/repo", nil
			default:
				t.Fatalf("unexpected command: %s", command)
				return "", nil
			}
		}},
	}

	err := app.Run(context.Background(), []string{"work", "COMMUNITY-102", "--team", "member", "--layer", "backend"})
	if err == nil {
		t.Fatal("expected duplicate branch error")
	}
	if !strings.Contains(err.Error(), "feature/member/backend/COMMUNITY-102") {
		t.Fatalf("expected child branch in error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "parent: feature/member/backend/COMMUNITY-100") {
		t.Fatalf("expected parent branch in error, got %q", err.Error())
	}
	if len(commands) != 2 {
		t.Fatalf("expected 2 git commands before duplicate check, got %d: %v", len(commands), commands)
	}
}

func TestPRCreatesPullRequestAndUpdatesBacklog(t *testing.T) {
	t.Parallel()

	updated := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("apiKey") != "secret" {
			t.Fatalf("missing api key")
		}
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/issues/COMMUNITY-102":
			writeJSON(t, w, map[string]interface{}{
				"issueKey": "COMMUNITY-102",
				"summary":  "API利用画面を実装",
				"status": map[string]interface{}{
					"name": "対応中",
				},
			})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v2/issues/COMMUNITY-102":
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			if r.Form.Get("statusId") != "5" {
				t.Fatalf("unexpected statusId: %s", r.Form.Get("statusId"))
			}
			updated = true
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	var commands []string
	out := &bytes.Buffer{}
	app := App{
		Stdin:  strings.NewReader("y\n"),
		Stdout: out,
		Stderr: &bytes.Buffer{},
		Config: config.Config{
			BacklogSpaceURL:     server.URL,
			BacklogAPIKey:       "secret",
			BacklogDoneStatusID: 5,
			DefaultBase:         "develop",
		},
		Store: store.New(filepath.Join(t.TempDir(), "tree.json")),
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			commands = append(commands, command)
			switch {
			case command == "git branch --show-current":
				return "feature/member/backend/COMMUNITY-102", nil
			case command == "git push -u origin feature/member/backend/COMMUNITY-102":
				return "", nil
			case strings.HasPrefix(command, "gh pr create "):
				return "https://github.com/example/repo/pull/1", nil
			default:
				t.Fatalf("unexpected command: %s", command)
				return "", nil
			}
		}},
		Backlog: backlog.Client{SpaceURL: server.URL, APIKey: "secret", HTTPClient: server.Client()},
	}

	if err := app.Run(context.Background(), []string{"pr"}); err != nil {
		t.Fatal(err)
	}
	if !updated {
		t.Fatal("expected Backlog status update")
	}
	if !strings.Contains(out.String(), "https://github.com/example/repo/pull/1") {
		t.Fatalf("expected PR URL in output, got %q", out.String())
	}
	if len(commands) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(commands))
	}
}

func TestPRYesSkipsConfirmationPrompt(t *testing.T) {
	t.Parallel()

	updated := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("apiKey") != "secret" {
			t.Fatalf("missing api key")
		}
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/issues/COMMUNITY-102":
			writeJSON(t, w, map[string]interface{}{
				"issueKey": "COMMUNITY-102",
				"summary":  "API利用画面を実装",
				"status": map[string]interface{}{
					"name": "対応中",
				},
			})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v2/issues/COMMUNITY-102":
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			if r.Form.Get("statusId") != "5" {
				t.Fatalf("unexpected statusId: %s", r.Form.Get("statusId"))
			}
			updated = true
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	var commands []string
	out := &bytes.Buffer{}
	app := App{
		Stdin:  strings.NewReader("n\n"),
		Stdout: out,
		Stderr: &bytes.Buffer{},
		Config: config.Config{
			BacklogSpaceURL:     server.URL,
			BacklogAPIKey:       "secret",
			BacklogDoneStatusID: 5,
			DefaultBase:         "develop",
		},
		Store: store.New(filepath.Join(t.TempDir(), "tree.json")),
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			commands = append(commands, command)
			switch {
			case command == "git branch --show-current":
				return "feature/member/backend/COMMUNITY-102", nil
			case command == "git push -u origin feature/member/backend/COMMUNITY-102":
				return "", nil
			case strings.HasPrefix(command, "gh pr create "):
				return "https://github.com/example/repo/pull/1", nil
			default:
				t.Fatalf("unexpected command: %s", command)
				return "", nil
			}
		}},
		Backlog: backlog.Client{SpaceURL: server.URL, APIKey: "secret", HTTPClient: server.Client()},
	}

	if err := app.Run(context.Background(), []string{"pr", "--yes"}); err != nil {
		t.Fatal(err)
	}
	if !updated {
		t.Fatal("expected Backlog status update")
	}
	output := out.String()
	if strings.Contains(output, "Create pull request?") {
		t.Fatalf("confirmation prompt should be skipped with --yes, got:\n%s", output)
	}
	if strings.Contains(output, "cancelled") {
		t.Fatalf("PR creation should not be cancelled with --yes, got:\n%s", output)
	}
	if !strings.Contains(output, "https://github.com/example/repo/pull/1") {
		t.Fatalf("expected PR URL in output, got %q", output)
	}
	if len(commands) != 3 {
		t.Fatalf("expected 3 commands, got %d: %v", len(commands), commands)
	}
}

func TestPRNoCancelsCreation(t *testing.T) {
	t.Parallel()

	updated := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("apiKey") != "secret" {
			t.Fatalf("missing api key")
		}
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/issues/COMMUNITY-102":
			writeJSON(t, w, map[string]interface{}{
				"issueKey": "COMMUNITY-102",
				"summary":  "API利用画面を実装",
				"status": map[string]interface{}{
					"name": "対応中",
				},
			})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v2/issues/COMMUNITY-102":
			updated = true
			t.Fatal("Backlog status update should not run when user declines")
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	var commands []string
	out := &bytes.Buffer{}
	app := App{
		Stdin:  strings.NewReader("n\n"),
		Stdout: out,
		Stderr: &bytes.Buffer{},
		Config: config.Config{
			BacklogSpaceURL:     server.URL,
			BacklogAPIKey:       "secret",
			BacklogDoneStatusID: 5,
			DefaultBase:         "develop",
		},
		Store: store.New(filepath.Join(t.TempDir(), "tree.json")),
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			commands = append(commands, command)
			switch command {
			case "git branch --show-current":
				return "feature/member/backend/COMMUNITY-102", nil
			case "git push -u origin feature/member/backend/COMMUNITY-102":
				t.Fatal("git push should not run when user declines")
			default:
				if strings.HasPrefix(command, "gh pr create ") {
					t.Fatal("gh pr create should not run when user declines")
				}
				t.Fatalf("unexpected command: %s", command)
			}
			return "", nil
		}},
		Backlog: backlog.Client{SpaceURL: server.URL, APIKey: "secret", HTTPClient: server.Client()},
	}

	if err := app.Run(context.Background(), []string{"pr"}); err != nil {
		t.Fatal(err)
	}
	if updated {
		t.Fatal("expected no Backlog status update")
	}
	output := out.String()
	if !strings.Contains(output, "Create pull request?") {
		t.Fatalf("expected confirmation prompt, got:\n%s", output)
	}
	if !strings.Contains(output, "cancelled") {
		t.Fatalf("expected cancelled message, got:\n%s", output)
	}
	if strings.Contains(output, "https://github.com") {
		t.Fatalf("PR URL should not appear when cancelled, got:\n%s", output)
	}
	if len(commands) != 1 {
		t.Fatalf("expected only current branch lookup, got %d: %v", len(commands), commands)
	}
}

func TestPRDryRunPrintsPreviewWithoutSideEffects(t *testing.T) {
	t.Parallel()

	backlogCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			backlogCalled = true
		}
		if r.Method != http.MethodGet || r.URL.Path != "/api/v2/issues/COMMUNITY-102" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		writeJSON(t, w, map[string]interface{}{
			"issueKey": "COMMUNITY-102",
			"summary":  "API利用画面を実装",
			"status": map[string]interface{}{
				"name": "対応中",
			},
		})
	}))
	defer server.Close()

	var commands []string
	out := &bytes.Buffer{}
	app := App{
		Stdin:  strings.NewReader(""),
		Stdout: out,
		Stderr: &bytes.Buffer{},
		Config: config.Config{
			BacklogSpaceURL:     server.URL,
			BacklogAPIKey:       "secret",
			BacklogDoneStatusID: 5,
			DefaultBase:         "develop",
			GitHubRepo:          "owner/repo",
		},
		Store: store.New(filepath.Join(t.TempDir(), "tree.json")),
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			commands = append(commands, command)
			if command == "git branch --show-current" {
				return "feature/member/backend/COMMUNITY-102", nil
			}
			t.Fatalf("unexpected command: %s", command)
			return "", nil
		}},
		Backlog: backlog.Client{SpaceURL: server.URL, APIKey: "secret", HTTPClient: server.Client()},
	}

	if err := app.Run(context.Background(), []string{"pr", "--dry-run"}); err != nil {
		t.Fatal(err)
	}

	output := out.String()
	for _, want := range []string{
		"=== Pull Request preview (dry-run) ===",
		"Title:",
		"API利用画面を実装",
		"Base:",
		"develop",
		"Body:",
		"## Backlog",
		"Commands (not executed):",
		"git push -u origin feature/member/backend/COMMUNITY-102",
		`gh pr create --title "API利用画面を実装" --body <above> --base develop --repo owner/repo`,
		"Backlog: update COMMUNITY-102 status to 5",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
	if len(commands) != 1 {
		t.Fatalf("expected only current branch lookup, got %d commands: %v", len(commands), commands)
	}
	if backlogCalled {
		t.Fatal("Backlog status update should not run in dry-run")
	}
}

func TestTodayPrintsChildrenWithBacklogStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v2/issues/COMMUNITY-103" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		writeJSON(t, w, map[string]interface{}{
			"issueKey": "COMMUNITY-103",
			"summary":  "テストを書く",
			"status": map[string]interface{}{
				"name": "未着手",
			},
		})
	}))
	defer server.Close()

	treePath := filepath.Join(t.TempDir(), "tree.json")
	st := store.New(treePath)
	if err := st.Save(store.Tree{Records: []store.Record{{
		RepoRoot:     "/repo",
		ParentBranch: "feature/member/backend/COMMUNITY-102",
		ChildBranch:  "feature/member/backend/COMMUNITY-103",
		IssueKey:     "COMMUNITY-103",
	}}}); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	app := App{
		Stdin:  strings.NewReader(""),
		Stdout: out,
		Stderr: &bytes.Buffer{},
		Config: config.Config{BacklogSpaceURL: server.URL, BacklogAPIKey: "secret"},
		Store:  st,
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			switch command {
			case "git branch --show-current":
				return "feature/member/backend/COMMUNITY-102", nil
			case "git rev-parse --show-toplevel":
				return "/repo", nil
			default:
				t.Fatalf("unexpected command: %s", command)
				return "", nil
			}
		}},
		Backlog: backlog.Client{SpaceURL: server.URL, APIKey: "secret", HTTPClient: server.Client()},
	}

	if err := app.Run(context.Background(), []string{"today"}); err != nil {
		t.Fatal(err)
	}
	output := out.String()
	if !strings.Contains(output, "COMMUNITY-103  feature/member/backend/COMMUNITY-103  テストを書く  未着手") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestTodayNoBacklogSkipsBacklogAPI(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Backlog API should not be called with --no-backlog")
	}))
	defer server.Close()

	treePath := filepath.Join(t.TempDir(), "tree.json")
	st := store.New(treePath)
	if err := st.Save(store.Tree{Records: []store.Record{{
		RepoRoot:     "/repo",
		ParentBranch: "feature/member/backend/COMMUNITY-102",
		ChildBranch:  "feature/member/backend/COMMUNITY-103",
		IssueKey:     "COMMUNITY-103",
	}}}); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	app := App{
		Stdin:  strings.NewReader(""),
		Stdout: out,
		Stderr: &bytes.Buffer{},
		Config: config.Config{BacklogSpaceURL: server.URL, BacklogAPIKey: "secret"},
		Store:  st,
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			switch command {
			case "git branch --show-current":
				return "feature/member/backend/COMMUNITY-102", nil
			case "git rev-parse --show-toplevel":
				return "/repo", nil
			default:
				t.Fatalf("unexpected command: %s", command)
				return "", nil
			}
		}},
		Backlog: backlog.Client{SpaceURL: server.URL, APIKey: "secret", HTTPClient: server.Client()},
	}

	if err := app.Run(context.Background(), []string{"today", "--no-backlog"}); err != nil {
		t.Fatal(err)
	}
	output := out.String()
	if !strings.Contains(output, "COMMUNITY-103  feature/member/backend/COMMUNITY-103  -  -") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestTodayFailsWithCorruptTreeJSON(t *testing.T) {
	t.Parallel()

	treePath := filepath.Join(t.TempDir(), "tree.json")
	if err := os.WriteFile(treePath, []byte("{broken"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := App{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Store:  store.New(treePath),
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			switch command {
			case "git branch --show-current":
				return "feature/member/backend/COMMUNITY-102", nil
			case "git rev-parse --show-toplevel":
				return "/repo", nil
			default:
				t.Fatalf("unexpected command: %s", command)
				return "", nil
			}
		}},
	}

	err := app.Run(context.Background(), []string{"today"})
	if err == nil {
		t.Fatal("expected error for corrupt tree.json")
	}
	for _, want := range []string{
		treePath,
		"invalid tree.json",
		"fix the JSON or remove the file",
		"gitwork config path",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected error to contain %q, got %q", want, err.Error())
		}
	}
}

func TestHelpPrintsGeneralUsage(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	app := App{Stdout: out, loadDeps: false}

	if err := app.Run(context.Background(), []string{"help"}); err != nil {
		t.Fatal(err)
	}
	output := out.String()
	for _, want := range []string{
		"work <issue-key>",
		"課題キーから作業用ブランチを作成",
		"Pull Request を作成",
		"今日見るべき子タスク",
		"epic status",
		"config path",
		"gitwork init",
		"よくある流れ",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected help to contain %q, got %q", want, output)
		}
	}
}

func TestHelpPrintsSubcommandUsage(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	app := App{Stdout: out, loadDeps: false}

	if err := app.Run(context.Background(), []string{"help", "pr"}); err != nil {
		t.Fatal(err)
	}
	output := out.String()
	if !strings.Contains(output, "--dry-run") || !strings.Contains(output, "--yes") {
		t.Fatalf("expected pr help flags, got %q", output)
	}
}

func TestHelpPrintsConfigSubcommandUsage(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	app := App{Stdout: out, loadDeps: false}

	if err := app.Run(context.Background(), []string{"help", "config"}); err != nil {
		t.Fatal(err)
	}
	output := out.String()
	for _, want := range []string{
		"config path",
		"gitwork config path",
		".env",
		"tree.json",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected config help to contain %q, got %q", want, output)
		}
	}
}

func TestConfigPathPrintsConfigAndTreeLocations(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	app := App{Stdout: out, loadDeps: false}

	if err := app.Run(context.Background(), []string{"config", "path"}); err != nil {
		t.Fatal(err)
	}

	configDir, err := config.ConfigDir()
	if err != nil {
		t.Fatal(err)
	}

	output := out.String()
	wantConfig := filepath.Join(configDir, ".env")
	wantTree := filepath.Join(configDir, "tree.json")
	if !strings.Contains(output, wantConfig) {
		t.Fatalf("expected config path %q in output, got %q", wantConfig, output)
	}
	if !strings.Contains(output, wantTree) {
		t.Fatalf("expected tree path %q in output, got %q", wantTree, output)
	}
}

func TestConfigPathRejectsUnknownSubcommand(t *testing.T) {
	t.Parallel()

	app := App{Stdout: &bytes.Buffer{}, loadDeps: false}
	if err := app.Run(context.Background(), []string{"config", "show"}); err == nil {
		t.Fatal("expected error for unknown config subcommand")
	}
}

func TestInitCreatesEnvTemplateWhenConfirmed(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	out := &bytes.Buffer{}
	app := App{
		Stdin:    strings.NewReader("y\n"),
		Stdout:   out,
		loadDeps: false,
	}

	if err := app.Run(context.Background(), []string{"init"}); err != nil {
		t.Fatal(err)
	}

	envPath := filepath.Join(configHome, "gitwork", ".env")
	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != config.EnvTemplate() {
		t.Fatalf("unexpected template content:\n%s", data)
	}

	output := out.String()
	for _, want := range []string{
		envPath,
		"Create .env template?",
		"created " + envPath,
		"gitwork doctor",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got %q", want, output)
		}
	}
}

func TestInitSkipsCreationWhenDeclined(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	app := App{
		Stdin:    strings.NewReader("n\n"),
		Stdout:   &bytes.Buffer{},
		loadDeps: false,
	}

	if err := app.Run(context.Background(), []string{"init"}); err != nil {
		t.Fatal(err)
	}

	envPath := filepath.Join(configHome, "gitwork", ".env")
	if _, err := os.Stat(envPath); !os.IsNotExist(err) {
		t.Fatalf("expected no env file, got err=%v", err)
	}
}

func TestInitDoesNotOverwriteExistingEnv(t *testing.T) {
	configHome := t.TempDir()
	configDir := filepath.Join(configHome, "gitwork")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	envPath := filepath.Join(configDir, ".env")
	if err := os.WriteFile(envPath, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", configHome)

	out := &bytes.Buffer{}
	app := App{
		Stdin:    strings.NewReader("y\n"),
		Stdout:   out,
		loadDeps: false,
	}

	if err := app.Run(context.Background(), []string{"init"}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "existing" {
		t.Fatalf("expected existing file to remain, got %q", data)
	}
	if !strings.Contains(out.String(), "既に存在します") {
		t.Fatalf("expected existing file message, got %q", out.String())
	}
}

func TestInitRejectsExtraArgs(t *testing.T) {
	t.Parallel()

	app := App{
		Stdout:   &bytes.Buffer{},
		loadDeps: false,
	}
	if err := app.Run(context.Background(), []string{"init", "--force"}); err == nil {
		t.Fatal("expected error for extra init args")
	}
}

func TestDoctorAllChecksPass(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	app := App{
		Stdout: out,
		Config: config.Config{
			BacklogSpaceURL:     "https://example.backlog.com",
			BacklogAPIKey:       "key",
			BacklogDoneStatusID: 5,
		},
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			switch command {
			case "git rev-parse --show-toplevel":
				return "/repo", nil
			case "gh auth status":
				return "", nil
			default:
				t.Fatalf("unexpected command: %s", command)
				return "", nil
			}
		}},
		loadDeps: false,
	}

	if err := app.Run(context.Background(), []string{"doctor"}); err != nil {
		t.Fatal(err)
	}

	output := out.String()
	for _, want := range []string{
		"git repository: ok (/repo)",
		"gh auth: ok",
		"backlog config: ok",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got %q", want, output)
		}
	}
}

func TestDoctorReportsFailures(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	app := App{
		Stdout: out,
		Config: config.Config{},
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			switch command {
			case "git rev-parse --show-toplevel":
				return "", fmt.Errorf("git rev-parse --show-toplevel: not a git repository")
			case "gh auth status":
				return "", fmt.Errorf("gh auth status: not logged in")
			default:
				t.Fatalf("unexpected command: %s", command)
				return "", nil
			}
		}},
		loadDeps: false,
	}

	err := app.Run(context.Background(), []string{"doctor"})
	if err == nil {
		t.Fatal("expected doctor to fail when checks fail")
	}
	if !strings.Contains(err.Error(), "3 check(s) failed") {
		t.Fatalf("expected failure count in error, got %q", err.Error())
	}

	output := out.String()
	for _, want := range []string{
		"git repository: not ok",
		"gh auth: not ok",
		"backlog config: not ok",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got %q", want, output)
		}
	}
}

func TestDoctorRejectsExtraArgs(t *testing.T) {
	t.Parallel()

	app := App{
		Stdout:   &bytes.Buffer{},
		loadDeps: false,
	}
	if err := app.Run(context.Background(), []string{"doctor", "--verbose"}); err == nil {
		t.Fatal("expected error for extra doctor args")
	}
}

func TestIssueKeyFromBranch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		branch  string
		wantKey string
		wantErr string
	}{
		{
			name:    "standard work branch",
			branch:  "feature/member/backend/COMMUNITY-102",
			wantKey: "COMMUNITY-102",
		},
		{
			name:    "lowercase issue key",
			branch:  "feature/admin/frontend/community-200",
			wantKey: "COMMUNITY-200",
		},
		{
			name:    "missing issue key",
			branch:  "feature/member/backend",
			wantErr: `issue key not found in branch "feature/member/backend" (expected format, e.g. feature/member/backend/COMMUNITY-102)`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			issueKey, err := issueKeyFromBranch(tt.branch)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("unexpected error: %q", err.Error())
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if issueKey != tt.wantKey {
				t.Fatalf("unexpected issue key: %s", issueKey)
			}
		})
	}
}

func TestPRShowsBranchNameHintWhenIssueKeyMissing(t *testing.T) {
	t.Parallel()

	app := App{
		Stdin:  strings.NewReader(""),
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Config: config.Config{
			BacklogSpaceURL:     "https://example.backlog.com",
			BacklogAPIKey:       "secret",
			BacklogDoneStatusID: 5,
		},
		Store:  store.New(filepath.Join(t.TempDir(), "tree.json")),
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			if name+" "+strings.Join(args, " ") == "git branch --show-current" {
				return "feature/member/backend", nil
			}
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}},
	}

	err := app.Run(context.Background(), []string{"pr"})
	if err == nil {
		t.Fatal("expected error")
	}
	want := `issue key not found in branch "feature/member/backend" (expected format, e.g. feature/member/backend/COMMUNITY-102)`
	if err.Error() != want {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}

func TestEpicStatusShowsBranchNameHintWhenIssueKeyMissing(t *testing.T) {
	t.Parallel()

	app := App{
		Stdin:    strings.NewReader(""),
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
		loadDeps: false,
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			if name+" "+strings.Join(args, " ") == "git branch --show-current" {
				return "feature/member/backend", nil
			}
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}},
	}

	err := app.Run(context.Background(), []string{"epic", "status"})
	if err == nil {
		t.Fatal("expected error")
	}
	want := `issue key not found in branch "feature/member/backend" (expected format, e.g. feature/member/backend/COMMUNITY-102)`
	if err.Error() != want {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}

func TestEpicStatusRejectsInvalidEpicKey(t *testing.T) {
	t.Parallel()

	var commands []string
	app := App{
		Stdin:    strings.NewReader(""),
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
		loadDeps: false,
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			commands = append(commands, command)
			t.Fatalf("unexpected command: %s", command)
			return "", nil
		}},
	}

	err := app.Run(context.Background(), []string{"epic", "status", "invalid-key"})
	if err == nil {
		t.Fatal("expected invalid issue key error")
	}
	if !strings.Contains(err.Error(), "invalid issue key") {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(commands) != 0 {
		t.Fatalf("expected no git commands, got %d: %v", len(commands), commands)
	}
}

func TestPRFailsWhenBacklogGetIssueReturns5xx(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v2/issues/COMMUNITY-102" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	var commands []string
	app := App{
		Stdin:  strings.NewReader(""),
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Config: config.Config{
			BacklogSpaceURL:     server.URL,
			BacklogAPIKey:       "secret",
			BacklogDoneStatusID: 5,
		},
		Store: store.New(filepath.Join(t.TempDir(), "tree.json")),
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			commands = append(commands, command)
			if command == "git branch --show-current" {
				return "feature/member/backend/COMMUNITY-102", nil
			}
			t.Fatalf("unexpected command: %s", command)
			return "", nil
		}},
		Backlog: backlog.Client{SpaceURL: server.URL, APIKey: "secret", HTTPClient: server.Client()},
	}

	err := app.Run(context.Background(), []string{"pr"})
	if err == nil {
		t.Fatal("expected error when Backlog API returns 5xx")
	}
	if !strings.Contains(err.Error(), "502 Bad Gateway") {
		t.Fatalf("expected HTTP status in error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "/api/v2/issues/COMMUNITY-102") {
		t.Fatalf("expected endpoint in error, got %q", err.Error())
	}
	if len(commands) != 1 {
		t.Fatalf("expected only current branch lookup, got %d: %v", len(commands), commands)
	}
}

func TestPRBacklogUpdateFailureAfterPushAndCreate(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("apiKey") != "secret" {
			t.Fatalf("missing api key")
		}
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/issues/COMMUNITY-102":
			writeJSON(t, w, map[string]interface{}{
				"issueKey": "COMMUNITY-102",
				"summary":  "API利用画面を実装",
				"status": map[string]interface{}{
					"name": "対応中",
				},
			})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v2/issues/COMMUNITY-102":
			w.WriteHeader(http.StatusBadGateway)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	var commands []string
	out := &bytes.Buffer{}
	app := App{
		Stdin:  strings.NewReader("y\n"),
		Stdout: out,
		Stderr: &bytes.Buffer{},
		Config: config.Config{
			BacklogSpaceURL:     server.URL,
			BacklogAPIKey:       "secret",
			BacklogDoneStatusID: 5,
			DefaultBase:         "develop",
		},
		Store: store.New(filepath.Join(t.TempDir(), "tree.json")),
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			commands = append(commands, command)
			switch {
			case command == "git branch --show-current":
				return "feature/member/backend/COMMUNITY-102", nil
			case command == "git push -u origin feature/member/backend/COMMUNITY-102":
				return "", nil
			case strings.HasPrefix(command, "gh pr create "):
				return "https://github.com/example/repo/pull/1", nil
			default:
				t.Fatalf("unexpected command: %s", command)
				return "", nil
			}
		}},
		Backlog: backlog.Client{SpaceURL: server.URL, APIKey: "secret", HTTPClient: server.Client()},
	}

	err := app.Run(context.Background(), []string{"pr"})
	if err == nil {
		t.Fatal("expected error when Backlog status update fails")
	}
	if !strings.Contains(err.Error(), "git push and pull request were already created") {
		t.Fatalf("expected partial success hint in error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "Backlog status update failed") {
		t.Fatalf("expected Backlog failure context in error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "502 Bad Gateway") {
		t.Fatalf("expected HTTP status in error, got %q", err.Error())
	}
	if !strings.Contains(out.String(), "https://github.com/example/repo/pull/1") {
		t.Fatalf("expected PR URL in output, got %q", out.String())
	}
	if len(commands) != 3 {
		t.Fatalf("expected push and PR creation before Backlog failure, got %d: %v", len(commands), commands)
	}
}

func TestTodayContinuesWhenOneBacklogFetchFails(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/issues/COMMUNITY-103":
			if r.Method != http.MethodGet {
				t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
			}
			w.WriteHeader(http.StatusNotFound)
		case "/api/v2/issues/COMMUNITY-104":
			if r.Method != http.MethodGet {
				t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
			}
			fmt.Fprint(w, `{"summary":"成功課題","status":{"name":"未着手"}}`)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	treePath := filepath.Join(t.TempDir(), "tree.json")
	st := store.New(treePath)
	if err := st.Save(store.Tree{Records: []store.Record{
		{
			RepoRoot:     "/repo",
			ParentBranch: "feature/member/backend/COMMUNITY-102",
			ChildBranch:  "feature/member/backend/COMMUNITY-103",
			IssueKey:     "COMMUNITY-103",
		},
		{
			RepoRoot:     "/repo",
			ParentBranch: "feature/member/backend/COMMUNITY-102",
			ChildBranch:  "feature/member/backend/COMMUNITY-104",
			IssueKey:     "COMMUNITY-104",
		},
	}}); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	app := App{
		Stdin:  strings.NewReader(""),
		Stdout: out,
		Stderr: errOut,
		Config: config.Config{BacklogSpaceURL: server.URL, BacklogAPIKey: "secret"},
		Store:  st,
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			switch command {
			case "git branch --show-current":
				return "feature/member/backend/COMMUNITY-102", nil
			case "git rev-parse --show-toplevel":
				return "/repo", nil
			default:
				t.Fatalf("unexpected command: %s", command)
				return "", nil
			}
		}},
		Backlog: backlog.Client{SpaceURL: server.URL, APIKey: "secret", HTTPClient: server.Client()},
	}

	if err := app.Run(context.Background(), []string{"today"}); err != nil {
		t.Fatalf("expected success with partial backlog failure, got %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "COMMUNITY-103  feature/member/backend/COMMUNITY-103  -  -") {
		t.Fatalf("expected placeholder for failed fetch, got %q", output)
	}
	if !strings.Contains(output, "COMMUNITY-104  feature/member/backend/COMMUNITY-104  成功課題  未着手") {
		t.Fatalf("expected successful fetch output, got %q", output)
	}
	stderr := errOut.String()
	if !strings.Contains(stderr, "COMMUNITY-103") || !strings.Contains(stderr, "404 Not Found") {
		t.Fatalf("expected warning on stderr, got %q", stderr)
	}
}

func TestEpicStatusContinuesWhenOneBacklogFetchFails(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/issues/COMMUNITY-101":
			if r.Method != http.MethodGet {
				t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
			}
			w.WriteHeader(http.StatusInternalServerError)
		case "/api/v2/issues/COMMUNITY-102":
			if r.Method != http.MethodGet {
				t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
			}
			fmt.Fprint(w, `{"summary":"設計書を作成","status":{"name":"完了"}}`)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	treePath := filepath.Join(t.TempDir(), "tree.json")
	st := store.New(treePath)
	if err := st.Save(store.Tree{Records: []store.Record{
		{
			RepoRoot:     "/repo",
			ParentBranch: "develop",
			ChildBranch:  "feature/member/backend/COMMUNITY-101",
			IssueKey:     "COMMUNITY-101",
		},
		{
			RepoRoot:     "/repo",
			ParentBranch: "develop",
			ChildBranch:  "feature/member/backend/COMMUNITY-102",
			IssueKey:     "COMMUNITY-102",
		},
	}}); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	app := App{
		Stdin:  strings.NewReader(""),
		Stdout: out,
		Stderr: errOut,
		Config: config.Config{BacklogSpaceURL: server.URL, BacklogAPIKey: "secret"},
		Store:  st,
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			switch command {
			case "git rev-parse --show-toplevel":
				return "/repo", nil
			default:
				t.Fatalf("unexpected command: %s", command)
				return "", nil
			}
		}},
		Backlog: backlog.Client{SpaceURL: server.URL, APIKey: "secret", HTTPClient: server.Client()},
	}

	if err := app.Run(context.Background(), []string{"epic", "status", "COMMUNITY-100"}); err != nil {
		t.Fatalf("expected success with partial backlog failure, got %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "COMMUNITY-101  feature/member/backend/COMMUNITY-101  -  -") {
		t.Fatalf("expected placeholder for failed fetch, got %q", output)
	}
	if !strings.Contains(output, "COMMUNITY-102  feature/member/backend/COMMUNITY-102  設計書を作成  完了") {
		t.Fatalf("expected successful fetch output, got %q", output)
	}
	stderr := errOut.String()
	if !strings.Contains(stderr, "COMMUNITY-101") || !strings.Contains(stderr, "500 Internal Server Error") {
		t.Fatalf("expected warning on stderr, got %q", stderr)
	}
}

func TestTodayJSONContinuesWhenOneBacklogFetchFails(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/issues/COMMUNITY-103":
			w.WriteHeader(http.StatusNotFound)
		case "/api/v2/issues/COMMUNITY-104":
			fmt.Fprint(w, `{"summary":"成功課題","status":{"name":"未着手"}}`)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	treePath := filepath.Join(t.TempDir(), "tree.json")
	st := store.New(treePath)
	if err := st.Save(store.Tree{Records: []store.Record{
		{
			RepoRoot:     "/repo",
			ParentBranch: "feature/member/backend/COMMUNITY-102",
			ChildBranch:  "feature/member/backend/COMMUNITY-103",
			IssueKey:     "COMMUNITY-103",
		},
		{
			RepoRoot:     "/repo",
			ParentBranch: "feature/member/backend/COMMUNITY-102",
			ChildBranch:  "feature/member/backend/COMMUNITY-104",
			IssueKey:     "COMMUNITY-104",
		},
	}}); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	app := App{
		Stdin:  strings.NewReader(""),
		Stdout: out,
		Stderr: errOut,
		Config: config.Config{BacklogSpaceURL: server.URL, BacklogAPIKey: "secret"},
		Store:  st,
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			switch command {
			case "git branch --show-current":
				return "feature/member/backend/COMMUNITY-102", nil
			case "git rev-parse --show-toplevel":
				return "/repo", nil
			default:
				t.Fatalf("unexpected command: %s", command)
				return "", nil
			}
		}},
		Backlog: backlog.Client{SpaceURL: server.URL, APIKey: "secret", HTTPClient: server.Client()},
	}

	if err := app.Run(context.Background(), []string{"today", "--json"}); err != nil {
		t.Fatalf("expected success with partial backlog failure, got %v", err)
	}

	var payload todayJSONOutput
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(payload.Children))
	}
	if payload.Children[0].IssueKey != "COMMUNITY-103" || payload.Children[0].Title != "" || payload.Children[0].Status != "" {
		t.Fatalf("expected empty title/status for failed fetch, got %+v", payload.Children[0])
	}
	if payload.Children[1].IssueKey != "COMMUNITY-104" || payload.Children[1].Title != "成功課題" || payload.Children[1].Status != "未着手" {
		t.Fatalf("unexpected successful child: %+v", payload.Children[1])
	}
	if !strings.Contains(errOut.String(), "COMMUNITY-103") {
		t.Fatalf("expected warning on stderr, got %q", errOut.String())
	}
}

func TestEpicStatusUsesCurrentBranchWhenEpicKeyIsMissing(t *testing.T) {
	t.Parallel()

	treePath := filepath.Join(t.TempDir(), "tree.json")
	st := store.New(treePath)
	if err := st.Save(store.Tree{Records: []store.Record{{
		RepoRoot:     "/repo",
		ParentBranch: "develop",
		ChildBranch:  "feature/member/backend/COMMUNITY-101",
		IssueKey:     "COMMUNITY-101",
	}}}); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	app := App{
		Stdin:    strings.NewReader(""),
		Stdout:   out,
		Stderr:   &bytes.Buffer{},
		Store:    st,
		loadDeps: false,
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			switch command {
			case "git rev-parse --show-toplevel":
				return "/repo", nil
			case "git branch --show-current":
				return "feature/member/backend/COMMUNITY-100", nil
			default:
				t.Fatalf("unexpected command: %s", command)
				return "", nil
			}
		}},
	}

	if err := app.Run(context.Background(), []string{"epic", "status"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Epic COMMUNITY-100") {
		t.Fatalf("unexpected output: %s", out.String())
	}
	if !strings.Contains(out.String(), "COMMUNITY-101") {
		t.Fatalf("expected epic child in output: %s", out.String())
	}
}

func TestTodayJSONOutputsChildren(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v2/issues/COMMUNITY-103" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		writeJSON(t, w, map[string]interface{}{
			"issueKey": "COMMUNITY-103",
			"summary":  "テストを書く",
			"status": map[string]interface{}{
				"name": "未着手",
			},
		})
	}))
	defer server.Close()

	treePath := filepath.Join(t.TempDir(), "tree.json")
	st := store.New(treePath)
	createdAt := time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)
	if err := st.Save(store.Tree{Records: []store.Record{{
		RepoRoot:     "/repo",
		ParentBranch: "feature/member/backend/COMMUNITY-102",
		ChildBranch:  "feature/member/backend/COMMUNITY-103",
		IssueKey:     "COMMUNITY-103",
		CreatedAt:    createdAt,
	}}}); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	app := App{
		Stdin:  strings.NewReader(""),
		Stdout: out,
		Stderr: &bytes.Buffer{},
		Config: config.Config{BacklogSpaceURL: server.URL, BacklogAPIKey: "secret"},
		Store:  st,
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			switch command {
			case "git branch --show-current":
				return "feature/member/backend/COMMUNITY-102", nil
			case "git rev-parse --show-toplevel":
				return "/repo", nil
			default:
				t.Fatalf("unexpected command: %s", command)
				return "", nil
			}
		}},
		Backlog: backlog.Client{SpaceURL: server.URL, APIKey: "secret", HTTPClient: server.Client()},
	}

	if err := app.Run(context.Background(), []string{"today", "--json"}); err != nil {
		t.Fatal(err)
	}

	var payload todayJSONOutput
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, out.String())
	}
	if payload.CurrentBranch != "feature/member/backend/COMMUNITY-102" {
		t.Fatalf("unexpected current branch: %s", payload.CurrentBranch)
	}
	if len(payload.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(payload.Children))
	}
	child := payload.Children[0]
	if child.IssueKey != "COMMUNITY-103" || child.Title != "テストを書く" || child.Status != "未着手" {
		t.Fatalf("unexpected child record: %+v", child)
	}
	if child.ChildBranch != "feature/member/backend/COMMUNITY-103" {
		t.Fatalf("unexpected child branch: %s", child.ChildBranch)
	}
}

func TestTodayJSONNoBacklogSkipsBacklogAPI(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Backlog API should not be called with --no-backlog")
	}))
	defer server.Close()

	treePath := filepath.Join(t.TempDir(), "tree.json")
	st := store.New(treePath)
	if err := st.Save(store.Tree{Records: []store.Record{{
		RepoRoot:     "/repo",
		ParentBranch: "feature/member/backend/COMMUNITY-102",
		ChildBranch:  "feature/member/backend/COMMUNITY-103",
		IssueKey:     "COMMUNITY-103",
	}}}); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	app := App{
		Stdin:  strings.NewReader(""),
		Stdout: out,
		Stderr: &bytes.Buffer{},
		Config: config.Config{BacklogSpaceURL: server.URL, BacklogAPIKey: "secret"},
		Store:  st,
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			switch command {
			case "git branch --show-current":
				return "feature/member/backend/COMMUNITY-102", nil
			case "git rev-parse --show-toplevel":
				return "/repo", nil
			default:
				t.Fatalf("unexpected command: %s", command)
				return "", nil
			}
		}},
		Backlog: backlog.Client{SpaceURL: server.URL, APIKey: "secret", HTTPClient: server.Client()},
	}

	if err := app.Run(context.Background(), []string{"today", "--json", "--no-backlog"}); err != nil {
		t.Fatal(err)
	}

	var payload todayJSONOutput
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, out.String())
	}
	if len(payload.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(payload.Children))
	}
	child := payload.Children[0]
	if child.Title != "" || child.Status != "" {
		t.Fatalf("expected empty title/status without backlog, got %+v", child)
	}
}

func TestEpicStatusJSONOutputsRecords(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v2/issues/COMMUNITY-101" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		writeJSON(t, w, map[string]interface{}{
			"issueKey": "COMMUNITY-101",
			"summary":  "設計書を作成",
			"status": map[string]interface{}{
				"name": "完了",
			},
		})
	}))
	defer server.Close()

	treePath := filepath.Join(t.TempDir(), "tree.json")
	st := store.New(treePath)
	if err := st.Save(store.Tree{Records: []store.Record{{
		RepoRoot:     "/repo",
		ParentBranch: "develop",
		ChildBranch:  "feature/member/backend/COMMUNITY-101",
		IssueKey:     "COMMUNITY-101",
	}}}); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	app := App{
		Stdin:  strings.NewReader(""),
		Stdout: out,
		Stderr: &bytes.Buffer{},
		Config: config.Config{BacklogSpaceURL: server.URL, BacklogAPIKey: "secret"},
		Store:  st,
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			if command == "git rev-parse --show-toplevel" {
				return "/repo", nil
			}
			t.Fatalf("unexpected command: %s", command)
			return "", nil
		}},
		Backlog: backlog.Client{SpaceURL: server.URL, APIKey: "secret", HTTPClient: server.Client()},
	}

	if err := app.Run(context.Background(), []string{"epic", "status", "--json", "COMMUNITY-100"}); err != nil {
		t.Fatal(err)
	}

	var payload epicJSONOutput
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, out.String())
	}
	if payload.EpicKey != "COMMUNITY-100" {
		t.Fatalf("unexpected epic key: %s", payload.EpicKey)
	}
	if len(payload.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(payload.Records))
	}
	record := payload.Records[0]
	if record.IssueKey != "COMMUNITY-101" || record.Title != "設計書を作成" || record.Status != "完了" {
		t.Fatalf("unexpected record: %+v", record)
	}
}

func TestEpicStatusNoBacklogSkipsBacklogAPI(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Backlog API should not be called with --no-backlog")
	}))
	defer server.Close()

	treePath := filepath.Join(t.TempDir(), "tree.json")
	st := store.New(treePath)
	if err := st.Save(store.Tree{Records: []store.Record{{
		RepoRoot:     "/repo",
		ParentBranch: "develop",
		ChildBranch:  "feature/member/backend/COMMUNITY-101",
		IssueKey:     "COMMUNITY-101",
	}}}); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	app := App{
		Stdin:  strings.NewReader(""),
		Stdout: out,
		Stderr: &bytes.Buffer{},
		Config: config.Config{BacklogSpaceURL: server.URL, BacklogAPIKey: "secret"},
		Store:  st,
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			if command == "git rev-parse --show-toplevel" {
				return "/repo", nil
			}
			t.Fatalf("unexpected command: %s", command)
			return "", nil
		}},
		Backlog: backlog.Client{SpaceURL: server.URL, APIKey: "secret", HTTPClient: server.Client()},
	}

	if err := app.Run(context.Background(), []string{"epic", "status", "--no-backlog", "COMMUNITY-100"}); err != nil {
		t.Fatal(err)
	}
	output := out.String()
	if !strings.Contains(output, "COMMUNITY-101  feature/member/backend/COMMUNITY-101  -  -") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestEpicStatusJSONNoBacklogSkipsBacklogAPI(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Backlog API should not be called with --no-backlog")
	}))
	defer server.Close()

	treePath := filepath.Join(t.TempDir(), "tree.json")
	st := store.New(treePath)
	if err := st.Save(store.Tree{Records: []store.Record{{
		RepoRoot:     "/repo",
		ParentBranch: "develop",
		ChildBranch:  "feature/member/backend/COMMUNITY-101",
		IssueKey:     "COMMUNITY-101",
	}}}); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	app := App{
		Stdin:  strings.NewReader(""),
		Stdout: out,
		Stderr: &bytes.Buffer{},
		Config: config.Config{BacklogSpaceURL: server.URL, BacklogAPIKey: "secret"},
		Store:  st,
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			if command == "git rev-parse --show-toplevel" {
				return "/repo", nil
			}
			t.Fatalf("unexpected command: %s", command)
			return "", nil
		}},
		Backlog: backlog.Client{SpaceURL: server.URL, APIKey: "secret", HTTPClient: server.Client()},
	}

	if err := app.Run(context.Background(), []string{"epic", "status", "--json", "--no-backlog", "COMMUNITY-100"}); err != nil {
		t.Fatal(err)
	}

	var payload epicJSONOutput
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, out.String())
	}
	if len(payload.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(payload.Records))
	}
	record := payload.Records[0]
	if record.Title != "" || record.Status != "" {
		t.Fatalf("expected empty title/status without backlog, got %+v", record)
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, value interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
