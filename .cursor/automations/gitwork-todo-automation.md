# gitwork TODO Automation Prompt

このリポジトリ `git-cli` で、定期的に `todo.md` をもとに小さな改善を1つ実装してください。

## 実行手順

1. まず `AGENTS.md`、`README.md`、`todo.md`、`.cursor/skills/gitwork-project-diagnosis/SKILL.md` を読む。
2. `todo.md` の未完了タスクから、現在の実装と矛盾せず、1回の作業で安全に完了できるものを1つだけ選ぶ。
   - 現行実装と矛盾する TODO はそのまま実装しない。
   - 例: 現在の config が `.env` ベースなら、`config.json` 前提の TODO は実装せず、必要なら `todo.md` 側を現状に合わせて修正する。
3. 選んだタスクを、既存のパッケージ境界に沿って最小限で実装する。
   - CLI dispatch / workflow は `internal/app`
   - Git / GitHub CLI 実行は `internal/git`
   - Backlog API は `internal/backlog`
   - ローカル保存は `internal/store`
   - config は `internal/config`
4. 実装に対応する focused な Go テストを追加または更新する。
5. `go test ./...` を実行する。
6. 完了した TODO は `todo.md` で `[x]` にする。実装中に現状と合わない TODO を見つけた場合は、実装せず、現状に即した表現へ更新する。
7. その後、`.cursor/skills/gitwork-project-diagnosis/SKILL.md` の内容に従ってプロジェクト診断を行い、次回以降に取り組むべき TODO を `todo.md` に反映する。
   - 重複 TODO は増やさない。
   - 優先度が下がったものや不要になったものは、理由が明確な場合だけ整理する。
   - 推測で TODO を追加する場合は、`todo.md` 上でも推測であることが分かる表現にする。
8. 最後に、変更内容、選んだ TODO、追加/更新したテスト、`go test ./...` の結果、次回向けに更新した TODO の要点を日本語で報告する。

## 制約

- 1回の Automation で実装する TODO は原則1つだけにする。
- 大きなリファクタリングや新依存の追加は避ける。必要に見える場合は実装せず、`todo.md` に検討タスクとして残す。
- 実際の GitHub、Backlog、破壊的な Git 操作は呼ばない。
- テストでは fake runner、`httptest.Server`、一時ディレクトリを使う。
- `tree.json` の保存形式は、明示的な TODO がない限り変更しない。
- main へ直接 push しない。

## 停止条件と作業上限

- 1回の実行で実装する TODO は必ず1つだけにする。
- 作業開始時に、今回取り組む TODO を1つ選んで明示する。
- 10分以内、または大きな衝突・設計判断・外部認証・ネットワーク失敗に遭遇した時点で実装を止める。
- 同じエラーに対する修正試行は最大2回までにする。
- `go test ./...` の失敗修正も最大2回までにする。
- 2回直しても失敗する場合は、それ以上探索せず、失敗内容と次に人間が判断すべき点を報告して終了する。
- `todo.md` の全項目を一度に処理しようとしない。
- 診断 skill の実行後に `todo.md` を更新したら、追加調査や次の TODO 実装には進まず終了する。
- 最後は必ず「今回実装したこと」「テスト結果」「残した TODO」「停止理由」を短く報告して終了する。
