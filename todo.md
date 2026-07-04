# gitwork TODO

## 使いやすさを上げる

- [ ] `gitwork init` を追加し、`.env` の雛形作成と保存先表示を対話形式で行えるようにする。（現行設定は `.env` ベース。`config.json` ではない）
- [ ] `gitwork doctor` を追加し、Git リポジトリ内か、`gh` がログイン済みか、Backlog 設定が揃っているかをまとめて検査する。
- [ ] `gitwork today` と `gitwork epic status` に `--json` を追加し、他のツールやスクリプトから扱いやすくする。
- [ ] （推測）`gitwork epic status` で課題キーが取れないときも、`pr` と同様に期待するブランチ名例をエラーに含める。

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
- [ ] `backlog.Client` のエラーに HTTP status と endpoint を含め、調査しやすくする。
- [ ] `git.Client` の実行ログを dry-run/doctor 用に再利用できる形で整理する。
- [ ] `Config` に `ValidatePR`、`ValidateToday` など用途別の検証関数を追加し、必要な設定だけを明確にする。
- [ ] `tree.json` の読み書きにバージョンフィールドを追加するか検討し、将来の形式変更に備える。
- [ ] `projectKey` と `BranchPattern` が未使用に見えるため、使う方針か削除する方針かを決める。（`work` は `feature/<team>/<layer>/<issue>` 固定）

## テストを増やす

- [ ] `tree.json` が壊れている場合のエラー文と復旧方針をテストする。
- [ ] Backlog API が 4xx/5xx を返したときのユーザー向けエラーをテストする。（`internal/backlog` には HTTP エラー検証あり。`internal/app` 経由の表示は未テスト）

## メンテナンスしやすくする

- [ ] GitHub Actions で `go test ./...` を実行する CI を追加する。
- [ ] `go vet ./...` を CI またはローカル確認手順に追加する。
- [ ] `README.md` に `doctor` や `init` 追加後の初回セットアップ手順を反映する。
- [ ] `README.md` に `today --no-backlog` の説明を追記する。
- [ ] `README.md` に `config path` の説明を追記する。（診断: help には追記済みだが README には未記載）
- [ ] `CHANGELOG.md` を追加し、日常利用に影響する変更を残す。
- [ ] リリース方法を決める。まずは `go install github.com/.../cmd/gitwork@latest` で入れられる形を目標にする。
- [ ] `AGENTS.md` に、新しいコマンドを追加するときのテスト観点を追記する。

## 完了済み

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
