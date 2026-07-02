# AGENTS.md

## Project Overview

`gitwork` is a small Go CLI that helps with a Backlog-based Git workflow. It creates child branches from Backlog issue keys, records parent-child branch relationships locally, shows related work, creates GitHub pull requests through `gh`, and updates Backlog issue status after PR creation.

## Repository Layout

- `cmd/gitwork/main.go`: CLI entry point.
- `internal/app`: command dispatch and workflow orchestration.
- `internal/backlog`: Backlog API client.
- `internal/config`: config file and environment variable loading.
- `internal/git`: local Git and GitHub CLI command wrapper.
- `internal/store`: local branch relationship store.

## Development Commands

```sh
go test ./...
go run ./cmd/gitwork help
```

## Implementation Guidelines

- Keep the CLI dependency-free unless a new dependency is clearly justified.
- Prefer small package-level functions and structs that match the current package boundaries.
- Keep external effects behind `internal/git` and `internal/backlog` so tests can use fake command runners and `httptest.Server`.
- Do not call real GitHub, Backlog, or local Git in tests.
- Preserve the current config behavior: config file values are loaded first, then non-empty environment variables override them.
- Preserve the local data format in `tree.json` unless the user explicitly asks for a migration.

## Testing Guidelines

- Add or update focused Go tests with `testing`, `httptest`, and temporary directories.
- Use `t.Setenv` for config environment tests.
- Use fake `gitcmd.Client.Run` functions for command workflow tests.
- Run `go test ./...` after substantive code or test changes.

## Documentation Guidelines

- Keep `README.md` user-facing and focused on commands, configuration, and daily workflow.
- Keep `AGENTS.md` agent-facing and focused on repository conventions and safe editing boundaries.
