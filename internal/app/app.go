package app

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"git-cli/internal/backlog"
	"git-cli/internal/config"
	gitcmd "git-cli/internal/git"
	"git-cli/internal/store"
)

type App struct {
	Dir      string
	Stdin    io.Reader
	Stdout   io.Writer
	Stderr   io.Writer
	Config   config.Config
	Store    *store.Store
	Git      gitcmd.Client
	Backlog  backlog.Client
	loadDeps bool
}

func New(dir string, stdin io.Reader, stdout io.Writer, stderr io.Writer) App {
	return App{Dir: dir, Stdin: stdin, Stdout: stdout, Stderr: stderr, loadDeps: true}
}

func (a App) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		a.printHelp("")
		return nil
	}

	if isHelpRequest(args) {
		a.printHelp(helpCommandName(args))
		return nil
	}

	if a.loadDeps {
		loaded, err := a.withDeps()
		if err != nil {
			return err
		}
		a = loaded
	}

	switch args[0] {
	case "work":
		if isSubcommandHelp(args[1:]) {
			a.printHelp("work")
			return nil
		}
		return a.runWork(ctx, args[1:])
	case "pr":
		if isSubcommandHelp(args[1:]) {
			a.printHelp("pr")
			return nil
		}
		return a.runPR(ctx, args[1:])
	case "today":
		if isSubcommandHelp(args[1:]) {
			a.printHelp("today")
			return nil
		}
		return a.runToday(ctx, args[1:])
	case "epic":
		if isSubcommandHelp(args[1:]) {
			a.printHelp("epic")
			return nil
		}
		return a.runEpic(ctx, args[1:])
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func (a App) withDeps() (App, error) {
	cfg, err := config.Load()
	if err != nil {
		return App{}, err
	}
	st, err := store.NewDefault()
	if err != nil {
		return App{}, err
	}
	a.Config = cfg
	a.Store = st
	a.Git = gitcmd.New(a.Dir)
	a.Backlog = backlog.Client{SpaceURL: cfg.BacklogSpaceURL, APIKey: cfg.BacklogAPIKey}
	a.loadDeps = false
	return a, nil
}

func (a App) runWork(ctx context.Context, args []string) error {
	if len(args) != 1 {
		return errors.New("usage: gitwork work <issue-key>")
	}

	issueKey := strings.ToUpper(args[0])
	parentBranch, err := a.Git.CurrentBranch(ctx)
	if err != nil {
		return err
	}
	repoRoot, err := a.Git.RepoRoot(ctx)
	if err != nil {
		return err
	}

	childBranch := branchName(a.Config.BranchPattern, issueKey)
	if err := a.Git.CreateBranch(ctx, childBranch); err != nil {
		return err
	}

	record := store.Record{
		RepoRoot:     repoRoot,
		ParentBranch: parentBranch,
		ChildBranch:  childBranch,
		IssueKey:     issueKey,
		CreatedAt:    time.Now(),
	}
	if err := a.Store.Add(record); err != nil {
		return err
	}

	fmt.Fprintf(a.Stdout, "created %s from %s\n", childBranch, parentBranch)
	return nil
}

func (a App) runPR(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("pr", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	dryRun := fs.Bool("dry-run", false, "print the pull request without creating it")
	yes := fs.Bool("yes", false, "skip confirmation")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if err := a.Config.ValidateDoneStatus(); err != nil {
		return err
	}

	currentBranch, err := a.Git.CurrentBranch(ctx)
	if err != nil {
		return err
	}
	issueKey, err := issueKeyFromBranch(currentBranch)
	if err != nil {
		return err
	}

	issue, err := a.Backlog.GetIssue(ctx, issueKey)
	if err != nil {
		return err
	}

	body := prBody(issue)
	title := issue.Summary
	base := a.Config.DefaultBase
	if base == "" {
		base = "develop"
	}

	fmt.Fprintf(a.Stdout, "title: %s\nbase: %s\n\n%s\n", title, base, body)
	if !*yes && !*dryRun {
		ok, err := a.confirm("Create pull request?")
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(a.Stdout, "cancelled")
			return nil
		}
	}

	if *dryRun {
		return nil
	}

	if err := a.Git.PushCurrentBranch(ctx, currentBranch); err != nil {
		return err
	}
	output, err := a.Git.CreatePullRequest(ctx, a.Config.GitHubRepo, title, body, base, false)
	if err != nil {
		return err
	}
	fmt.Fprintln(a.Stdout, output)

	if err := a.Backlog.UpdateIssueStatus(ctx, issueKey, a.Config.BacklogDoneStatusID); err != nil {
		return err
	}
	fmt.Fprintf(a.Stdout, "updated Backlog status: %s\n", issueKey)
	return nil
}

func (a App) runToday(ctx context.Context, args []string) error {
	if len(args) != 0 {
		return errors.New("usage: gitwork today")
	}
	return a.printRecordsForCurrentBranch(ctx)
}

func (a App) runEpic(ctx context.Context, args []string) error {
	if len(args) != 2 || args[0] != "status" {
		return errors.New("usage: gitwork epic status <epic-key>")
	}

	repoRoot, err := a.Git.RepoRoot(ctx)
	if err != nil {
		return err
	}
	tree, err := a.Store.Load()
	if err != nil {
		return err
	}

	fmt.Fprintf(a.Stdout, "Epic %s\n\n", strings.ToUpper(args[1]))
	return a.printRecords(ctx, tree.ForEpic(repoRoot, args[1]))
}

func (a App) printRecordsForCurrentBranch(ctx context.Context) error {
	currentBranch, err := a.Git.CurrentBranch(ctx)
	if err != nil {
		return err
	}
	repoRoot, err := a.Git.RepoRoot(ctx)
	if err != nil {
		return err
	}
	tree, err := a.Store.Load()
	if err != nil {
		return err
	}

	fmt.Fprintf(a.Stdout, "Current branch\n%s\n\nChildren\n", currentBranch)
	return a.printRecords(ctx, tree.Children(repoRoot, currentBranch))
}

func (a App) printRecords(ctx context.Context, records []store.Record) error {
	if len(records) == 0 {
		fmt.Fprintln(a.Stdout, "(none)")
		return nil
	}

	for _, record := range records {
		title := "-"
		status := "-"
		if a.Config.ValidateBacklog() == nil {
			issue, err := a.Backlog.GetIssue(ctx, record.IssueKey)
			if err != nil {
				return err
			}
			title = issue.Summary
			status = issue.Status
		}
		fmt.Fprintf(a.Stdout, "- %s  %s  %s\n", record.IssueKey, title, status)
	}
	return nil
}

func (a App) confirm(question string) (bool, error) {
	fmt.Fprintf(a.Stdout, "%s [y/N]: ", question)
	reader := bufio.NewReader(a.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes", nil
}

func branchName(pattern string, issueKey string) string {
	if pattern == "" {
		pattern = "feature/{issueKey}"
	}
	return strings.ReplaceAll(pattern, "{issueKey}", strings.ToLower(issueKey))
}

func issueKeyFromBranch(branch string) (string, error) {
	re := regexp.MustCompile(`[A-Za-z]+-\d+`)
	match := re.FindString(branch)
	if match == "" {
		return "", fmt.Errorf("issue key not found in branch: %s", branch)
	}
	return strings.ToUpper(match), nil
}

func prBody(issue backlog.Issue) string {
	return fmt.Sprintf(`## Backlog

%s

## 概要

%s

## 確認項目

- [ ] 動作確認
`, issue.URL, issue.Summary)
}
