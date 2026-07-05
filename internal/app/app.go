package app

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
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

	if args[0] == "config" {
		if isSubcommandHelp(args[1:]) {
			a.printHelp("config")
			return nil
		}
		return a.runConfig(args[1:])
	}

	if args[0] == "init" {
		if isSubcommandHelp(args[1:]) {
			a.printHelp("init")
			return nil
		}
		return a.runInit(args[1:])
	}

	if !isKnownCommand(args[0]) {
		return fmt.Errorf("unknown command: %s", args[0])
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
	case "doctor":
		if isSubcommandHelp(args[1:]) {
			a.printHelp("doctor")
			return nil
		}
		return a.runDoctor(ctx, args[1:])
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
	issueKeyArg, teamFlag, layerFlag, err := parseWorkArgs(args)
	if err != nil {
		return err
	}

	issueKey, err := parseIssueKey(issueKeyArg)
	if err != nil {
		return err
	}
	parentBranch, err := a.Git.CurrentBranch(ctx)
	if err != nil {
		return err
	}
	repoRoot, err := a.Git.RepoRoot(ctx)
	if err != nil {
		return err
	}

	if err := requireWorkFlagsForNonInteractive(teamFlag, layerFlag, a.Stdin); err != nil {
		return err
	}

	reader := bufio.NewReader(a.Stdin)
	team, err := resolveWorkTeamChoice(teamFlag, reader, a.Stdout)
	if err != nil {
		return err
	}
	layer, err := resolveWorkLayerChoice(layerFlag, reader, a.Stdout)
	if err != nil {
		return err
	}

	childBranch := workBranchName(team, layer, issueKey)
	tree, err := a.Store.Load()
	if err != nil {
		return err
	}
	if existing, ok := tree.FindChildBranch(repoRoot, childBranch); ok {
		return store.DuplicateBranchError(childBranch, existing.ParentBranch)
	}

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
	if err := a.Config.ValidateGitHub(); err != nil {
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

	if *dryRun {
		printPRDryRun(a.Stdout, currentBranch, title, base, body, issueKey, a.Config)
		return nil
	}

	fmt.Fprintf(a.Stdout, "title: %s\nbase: %s\n\n%s\n", title, base, body)
	if !*yes {
		ok, err := a.confirm("Create pull request?")
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(a.Stdout, "cancelled")
			return nil
		}
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
		return fmt.Errorf("git push and pull request were already created; Backlog status update failed: %w", err)
	}
	fmt.Fprintf(a.Stdout, "updated Backlog status: %s\n", issueKey)
	return nil
}

func (a App) runDoctor(ctx context.Context, args []string) error {
	if len(args) != 0 {
		return errors.New("usage: gitwork doctor")
	}

	failed := 0

	repoRoot, err := a.Git.RepoRoot(ctx)
	if err != nil {
		fmt.Fprintf(a.Stdout, "git repository: not ok (%v)\n", err)
		failed++
	} else {
		fmt.Fprintf(a.Stdout, "git repository: ok (%s)\n", repoRoot)
	}

	if err := a.Git.GHAuthStatus(ctx); err != nil {
		fmt.Fprintf(a.Stdout, "gh auth: not ok (%v)\n", err)
		failed++
	} else {
		fmt.Fprintln(a.Stdout, "gh auth: ok")
	}

	if err := a.Config.ValidateDoneStatus(); err != nil {
		fmt.Fprintf(a.Stdout, "backlog config: not ok (%v)\n", err)
		failed++
	} else {
		fmt.Fprintln(a.Stdout, "backlog config: ok")
	}

	if err := a.Config.ValidateGitHub(); err != nil {
		fmt.Fprintf(a.Stdout, "github config: not ok (%v)\n", err)
		failed++
	} else {
		fmt.Fprintln(a.Stdout, "github config: ok")
	}

	if failed > 0 {
		return fmt.Errorf("%d check(s) failed", failed)
	}
	return nil
}

