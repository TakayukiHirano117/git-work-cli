package app

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var workTeamOptions = []string{"member", "admin", "agency"}
var workLayerOptions = []string{"frontend", "backend"}

func workBranchName(team string, layer string, issueKey string) string {
	return fmt.Sprintf("feature/%s/%s/%s", team, layer, strings.ToUpper(issueKey))
}

func printWorkSuccess(stdout io.Writer, childBranch, parentBranch string) {
	fmt.Fprintf(stdout, "created %s from %s\n", childBranch, parentBranch)
	fmt.Fprintln(stdout, "next:")
	fmt.Fprintln(stdout, "  totonou pr --dry-run")
}

var stdinIsTTY = defaultStdinIsTTY

func defaultStdinIsTTY(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func requireWorkFlagsForNonInteractive(teamFlag, layerFlag string, stdin io.Reader) error {
	if teamFlag != "" && layerFlag != "" {
		return nil
	}
	if stdinIsTTY(stdin) {
		return nil
	}

	missing := make([]string, 0, 2)
	if teamFlag == "" {
		missing = append(missing, "--team")
	}
	if layerFlag == "" {
		missing = append(missing, "--layer")
	}
	return fmt.Errorf(
		"non-interactive stdin: specify %s (e.g. --team member --layer backend)",
		strings.Join(missing, " and "),
	)
}

func resolveWorkTeamChoice(teamFlag string, reader *bufio.Reader, stdout io.Writer) (string, error) {
	if teamFlag != "" {
		return normalizeWorkTeam(teamFlag)
	}
	return selectWorkOption(reader, stdout, "Select team", workTeamOptions, normalizeWorkTeam)
}

func resolveWorkLayerChoice(layerFlag string, reader *bufio.Reader, stdout io.Writer) (string, error) {
	if layerFlag != "" {
		return normalizeWorkLayer(layerFlag)
	}
	return selectWorkOption(reader, stdout, "Select layer", workLayerOptions, normalizeWorkLayer)
}

func selectWorkOption(
	reader *bufio.Reader,
	stdout io.Writer,
	label string,
	options []string,
	normalize func(string) (string, error),
) (string, error) {
	for index, option := range options {
		fmt.Fprintf(stdout, "%d) %s\n", index+1, option)
	}

	for {
		fmt.Fprintf(stdout, "%s: ", label)
		input, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		}

		choice := strings.TrimSpace(input)
		if choice == "" {
			continue
		}

		normalized, err := normalize(choice)
		if err != nil {
			fmt.Fprintln(stdout, err.Error())
			continue
		}
		return normalized, nil
	}
}

func normalizeWorkTeam(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "1" {
		return workTeamOptions[0], nil
	}
	if normalized == "2" {
		return workTeamOptions[1], nil
	}
	if normalized == "3" {
		return workTeamOptions[2], nil
	}

	for _, option := range workTeamOptions {
		if normalized == option {
			return option, nil
		}
	}

	return "", fmt.Errorf("invalid team: choose member, admin, or agency")
}

func normalizeWorkLayer(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "1" {
		return workLayerOptions[0], nil
	}
	if normalized == "2" {
		return workLayerOptions[1], nil
	}

	for _, option := range workLayerOptions {
		if normalized == option {
			return option, nil
		}
	}

	return "", fmt.Errorf("invalid layer: choose frontend or backend")
}
