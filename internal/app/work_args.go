package app

import "errors"

func parseWorkArgs(args []string) (issueKey string, team string, layer string, err error) {
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--team":
			team, index, err = readWorkFlagValue(args, index, "team")
			if err != nil {
				return "", "", "", err
			}
		case "--layer":
			layer, index, err = readWorkFlagValue(args, index, "layer")
			if err != nil {
				return "", "", "", err
			}
		default:
			if issueKey != "" {
				return "", "", "", errors.New("usage: gitwork work <issue-key> [--team member|admin|agency] [--layer frontend|backend]")
			}
			issueKey = args[index]
		}
	}

	if issueKey == "" {
		return "", "", "", errors.New("usage: gitwork work <issue-key> [--team member|admin|agency] [--layer frontend|backend]")
	}

	return issueKey, team, layer, nil
}

func readWorkFlagValue(args []string, index int, flagName string) (string, int, error) {
	if index+1 >= len(args) {
		return "", index, errors.New("usage: gitwork work <issue-key> [--team member|admin|agency] [--layer frontend|backend]")
	}
	return args[index+1], index + 1, nil
}