func (a App) runInit(args []string) error {
	if len(args) != 0 {
		return errors.New("usage: gitwork init")
	}

	envPath, err := config.DefaultEnvPath()
	if err != nil {
		return err
	}
	treePath, err := config.DefaultTreePath()
	if err != nil {
		return err
	}

	fmt.Fprintln(a.Stdout, "gitwork の設定ファイルは次の場所に保存されます:")
	fmt.Fprintf(a.Stdout, "  config: %s\n", envPath)
	fmt.Fprintf(a.Stdout, "  tree:   %s\n\n", treePath)

	if _, err := os.Stat(envPath); err == nil {
		fmt.Fprintf(a.Stdout, ".env は既に存在します: %s\n", envPath)
		fmt.Fprintln(a.Stdout, "編集後は gitwork doctor で設定を確認できます。")
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	ok, err := a.confirm("Create .env template?")
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintln(a.Stdout, "cancelled")
		return nil
	}

	if err := config.WriteEnvTemplate(envPath); err != nil {
		return err
	}

	fmt.Fprintf(a.Stdout, "created %s\n", envPath)
	fmt.Fprintln(a.Stdout, "値を編集したあと、gitwork doctor で設定を確認できます。")
	return nil
}

func (a App) runConfig(args []string) error {
	if len(args) != 1 || args[0] != "path" {
		return errors.New("usage: gitwork config path")
	}

	envPath, err := config.DefaultEnvPath()
	if err != nil {
		return err
	}
	treePath, err := config.DefaultTreePath()
	if err != nil {
		return err
	}

	fmt.Fprintf(a.Stdout, "config: %s\n", envPath)
	fmt.Fprintf(a.Stdout, "tree:   %s\n", treePath)
	return nil
}

