package app

import (
	"testing"
)

func TestWorkBranchName(t *testing.T) {
	t.Parallel()

	got := workBranchName("member", "backend", "community-100")
	want := "feature/member/backend/COMMUNITY-100"
	if got != want {
		t.Fatalf("unexpected branch name: got %q want %q", got, want)
	}
}

func TestNormalizeWorkTeam(t *testing.T) {
	t.Parallel()

	got, err := normalizeWorkTeam("2")
	if err != nil {
		t.Fatal(err)
	}
	if got != "admin" {
		t.Fatalf("unexpected team: %s", got)
	}
}

func TestNormalizeWorkLayer(t *testing.T) {
	t.Parallel()

	got, err := normalizeWorkLayer("frontend")
	if err != nil {
		t.Fatal(err)
	}
	if got != "frontend" {
		t.Fatalf("unexpected layer: %s", got)
	}
}

func TestNormalizeWorkTeamRejectsUnknownValue(t *testing.T) {
	t.Parallel()

	if _, err := normalizeWorkTeam("ops"); err == nil {
		t.Fatal("expected error for unknown team")
	}
}
