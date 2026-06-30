# gitwork

Go製の個人用GitラッパーCLIです。Backlog、GitHub CLI、ローカルGitの情報を合わせて、ブランチ作成、PR作成、今日見るべき作業の確認を短いコマンドで行います。

## Install

```sh
go install ./cmd/gitwork
```

## Config

設定はOS標準のユーザー設定ディレクトリ配下の `gitwork/config.json` から読みます。macOSでは通常 `~/Library/Application Support/gitwork/config.json` です。

```json
{
  "backlogSpaceUrl": "https://example.backlog.com",
  "backlogApiKey": "your-api-key",
  "backlogDoneStatusId": 5,
  "githubRepo": "owner/repo",
  "defaultBase": "develop",
  "branchPattern": "feature/member/backend/{issueKey}",
  "projectKey": "COMMUNITY"
}
```

`backlogApiKey` などは環境変数でも上書きできます。

- `BACKLOG_SPACE_URL`
- `BACKLOG_API_KEY`
- `BACKLOG_DONE_STATUS_ID`
- `GITHUB_REPO`
- `GITWORK_DEFAULT_BASE`
- `GITWORK_BRANCH_PATTERN`
- `GITWORK_PROJECT_KEY`

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
