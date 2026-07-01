# gitwork

Go製の個人用GitラッパーCLIです。Backlog、GitHub CLI、ローカルGitの情報を合わせて、ブランチ作成、PR作成、今日見るべき作業の確認を短いコマンドで行います。

## Install

```sh
go install ./cmd/gitwork
```

## Config

設定は `.env` のみです。macOS では通常 `~/Library/Application Support/gitwork/.env` です。

```env
BACKLOG_SPACE_URL=https://example.backlog.com
BACKLOG_API_KEY=your-api-key
BACKLOG_DONE_STATUS_ID=5
GITHUB_REPO=owner/repo
GITWORK_DEFAULT_BASE=develop
GITWORK_PROJECT_KEY=COMMUNITY
```

`.env` は起動時に自動読み込みされます。既にシェルに設定済みの環境変数は上書きしません。

任意の `.env` を使う場合:

```sh
export GITWORK_ENV_FILE=/path/to/.env
```

## Commands

```sh
gitwork work COMMUNITY-102
gitwork pr
gitwork today
gitwork epic status
gitwork epic status COMMUNITY-100
```

### work

`gw COMMUNITY-102` 実行時に team / layer を選択します。

- team: `member`, `admin`, `agency`
- layer: `frontend`, `backend`

生成例: `feature/member/backend/COMMUNITY-102`

非対話で指定する場合:

```sh
gitwork work COMMUNITY-102 --team member --layer backend
```

### epic status

epic キーを省略すると、現在のブランチ名から課題キーを取得して epic として使います。

```sh
gitwork epic status
```

日常利用はaliasで短くできます。

```sh
alias gw='gitwork work'
alias gp='gitwork pr'
alias gt='gitwork today'
alias ges='gitwork epic status'
```

`gitwork pr` はPR作成前に確認します。確認だけしたい場合は `gitwork pr --dry-run` を使います。
