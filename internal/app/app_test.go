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
		Config: config.Config{BranchPattern: "feature/member/backend/{issueKey}"},
		Store:  st,
		Git: gitcmd.Client{Run: func(_ context.Context, _ string, name string, args ...string) (string, error) {
			command := name + " " + strings.Join(args, " ")
			commands = append(commands, command)
			switch command {
			case "git branch --show-current":
				return "feature/member/backend/community-101", nil
			case "git rev-parse --show-toplevel":
				return "/repo", nil
			case "git switch -c feature/member/backend/community-102":
				return "", nil
			default:
				t.Fatalf("unexpected command: %s", command)
				return "", nil
			}
		}},
	}

	if err := app.Run(context.Background(), []string{"work", "community-102"}); err != nil {
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
	if record.ParentBranch != "feature/member/backend/community-101" {
		t.Fatalf("unexpected parent branch: %s", record.ParentBranch)
	}
	if record.ChildBranch != "feature/member/backend/community-102" {
		t.Fatalf("unexpected child branch: %s", record.ChildBranch)
	}
	if len(commands) != 3 {
		t.Fatalf("expected 3 git commands, got %d", len(commands))
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
				return "feature/member/backend/community-102", nil
			case command == "git push -u origin feature/member/backend/community-102":
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
		ParentBranch: "feature/member/backend/community-102",
		ChildBranch:  "feature/member/backend/community-103",
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
				return "feature/member/backend/community-102", nil
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

func TestIssueKeyFromBranch(t *testing.T) {
	t.Parallel()

	issueKey, err := issueKeyFromBranch("feature/member/backend/community-102")
	if err != nil {
		t.Fatal(err)
	}
	if issueKey != "COMMUNITY-102" {
		t.Fatalf("unexpected issue key: %s", issueKey)
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
