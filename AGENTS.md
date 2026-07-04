# AGENTS.md

## Project Overview

`totonou` (整) is a small Go CLI that helps with a Backlog-based Git workflow. It creates child branches from Backlog issue keys, records parent-child branch relationships locally, shows related work, creates GitHub pull requests through `gh`, and updates Backlog issue status after PR creation.

## Repository Layout

- `cmd/totonou/main.go`: CLI entry point.
- `internal/app`: command dispatch, workflow orchestration, help text, and the startup banner.
- `internal/backlog`: Backlog API client.
- `internal/config`: `.env` file and environment variable loading.
- `internal/git`: local Git and GitHub CLI command wrapper.
- `internal/store`: local branch relationship store.

## Development Commands

```sh
go test ./...
go run ./cmd/totonou help
```

## Implementation Guidelines

- Keep the CLI dependency-free unless a new dependency is clearly justified.
- Prefer small package-level functions and structs that match the current package boundaries.
- Keep external effects behind `internal/git` and `internal/backlog` so tests can use fake command runners and `httptest.Server`.
- Do not call real GitHub, Backlog, or local Git in tests.
- Config is `.env`-based only (see `internal/config/config.go`). Preserve this behavior: env vars set in the shell win over `.env` file values.
- Config directory and env var names use the `totonou`/`TOTONOU_` prefix (e.g. `TOTONOU_ENV_FILE`, `TOTONOU_DEFAULT_BASE`). Do not reintroduce `gitwork`/`GITWORK_` naming.
- Preserve the local data format in `tree.json` unless the user explicitly asks for a migration.
- The startup banner (`internal/app/help.go`, `bannerArt` + `printWelcome`) is shown only for a bare `totonou` invocation with no arguments. Do not show it on every command execution.

## Testing Guidelines

- Add or update focused Go tests with `testing`, `httptest`, and temporary directories.
- Use `t.Setenv` for config environment tests.
- Use fake `gitcmd.Client.Run` functions for command workflow tests.
- Run `go test ./...` after substantive code or test changes.

## Documentation Guidelines

- Keep `README.md` user-facing and focused on commands, configuration, and daily workflow.
- Keep `AGENTS.md` agent-facing and focused on repository conventions and safe editing boundaries.
