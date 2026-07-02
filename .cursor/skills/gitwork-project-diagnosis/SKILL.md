---
name: gitwork-project-diagnosis
description: Diagnose this gitwork Go CLI for prioritized improvements and possible feature additions. Use when reviewing this project for product usefulness, implementation quality, tests, documentation, or future roadmap ideas.
disable-model-invocation: true
---

# gitwork Project Diagnosis

## Purpose

Use this skill to inspect the `gitwork` project and produce a concise, prioritized diagnosis of improvements and feature candidates.

The diagnosis should balance:
- User-facing usefulness
- Implementation quality
- Test coverage
- Documentation
- Maintenance cost

## Project Context

`gitwork` is a small Go CLI for a Backlog-based Git workflow. It connects local Git, GitHub CLI, Backlog API, and a local `tree.json` branch relationship store.

Respect the repository boundaries:
- CLI dispatch and workflows live in `internal/app`.
- External Git and GitHub CLI effects live behind `internal/git`.
- Backlog API effects live behind `internal/backlog`.
- Local branch relationship storage lives in `internal/store`.
- Config loading lives in `internal/config`.

Do not recommend new dependencies unless they materially reduce complexity or risk.

## Diagnosis Workflow

1. Read `AGENTS.md`, `README.md`, and `todo.md` first.
2. Inspect the current command surface in `cmd/gitwork/main.go` and `internal/app`.
3. Inspect package boundaries in `internal/config`, `internal/git`, `internal/backlog`, and `internal/store`.
4. Check existing tests before claiming a test gap.
5. Treat `todo.md` as prior art, not as final truth. Confirm whether each idea is still valuable from the current code.
6. Prefer improvements that preserve the current local `tree.json` data format unless the user explicitly asks for migration planning.

## What To Look For

Prioritize findings that improve everyday CLI use:
- Missing setup or diagnostics commands such as `init`, `doctor`, or config path discovery.
- Confusing or incomplete error messages.
- Unsafe or surprising side effects in `pr`, especially around pushing, PR creation, and Backlog status updates.
- Output formats that block scripting, such as missing `--json`.
- Features already hinted by existing config fields or TODOs.

Also check maintainability:
- Command parsing mixed with execution logic.
- Testability of external effects.
- Error messages from HTTP, Git, GitHub CLI, and config loading.
- Duplication in output formatting.
- Documentation drift between README, help text, and implemented behavior.

## Output Format

Respond in Japanese with this structure:

```markdown
## 優先度高
- `Title`: Why it matters. Evidence from files. Suggested first implementation step.

## 優先度中
- `Title`: Why it matters. Evidence from files. Suggested first implementation step.

## 優先度低
- `Title`: Why it matters. Evidence from files. Suggested first implementation step.

## 見送ってよさそうなもの
- `Title`: Reason to defer or avoid.

## 次に着手するなら
One short recommendation for the best next task.
```

Keep each item grounded in code or docs. If something is speculative, label it as speculative.

## Constraints

- Do not change implementation as part of the diagnosis unless the user explicitly asks.
- Do not call real GitHub, Backlog, or local destructive Git commands.
- If running checks, prefer `go test ./...` and summarize the result.
- Avoid broad refactors as recommendations unless they unblock a concrete feature or reduce real test complexity.
