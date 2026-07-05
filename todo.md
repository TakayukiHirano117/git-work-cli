# gitwork TODO

## 使いやすさを上げる

## かっこよくする

- [ ] `gitwork` 単体実行時に小さめの ASCII アートロゴと主要コマンド一覧を表示する。
- [ ] `--plain` オプションを用意し、CI やスクリプトでは装飾なしの出力に切り替えられるようにする。
- [ ] 端末が TTY のときだけ成功/警告/失敗の記号と色を付ける。
- [ ] `work` 成功時に `created <child> from <parent>` だけでなく、次に使う `gitwork pr --dry-run` などの候補を表示する。
- [ ] `pr` 作成前の確認画面を、PR タイトル、Backlog URL、base、作成後ステータス更新の順で見やすく表示する。

## 保守性を上げる

- [ ] `internal/app` の各コマンド処理を `work.go`、`pr.go`、`today.go`、`epic.go` に分ける。
- [ ] コマンド引数の parse と実行処理を分離し、各コマンドのテストを小さく書けるようにする。
- [ ] 出力整形を `internal/output` のような小さなパッケージに切り出し、JSON 出力や装飾出力を追加しやすくする。
- [ ] `backlog.Client` のエラーに完全な endpoint URL を含め、調査しやすくする。（診断: method/path/HTTP status は既に含まれている）
- [ ] `git.Client` の実行ログを dry-run/doctor 用に再利用できる形で整理する。
- [ ] `Config` に `ValidatePR`、`ValidateToday` など用途別の検証関数を追加し、必要な設定だけを明確にする。
- [ ] `tree.json` の読み書きにバージョンフィールドを追加するか検討し、将来の形式変更に備える。
- [ ] `projectKey` と `BranchPattern` が未使用に見えるため、使う方針か削除する方針かを決める。（`work` は `feature/<team>/<layer>/<issue>` 固定）
- [ ] `.env` の形式不正があると `doctor` を含む主要コマンドが起動前に失敗する。切り分け用コマンドでも設定読み込みを緩和するか検討する。（診断: `config path` / `init` のみ `loadDeps: false`。次回最優先）
- [ ] `work` でブランチ作成後に `tree.json` 記録が失敗した場合の復旧方針を決める。（診断: Git 操作と記録が非トランザクション。失敗時に作成済みブランチ名を明示するのが第一歩）
- [ ] `internal/git` に fake runner を使った単体テストを追加する。（診断: `internal/git` にテストファイルがない）

## メンテナンスしやすくする

- [ ] `README.md` に `today --no-backlog` / `epic status --no-backlog` と `--json` の説明を追記する。
- [ ] `README.md` に `config path` の説明を追記する。（診断: First-time Setup には追記済みだが Commands 節には未記載）
- [ ] `README.md` の Config 節で `.env` 自動読み込みの説明が重複しているため整理する。（診断: 同一文が2行連続）
- [ ] `README.md` / `help` に Linux 向けの設定ファイルパス例を追記する。（診断: 実装は `os.UserConfigDir()` ベースだが説明は macOS のみ）
- [ ] `help` の doctor 説明に `github config` 検査を追記する。（診断: 実装済みだが help 未更新）
- [ ] `pr --dry-run` が `GITHUB_REPO` 未設定でも PR プレビューできるようにするか検討する。（診断: `ValidateGitHub` が dry-run 前に走る。現状は早期検出を優先）
- [ ] `help` の一般説明に `today` / `epic status` の `--json` / `--no-backlog` オプションを追記する。（診断: コマンド別ヘルプにはあるが一般ヘルプに未記載。`today` の `--json` も同様）
- [ ] `help` の一般説明に `work` の `--team` / `--layer` オプションを追記する。（診断: コマンド別ヘルプにはあるが一般ヘルプに未記載）
- [ ] `CHANGELOG.md` を追加し、日常利用に影響する変更を残す。
- [ ] リリース方法を決める。まずは `go install github.com/.../cmd/gitwork@latest` で入れられる形を目標にする。
- [ ] `AGENTS.md` に、新しいコマンドを追加するときのテスト観点を追記する。

## 完了済み

