# totonou TODO

## 使いやすさを上げる

- [ ] `totonou init` を追加し、`.env` の雛形作成と保存先表示を対話形式で行えるようにする。（現行設定は `.env` ベース。`config.json` ではない）
- [ ] `totonou doctor` を追加し、Git リポジトリ内か、`gh` がログイン済みか、Backlog 設定が揃っているかをまとめて検査する。
- [ ] `totonou today` と `totonou epic status` に `--json` を追加し、他のツールやスクリプトから扱いやすくする。
- [ ] （推測）`totonou epic status` で課題キーが取れないときも、`pr` と同様に期待するブランチ名例をエラーに含める。

## かっこよくする

- [ ] `--plain` オプションを用意し、CI やスクリプトでは装飾なしの出力に切り替えられるようにする。
- [ ] 端末が TTY のときだけ成功/警告/失敗の記号と色を付ける。
- [ ] `work` 成功時に `created <child> from <parent>` だけでなく、次に使う `totonou pr --dry-run` などの候補を表示する。
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

- [ ] `totonou pr --yes` が確認プロンプトを省略することをテストする。
- [ ] `totonou pr` でユーザーが `n` を入力したときに作成処理を止めることをテストする。
- [ ] `tree.json` が壊れている場合のエラー文と復旧方針をテストする。
- [ ] Backlog API が 4xx/5xx を返したときのユーザー向けエラーをテストする。
- [ ] （推測）`printWelcome`（引数なし起動時のバナー表示）の出力をテストする。

## メンテナンスしやすくする

- [ ] GitHub Actions で `go test ./...` を実行する CI を追加する。
- [ ] `go vet ./...` を CI またはローカル確認手順に追加する。
- [ ] `README.md` に `doctor` や `init` 追加後の初回セットアップ手順を反映する。
- [ ] `CHANGELOG.md` を追加し、日常利用に影響する変更を残す。
- [ ] リリース方法を決める。まずは `go install github.com/.../cmd/totonou@latest` で入れられる形を目標にする。
- [ ] `AGENTS.md` に、新しいコマンドを追加するときのテスト観点を追記する。

## 完了済み

- [x] プロジェクト名を `gitwork` から `Totonou (整)` へリネームする。Go module / バイナリ / コマンド名 / 設定ディレクトリ名 / 環境変数プレフィックス（`GITWORK_` → `TOTONOU_`）/ GitHub リポジトリ名 / ローカルフォルダ名をすべて `totonou` に統一。
- [x] `totonou` 単体実行時に小さめの ASCII アートロゴと主要コマンド一覧を表示する。（`internal/app/help.go` の `bannerArt` / `printWelcome`）
- [x] `README.md` にロゴ（起動時と同じ ASCII アート）を記載する。
- [x] `README.md` に `today --no-backlog` の説明を追記する。
- [x] `README.md` に `config path` の説明を追記する。
- [x] `help` に `config path` コマンドを追記する。（`feature/automation/2026-07-04-help-config-path`）
- [x] `totonou config path` を追加し、設定ファイルと `tree.json` の場所をすぐ確認できるようにする。（`feature/automation/2026-07-04-config-path`）
- [x] 課題キーが見つからないブランチで `totonou pr` を実行したとき、期待するブランチ名例をエラーメッセージに出す。（`feature/automation/2026-07-04-pr-issue-key-hint`）
- [x] `issueKeyFromBranch` の複数パターンをテーブルテスト化する。（`feature/automation/2026-07-04-pr-issue-key-hint`）
- [x] `work` 実行時に同名ブランチがすでに記録されている場合、記録済みの親ブランチも表示する。（`feature/automation/2026-07-04-work-duplicate-parent-hint`）
- [x] `totonou today` に `--no-backlog` を追加し、Backlog API を呼ばずにローカル記録だけ確認できるようにする。（`feature/automation/2026-07-04-today-no-backlog`）
- [x] `totonou pr --dry-run` の出力を、タイトル、base、本文、実行予定コマンドが見やすい順に整える。（`feature/automation/2026-07-04-pr-dry-run-output`）
- [x] `totonou pr --dry-run` が Git push、`gh pr create`、Backlog 更新を実行しないことをテストする。（`feature/automation/2026-07-04-pr-dry-run-output`）
