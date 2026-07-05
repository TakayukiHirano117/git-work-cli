```
█████   ███   █████   ███   █   █   ███   █   █
  █    █   █    █    █   █  ██  █  █   █  █   █
  █    █   █    █    █   █  █ █ █  █   █  █   █
  █    █   █    █    █   █  █  ██  █   █  █   █
  █     ███     █     ███   █   █   ███    ███
```

# Totonou (整)

`totonou` は Go 製の個人用 Git ラッパー CLI です。Backlog の課題、GitHub CLI、ローカル Git の情報をつなぎ、作業ブランチの作成、Pull Request 作成、派生作業の確認を短いコマンドで実行します。

タスクとブランチを「整える」ことで、頭の中で追いかける情報を減らし、脳のリソースをコーディングに使えるようにすることを目指しています。

## Features

- Backlog 課題キーから規約に沿った作業ブランチを作成します。
- 親ブランチと子ブランチの関係をローカルの `tree.json` に記録します。
- 現在のブランチから作成した子ブランチを `today` で一覧表示します。
- 同じプロジェクトキーの作業を `epic status` で一覧表示します。
- Backlog の課題概要から PR タイトル/本文を作り、`gh pr create` で PR を作成します。
- PR 作成後に Backlog 課題のステータスを完了ステータスへ更新します。
- 引数なしで実行すると、ロゴと主なコマンド一覧を表示します。

## Requirements

- Go 1.24
- Git
- GitHub CLI (`gh`)
- Backlog API key

`totonou pr` は `gh pr create` と Backlog API を使います。事前に `gh auth login` などで GitHub CLI を利用できる状態にしてください。

## Install

```sh
go install ./cmd/totonou
```

開発中に直接実行する場合は次のようにします。

```sh
go run ./cmd/totonou --help
```

## First-time Setup

初回は次の順で設定します。

1. 保存先を確認する:

```sh
gitwork config path
```

`config` に `.env`、`tree` に `tree.json` のパスが表示されます。OS ごとの実際のパスは `os.UserConfigDir()` 配下です（macOS では `~/Library/Application Support/gitwork/`、Linux では通常 `~/.config/gitwork/`）。

2. `.env` の雛形を作成する:

```sh
gitwork init
```

保存先を表示し、対話形式で雛形作成を確認します。既に `.env` がある場合は上書きしません。

3. 表示された `.env` を編集し、Backlog と GitHub の値を入れます（下記 Config を参照）。

4. 設定と環境を検査する:

```sh
gitwork doctor
```

Git リポジトリ内か、`gh` 認証、Backlog 設定、`GITHUB_REPO` が揃っているかを確認します。問題があればメッセージに従って修正してください。

## Config

設定は `.env` のみです。パスは `gitwork config path` で確認できます。macOS では通常 `~/Library/Application Support/gitwork/.env` です。

```env
BACKLOG_SPACE_URL=https://example.backlog.com
BACKLOG_API_KEY=
BACKLOG_DONE_STATUS_ID=
GITHUB_REPO=owner/repo
TOTONOU_DEFAULT_BASE=develop
TOTONOU_PROJECT_KEY=COMMUNITY
```

`.env` は起動時に自動読み込みされます。既にシェルに設定済みの環境変数は上書きしません。`TOTONOU_DEFAULT_BASE` は未指定の場合 `develop` です。

任意の `.env` を使う場合:

```sh
export TOTONOU_ENV_FILE=/path/to/.env
```

## Commands

### `totonou work <issue-key>`

現在のブランチを親ブランチとして記録し、指定した Backlog 課題キーから子ブランチを作成します。

```sh
totonou work COMMUNITY-102
```

team / layer を選択します。

- team: `member`, `admin`, `agency`
- layer: `frontend`, `backend`

生成例: `feature/member/backend/COMMUNITY-102`

非対話で指定する場合:

```sh
totonou work COMMUNITY-102 --team member --layer backend
```

### `totonou pr [--dry-run] [--yes]`

現在のブランチ名から Backlog 課題キーを抽出し、課題概要をもとに PR を作成します。

```sh
totonou pr --dry-run
totonou pr --yes
```

`--dry-run` は PR タイトル、base、本文、実行予定コマンドだけを表示します。`--yes` は確認プロンプトを省略します。

### `totonou today [--no-backlog]`

現在のブランチから作成された子ブランチを表示します。Backlog 設定がある場合は課題タイトルとステータスも取得します。

```sh
totonou today
totonou today --no-backlog
```

`--no-backlog` を指定すると Backlog API を呼ばず、`tree.json` のローカル記録だけを表示します。

### `totonou epic status [epic-key]`

現在のリポジトリで記録されている作業のうち、指定した課題キーと同じプロジェクトキーを持つものを表示します。epic-key を省略すると、現在のブランチ名から課題キーを取得します。

```sh
totonou epic status
totonou epic status COMMUNITY-100
```

### `totonou config path`

設定ファイル (`.env`) と `tree.json` の保存場所を表示します。

```sh
totonou config path
```

## Daily Aliases

```sh
alias tw='totonou work'
alias tp='totonou pr'
alias tt='totonou today'
alias tes='totonou epic status'
```

## Local Data

`totonou work` で作成した親子関係は設定ディレクトリ配下の `totonou/tree.json` に保存されます。macOS では通常 `~/Library/Application Support/totonou/tree.json` です。

保存されるのはリポジトリルート、親ブランチ、子ブランチ、課題キー、作成日時です。Backlog API key などの認証情報は `tree.json` には保存されません。

## Development

```sh
go test ./...
go run ./cmd/totonou help
```

外部コマンドや Backlog API を扱うコードは、`internal/git` と `internal/backlog` に分離しています。テストでは fake runner や `httptest.Server` を使い、実際の GitHub/Backlog には接続しません。
