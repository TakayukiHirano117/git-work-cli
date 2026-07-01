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
GITWORK_BRANCH_PATTERN=feature/{issueKey}
GITWORK_PROJECT_KEY=COMMUNITY
```

`.env` は起動時に自動読み込みされます。既にシェルに設定済みの環境変数は上書きしません。

任意の `.env` を使う場合:

```sh
export GITWORK_ENV_FILE=/path/to/.env
```

`GITWORK_BRANCH_PATTERN` の `{issueKey}` は課題キー（小文字）に置換されます。  
`member` や `backend` はコード固定ではなく、README やテストに載っていた例です。  
自分の命名規則に合わせて `.env` で自由に変えられます。

例:

```env
GITWORK_BRANCH_PATTERN=feature/member/backend/{issueKey}
```

## Commands

```sh
gitwork work COMMUNITY-102
gitwork pr
gitwork today
gitwork epic status COMMUNITY-100
```

日常利用はaliasで短くできます。

```sh
alias gw='gitwork work'
alias gp='gitwork pr'
alias gt='gitwork today'
alias ges='gitwork epic status'
```

`gitwork pr` はPR作成前に確認します。確認だけしたい場合は `gitwork pr --dry-run` を使います。
