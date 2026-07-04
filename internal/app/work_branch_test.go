package app

import (
	"bufio"
	"bytes"
	"io"
	"strings"
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

func TestRequireWorkFlagsForNonInteractiveRejectsMissingFlags(t *testing.T) {
	t.Parallel()

	err := requireWorkFlagsForNonInteractive("", "", strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for missing team and layer")
	}
	if !strings.Contains(err.Error(), "--team") || !strings.Contains(err.Error(), "--layer") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequireWorkFlagsForNonInteractiveAllowsFlags(t *testing.T) {
	t.Parallel()

	if err := requireWorkFlagsForNonInteractive("member", "backend", strings.NewReader("")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequireWorkFlagsForNonInteractiveAllowsTTYWithoutFlags(t *testing.T) {
	t.Parallel()

	old := stdinIsTTY
	stdinIsTTY = func(io.Reader) bool { return true }
	t.Cleanup(func() { stdinIsTTY = old })

	if err := requireWorkFlagsForNonInteractive("", "", strings.NewReader("")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSelectWorkOptionAcceptsChoice(t *testing.T) {
	t.Parallel()

	reader := bufio.NewReader(strings.NewReader("2\n"))
	stdout := &bytes.Buffer{}
	got, err := selectWorkOption(reader, stdout, "Select team", workTeamOptions, normalizeWorkTeam)
	if err != nil {
		t.Fatal(err)
	}
	if got != "admin" {
		t.Fatalf("unexpected team: %s", got)
	}
}
