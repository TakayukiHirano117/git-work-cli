package app

import "testing"

func TestParseWorkArgs(t *testing.T) {
	t.Parallel()

	issueKey, team, layer, err := parseWorkArgs([]string{"COMMUNITY-102", "--team", "member", "--layer", "backend"})
	if err != nil {
		t.Fatal(err)
	}
	if issueKey != "COMMUNITY-102" || team != "member" || layer != "backend" {
		t.Fatalf("unexpected parse result: %q %q %q", issueKey, team, layer)
	}
}

func TestParseWorkArgsRequiresIssueKey(t *testing.T) {
	t.Parallel()

	if _, _, _, err := parseWorkArgs([]string{"--team", "member"}); err == nil {
		t.Fatal("expected error when issue key is missing")
	}
}
