package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAddRejectsDuplicateChildBranchInSameRepo(t *testing.T) {
	t.Parallel()

	st := New(filepath.Join(t.TempDir(), "tree.json"))
	record := Record{
		RepoRoot:     "/repo",
		ParentBranch: "develop",
		ChildBranch:  "feature/community-101",
		IssueKey:     "COMMUNITY-101",
		CreatedAt:    time.Now(),
	}

	if err := st.Add(record); err != nil {
		t.Fatal(err)
	}
	err := st.Add(record)
	if err == nil {
		t.Fatal("expected duplicate branch error")
	}
	if !strings.Contains(err.Error(), "feature/community-101") {
		t.Fatalf("expected child branch in error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "parent: develop") {
		t.Fatalf("expected parent branch in error, got %q", err.Error())
	}
}

func TestAddAllowsSameChildBranchInDifferentRepos(t *testing.T) {
	t.Parallel()

	st := New(filepath.Join(t.TempDir(), "tree.json"))
	first := Record{
		RepoRoot:     "/repo/a",
		ParentBranch: "develop",
		ChildBranch:  "feature/community-101",
		IssueKey:     "COMMUNITY-101",
	}
	second := first
	second.RepoRoot = "/repo/b"

	if err := st.Add(first); err != nil {
		t.Fatal(err)
	}
	if err := st.Add(second); err != nil {
		t.Fatal(err)
	}

	tree, err := st.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.Records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(tree.Records))
	}
}

func TestLoadInvalidTreeJSONReturnsRecoveryHint(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "tree.json")
	if err := os.WriteFile(path, []byte("{not-json"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := New(path).Load()
	if err == nil {
		t.Fatal("expected error for invalid tree.json")
	}
	for _, want := range []string{
		path,
		"invalid tree.json",
		"fix the JSON or remove the file",
		"totonou config path",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected error to contain %q, got %q", want, err.Error())
		}
	}
}

func TestAddRejectsInvalidTreeJSON(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "tree.json")
	if err := os.WriteFile(path, []byte("[]"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := New(path).Add(Record{
		RepoRoot:     "/repo",
		ParentBranch: "develop",
		ChildBranch:  "feature/community-101",
		IssueKey:     "COMMUNITY-101",
	})
	if err == nil {
		t.Fatal("expected error when adding to invalid tree.json")
	}
	if !strings.Contains(err.Error(), "invalid tree.json") {
		t.Fatalf("expected invalid tree.json error, got %q", err.Error())
	}
}

func TestTreeFiltersRecords(t *testing.T) {
	t.Parallel()

	tree := Tree{Records: []Record{
		{
			RepoRoot:     "/repo",
			ParentBranch: "develop",
			ChildBranch:  "feature/community-101",
			IssueKey:     "COMMUNITY-101",
		},
		{
			RepoRoot:     "/repo",
			ParentBranch: "feature/community-101",
			ChildBranch:  "feature/community-102",
			IssueKey:     "COMMUNITY-102",
		},
		{
			RepoRoot:     "/other",
			ParentBranch: "develop",
			ChildBranch:  "feature/community-103",
			IssueKey:     "COMMUNITY-103",
		},
		{
			RepoRoot:     "/repo",
			ParentBranch: "develop",
			ChildBranch:  "feature/other-201",
			IssueKey:     "OTHER-201",
		},
	}}

	children := tree.Children("/repo", "develop")
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}

	epicRecords := tree.ForEpic("/repo", "community-999")
	if len(epicRecords) != 2 {
		t.Fatalf("expected 2 epic records, got %d", len(epicRecords))
	}
}