func (a App) runToday(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("today", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	noBacklog := fs.Bool("no-backlog", false, "show local records only without calling Backlog API")
	jsonOutput := fs.Bool("json", false, "output as JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("usage: gitwork today [--no-backlog] [--json]")
	}
	return a.printRecordsForCurrentBranch(ctx, *noBacklog, *jsonOutput)
}

func (a App) runEpic(ctx context.Context, args []string) error {
	if len(args) == 0 || args[0] != "status" {
		return errors.New("usage: gitwork epic status [--no-backlog] [--json] [epic-key]")
	}

	fs := flag.NewFlagSet("epic status", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	noBacklog := fs.Bool("no-backlog", false, "show local records only without calling Backlog API")
	jsonOutput := fs.Bool("json", false, "output as JSON")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	epicKey, err := a.resolveEpicKey(ctx, fs.Args())
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

	records := tree.ForEpic(repoRoot, epicKey)
	if *jsonOutput {
		return a.printEpicJSON(ctx, epicKey, records, *noBacklog)
	}

	fmt.Fprintf(a.Stdout, "Epic %s\n\n", strings.ToUpper(epicKey))
	return a.printRecords(ctx, records, *noBacklog)
}

func (a App) resolveEpicKey(ctx context.Context, args []string) (string, error) {
	if len(args) == 1 {
		return parseIssueKey(args[0])
	}
	if len(args) > 1 {
		return "", errors.New("usage: gitwork epic status [epic-key]")
	}

	currentBranch, err := a.Git.CurrentBranch(ctx)
	if err != nil {
		return "", err
	}

	return issueKeyFromBranch(currentBranch)
}

type recordOutput struct {
	IssueKey     string    `json:"issueKey"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	ChildBranch  string    `json:"childBranch"`
	ParentBranch string    `json:"parentBranch"`
	CreatedAt    time.Time `json:"createdAt"`
}

type todayJSONOutput struct {
	CurrentBranch string         `json:"currentBranch"`
	Children      []recordOutput `json:"children"`
}

type epicJSONOutput struct {
	EpicKey string         `json:"epicKey"`
	Records []recordOutput `json:"records"`
}

func (a App) printRecordsForCurrentBranch(ctx context.Context, skipBacklog, jsonOutput bool) error {
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

	records := tree.Children(repoRoot, currentBranch)
	if jsonOutput {
		return a.printTodayJSON(ctx, currentBranch, records, skipBacklog)
	}

	fmt.Fprintf(a.Stdout, "Current branch\n%s\n\nChildren\n", currentBranch)
	return a.printRecords(ctx, records, skipBacklog)
}

func (a App) fetchIssueInfo(ctx context.Context, issueKey string, skipBacklog bool) (title, status string, err error) {
	if skipBacklog || a.Config.ValidateBacklog() != nil {
		return "", "", nil
	}
	issue, err := a.Backlog.GetIssue(ctx, issueKey)
	if err != nil {
		return "", "", err
	}
	return issue.Summary, issue.Status, nil
}

func (a App) warnBacklogFetchFailure(issueKey string, err error) {
	fmt.Fprintf(a.Stderr, "warning: failed to fetch %s from Backlog: %v\n", issueKey, err)
}

func (a App) enrichRecords(ctx context.Context, records []store.Record, skipBacklog bool) ([]recordOutput, error) {
	outputs := make([]recordOutput, 0, len(records))
	for _, record := range records {
		title, status, err := a.fetchIssueInfo(ctx, record.IssueKey, skipBacklog)
		if err != nil {
			a.warnBacklogFetchFailure(record.IssueKey, err)
		}
		outputs = append(outputs, recordOutput{
			IssueKey:     record.IssueKey,
			Title:        title,
			Status:       status,
			ChildBranch:  record.ChildBranch,
			ParentBranch: record.ParentBranch,
			CreatedAt:    record.CreatedAt,
		})
	}
	return outputs, nil
}

func (a App) printTodayJSON(ctx context.Context, currentBranch string, records []store.Record, skipBacklog bool) error {
	children, err := a.enrichRecords(ctx, records, skipBacklog)
	if err != nil {
		return err
	}
	if children == nil {
		children = []recordOutput{}
	}
	return encodeJSON(a.Stdout, todayJSONOutput{
		CurrentBranch: currentBranch,
		Children:      children,
	})
}

func (a App) printEpicJSON(ctx context.Context, epicKey string, records []store.Record, skipBacklog bool) error {
	outputs, err := a.enrichRecords(ctx, records, skipBacklog)
	if err != nil {
		return err
	}
	if outputs == nil {
		outputs = []recordOutput{}
	}
	return encodeJSON(a.Stdout, epicJSONOutput{
		EpicKey: strings.ToUpper(epicKey),
		Records: outputs,
	})
}

func encodeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func (a App) printRecords(ctx context.Context, records []store.Record, skipBacklog bool) error {
	if len(records) == 0 {
		fmt.Fprintln(a.Stdout, "(none)")
		return nil
	}

	for _, record := range records {
		title, status, err := a.fetchIssueInfo(ctx, record.IssueKey, skipBacklog)
		if err != nil {
			a.warnBacklogFetchFailure(record.IssueKey, err)
			title = "-"
			status = "-"
		} else if skipBacklog || a.Config.ValidateBacklog() != nil {
			title = "-"
			status = "-"
		}
		fmt.Fprintf(a.Stdout, "- %s  %s  %s  %s\n", record.IssueKey, record.ChildBranch, title, status)
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


func issueKeyFromBranch(branch string) (string, error) {
	re := regexp.MustCompile(`[A-Za-z]+-\d+`)
	match := re.FindString(branch)
	if match == "" {
		example := workBranchName("member", "backend", "COMMUNITY-102")
		return "", fmt.Errorf(
			"issue key not found in branch %q (expected format, e.g. %s)",
			branch,
			example,
		)
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

func printPRDryRun(w io.Writer, branch, title, base, body, issueKey string, cfg config.Config) {
	fmt.Fprintln(w, "=== Pull Request preview (dry-run) ===")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Title:\n  %s\n\n", title)
	fmt.Fprintf(w, "Base:\n  %s\n\n", base)
	fmt.Fprintln(w, "Body:")
	fmt.Fprintln(w, body)
	fmt.Fprintln(w, "Commands (not executed):")
	fmt.Fprintf(w, "  git push -u origin %s\n", branch)
	fmt.Fprintf(w, "  gh pr create --title %q --body <above> --base %s", title, base)
	if cfg.GitHubRepo != "" {
		fmt.Fprintf(w, " --repo %s", cfg.GitHubRepo)
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  Backlog: update %s status to %d\n", issueKey, cfg.BacklogDoneStatusID)
}
