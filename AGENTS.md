# gitwork

Go製の個人用GitラッパーCLI。Backlog / GitHub CLI / ローカルGit の情報を組み合わせて、ブランチ作成・PR作成・作業一覧を短いコマンドで行う。詳細は `README.md` を参照。

## Cursor Cloud specific instructions

### サービス構成
- 単一の Go 製 CLI (`cmd/gitwork`) のみ。常駐サーバーやDB等の外部サービスはない。
- 実行時に外部プロセス `git` と `gh` (GitHub CLI) を呼び出す。`pr` コマンドは `gh` と Backlog API を使うため、認証・設定なしでは失敗する。

### Go バージョン (重要)
- `go.mod` は `go 1.24` を要求する。VMに標準で入っている Go 1.22 ではビルドできず、`GOTOOLCHAIN` によるダウンロードも環境によっては失敗する。
- Go 1.24 は `/usr/local/go` にインストール済みで、`/usr/bin/go`・`/usr/bin/gofmt` からシンボリックリンクされている（スナップショットに保存済み）。`go version` が 1.24 系を返せばOK。

### ビルド / lint / テスト / 実行
- このリポジトリには CI 設定・Makefile・lint 設定は無い。正規チェックは標準の Go コマンド。
  - ビルド: `go build ./...`
  - vet: `go vet ./...`
  - テスト: `go test ./...`
- `gofmt -l .` と `staticcheck ./...` は既存コードに指摘がある（`internal/app/app.go` のフォーマット、`internal/backlog/client.go` のエラー文言など）。これらはセットアップ起因ではなく元からの状態なので、依頼がない限り修正しない。
- CLI のインストール: `go install ./cmd/gitwork`。バイナリは `$(go env GOPATH)/bin/gitwork`。実行するには `$(go env GOPATH)/bin` を `PATH` に追加する。

### 動作確認の勘所
- `work` / `today` / `epic` は「カレントディレクトリがgitリポジトリであること」を前提にする（`git branch --show-current` 等を呼ぶため）。リポジトリ外で実行するとエラーになる。手元検証は使い捨てのgitリポジトリを作って行うとよい。
- `work` は対話で team (`member`/`admin`/`agency`) と layer (`frontend`/`backend`) を尋ねる。非対話で回すには `--team` / `--layer` を渡す。
- 親子ブランチの記録は `~/.config/gitwork/tree.json`（Linux。`os.UserConfigDir()` 基準）に保存される。`today` / `epic status` はこれを読む。
- Backlog 設定 (`BACKLOG_SPACE_URL` / `BACKLOG_API_KEY`) が無くても `today` / `epic status` は動くが、課題タイトル・ステータスは `-` 表示になる。`pr` は Backlog と `gh` の設定・認証が必須。
