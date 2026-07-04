package app

import (
	"fmt"
	"regexp"
	"strings"
)

var issueKeyPattern = regexp.MustCompile(`^[A-Za-z]+-\d+$`)

func parseIssueKey(input string) (string, error) {
	key := strings.ToUpper(strings.TrimSpace(input))
	if !issueKeyPattern.MatchString(key) {
		return "", fmt.Errorf("invalid issue key %q (expected format, e.g. COMMUNITY-102)", input)
	}
	return key, nil
}