- [x] `help` の doctor 説明に `github config` 検査を追記する。（`feature/automation/2026-07-05-help-doctor-github-config`）
- [x] `README.md` に `doctor`・`init`・`config path` を含む初回セットアップ手順を反映する。（`feature/automation/2026-07-05-readme-setup`）
- [x] `pr` / `doctor` で `GITHUB_REPO` 未設定を早期に検出し、`gh pr create` 失敗前に案内する。（`feature/automation/2026-07-05-github-repo-validation`）
- [x] 未知コマンドでも `config.Load()` が先に走る。`isKnownCommand` で `withDeps` より前に検証する。（`feature/automation/2026-07-04-unknown-command-before-config`）
- [x] `today` / `epic status` のプレーンテキスト出力に子ブランチ名を含める。（`feature/automation/2026-07-04-today-epic-plain-branch-name`）
- [x] `.env` 形式不正時に `Load()` が行番号付きエラーを返すことをテストする。（`feature/automation/2026-07-04-invalid-env-line-number-test`）
- [x] `work` で `--team` / `--layer` 未指定かつ stdin が TTY でないとき、対話待ちせずエラーにする。（`feature/automation/2026-07-04-work-non-interactive-flags`）
- [x] `BACKLOG_DONE_STATUS_ID` が数値でない場合に黙って無視される。`Load` で明示的にエラーにする。（`feature/automation/2026-07-04-invalid-done-status-id`）
- [x] `pr` で Backlog 更新失敗時に push/PR が実行済みであることをテストする。（`feature/automation/2026-07-04-pr-backlog-update-failure-hint`）
- [x] `today` / `epic status` で一部の Backlog 取得が失敗しても、他の記録は表示を続ける。失敗分は stderr に warning を出し、plain 出力では `-`、JSON では空の title/status で残す。（`feature/automation/2026-07-04-partial-backlog-failure`）
- [x] `epic status <epic-key>` にも課題キー形式の検証を追加する。`parseIssueKey` を再利用し、不正キーを早期に弾く。（`feature/automation/2026-07-04-epic-issue-key-validation`）
- [x] `work` に課題キー形式の検証を追加する。`pr` / `epic status` と同様に `COMMUNITY-102` 形式を要求し、不正キーではブランチを作らない。（`feature/automation/2026-07-04-work-issue-key-validation`）
- [x] `go vet ./...` を CI またはローカル確認手順に追加する。（`feature/automation/2026-07-04-go-vet-ci`）
- [x] `epic status` に `--no-backlog` を追加し、`today` と同様に Backlog API を省略できるようにする。（`feature/automation/2026-07-04-epic-no-backlog`）
- [x] `gitwork today` と `gitwork epic status` に `--json` を追加し、他のツールやスクリプトから扱いやすくする。（`feature/automation/2026-07-04-today-epic-json`）
- [x] `gitwork init` を追加し、`.env` の雛形作成と保存先表示を対話形式で行えるようにする。（`feature/automation/2026-07-04-init`）
- [x] GitHub Actions で `go test ./...` を実行する CI を追加する。（`feature/automation/2026-07-04-github-actions-ci`）
- [x] Backlog API が 4xx/5xx を返したときのユーザー向けエラーをテストする。（`internal/backlog` には HTTP エラー検証あり。`internal/app` 経由の表示は未テスト）（`feature/automation/2026-07-04-backlog-api-error-app-test`）
- [x] `gitwork doctor` を追加し、Git リポジトリ内か、`gh` がログイン済みか、Backlog 設定が揃っているかをまとめて検査する。（`feature/automation/2026-07-04-doctor`）
- [x] `tree.json` が壊れている場合のエラー文と復旧方針をテストする。（`feature/automation/2026-07-04-corrupt-tree-json-test`）
- [x] `gitwork epic status` で課題キーが取れないとき、`pr` と同様に期待するブランチ名例をエラーに含める。（`feature/automation/2026-07-04-epic-issue-key-hint`）
- [x] `gitwork pr` でユーザーが `n` を入力したときに作成処理を止めることをテストする。（`feature/automation/2026-07-04-pr-no-cancel`）
- [x] `gitwork pr --yes` が確認プロンプトを省略することをテストする。（`feature/automation/2026-07-04-pr-yes-skip-confirm`）
- [x] `help` に `config path` コマンドを追記する。（`feature/automation/2026-07-04-help-config-path`）
- [x] `gitwork config path` を追加し、設定ファイルと `tree.json` の場所をすぐ確認できるようにする。（`feature/automation/2026-07-04-config-path`）
- [x] 課題キーが見つからないブランチで `gitwork pr` を実行したとき、期待するブランチ名例をエラーメッセージに出す。（`feature/automation/2026-07-04-pr-issue-key-hint`）
- [x] `issueKeyFromBranch` の複数パターンをテーブルテスト化する。（`feature/automation/2026-07-04-pr-issue-key-hint`）
- [x] `work` 実行時に同名ブランチがすでに記録されている場合、記録済みの親ブランチも表示する。（`feature/automation/2026-07-04-work-duplicate-parent-hint`）
- [x] `gitwork today` に `--no-backlog` を追加し、Backlog API を呼ばずにローカル記録だけ確認できるようにする。（`feature/automation/2026-07-04-today-no-backlog`）
- [x] `gitwork pr --dry-run` の出力を、タイトル、base、本文、実行予定コマンドが見やすい順に整える。（`feature/automation/2026-07-04-pr-dry-run-output`）
- [x] `gitwork pr --dry-run` が Git push、`gh pr create`、Backlog 更新を実行しないことをテストする。（`feature/automation/2026-07-04-pr-dry-run-output`）
