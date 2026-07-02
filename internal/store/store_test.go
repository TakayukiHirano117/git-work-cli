package store

import (
	"path/filepath"
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
	if err := st.Add(record); err == nil {
		t.Fatal("expected duplicate branch error")
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
