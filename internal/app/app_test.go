package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	if !strings.Contains(output, "COMMUNITY-103  テストを書く  未着手") {
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
	if !strings.Contains(output, "COMMUNITY-103  -  -") {
		t.Fatalf("unexpected output: %s", output)
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
