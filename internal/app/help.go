package app

import "fmt"

func (a App) printHelp(command string) {
	if command != "" {
		a.printCommandHelp(command)
		return
	}
	a.printGeneralHelp()
}

func (a App) printGeneralHelp() {
	fmt.Fprintln(a.Stdout, `gitwork - Backlog / GitHub CLI / Git ラッパー

Backlog の課題、GitHub の PR、ローカル Git のブランチ操作を
短いコマンドでまとめて行う個人用 CLI です。

使い方:
  gitwork <command> [arguments]

コマンド一覧:

  work <issue-key>
    課題キーから作業用ブランチを作成します。
    ・今いるブランチを「親」として記録
    ・team / layer を対話選択（member/admin/agency, frontend/backend）
    ・feature/<team>/<layer>/<ISSUE-KEY> 形式で子ブランチを作成
      （例: feature/member/backend/COMMUNITY-100）
    ・親子関係を tree.json に保存（today / epic で後から参照）
    例: gitwork work COMMUNITY-102

  pr
    現在のブランチから Pull Request を作成します。
    ・ブランチ名から課題キーを自動抽出（COMMUNITY-102 形式）
    ・Backlog から課題タイトル・URL を取得して PR 本文を生成
    ・git push → gh pr create を実行
    ・PR 作成後、Backlog のステータスを完了に更新
    オプション: --dry-run（確認のみ） / --yes（確認スキップ）
    例: gitwork pr

  today
    今日見るべき子タスクを一覧表示します。
    ・今いるブランチを「親」として、そこから切った子ブランチを表示
    ・各課題の Backlog タイトル・ステータスも表示
    ・エピックブランチにいるとき、配下の作業状況を確認する用途
    オプション: --no-backlog（Backlog API を呼ばずローカル記録のみ表示）
    例: gitwork today

  epic status [epic-key]
    エピック配下のブランチ・課題を一覧表示します。
    ・epic-key 省略時は現在ブランチの課題キーを epic として使う
    ・課題キーのプレフィックス（例: COMMUNITY-100 → COMMUNITY）で絞り込み
    ・エピック全体の進捗を俯瞰する用途
    例: gitwork epic status COMMUNITY-100
    例: gitwork epic status

  config path
    設定ファイル (.env) と tree.json の保存場所を表示します。
    初回セットアップ時やパス確認に使います。
    例: gitwork config path

  help [command]
    ヘルプを表示します。コマンド名を指定すると詳細を表示します。
    例: gitwork help pr

よくある流れ:
  1. エピック用ブランチに checkout
  2. gitwork work COMMUNITY-102   # 子タスク用ブランチを作成
  3. コーディング・コミット
  4. gitwork pr                   # PR 作成 + Backlog 更新
  5. gitwork today                # 残りの子タスクを確認

設定ファイル (macOS):
  ~/Library/Application Support/gitwork/.env

環境変数:
  BACKLOG_SPACE_URL, BACKLOG_API_KEY, BACKLOG_DONE_STATUS_ID, GITHUB_REPO
  GITWORK_DEFAULT_BASE, GITWORK_PROJECT_KEY

詳細ヘルプ:
  gitwork help <command>`)
}

func (a App) printCommandHelp(command string) {
	switch command {
	case "work":
		fmt.Fprintln(a.Stdout, `コマンド: work

使い方:
  gitwork work <issue-key> [--team member|admin|agency] [--layer frontend|backend]

説明:
  現在 checkout しているブランチを親として、課題キーから子ブランチを作成します。
  エピックブランチの上に、個別タスク用のブランチを積み上げるときに使います。

動作:
  1. 現在のブランチ名を「親」として記録
  2. team / layer を選択（--team / --layer 省略時は対話選択）
  3. feature/<team>/<layer>/<ISSUE-KEY> のブランチ名を生成
  4. git switch -c で子ブランチを作成
  5. 親子関係・課題キーを tree.json に保存

前提:
  ・Git リポジトリ内で実行すること

例:
  gitwork work COMMUNITY-102
  → feature/member/backend/COMMUNITY-102 を作成（選択内容による）`)
	case "pr":
		fmt.Fprintln(a.Stdout, `コマンド: pr

使い方:
  gitwork pr [--dry-run] [--yes]

説明:
  現在のブランチから Pull Request を作成します。
  Backlog の課題情報を PR 本文に自動で埋め込み、作成後に Backlog のステータスも更新します。

動作:
  1. ブランチ名から課題キーを抽出（例: community-102 → COMMUNITY-102）
  2. Backlog API で課題のタイトル・URL を取得
  3. PR タイトル・本文を表示し、確認プロンプトを表示
  4. git push -u origin <branch> でリモートに push
  5. gh pr create で PR を作成
  6. Backlog のステータスを backlogDoneStatusId に更新

オプション:
  --dry-run   PR の内容を表示するだけ。push も作成もしない
  --yes       確認プロンプトをスキップしてそのまま作成

前提:
  ・ブランチ名に課題キー（COMMUNITY-102 形式）が含まれていること
  ・gh（GitHub CLI）がインストール・認証済みであること
  ・.env に Backlog API 設定と BACKLOG_DONE_STATUS_ID が設定されていること

例:
  gitwork pr
  gitwork pr --dry-run
  gitwork pr --yes`)
	case "today":
		fmt.Fprintln(a.Stdout, `コマンド: today

使い方:
  gitwork today [--no-backlog]

説明:
  現在のブランチから作成した子ブランチ（子タスク）を一覧表示します。
  「今日何をやるか」「エピック配下に何が残っているか」を確認する用途です。

動作:
  1. 現在のブランチ名を取得
  2. tree.json から、このブランチを親に持つ記録を検索
  3. 各子ブランチの課題キーで Backlog からタイトル・ステータスを取得して表示
     （--no-backlog 指定時は Backlog API を呼ばず、課題キーのみ表示）

表示例:
  Current branch
  feature/community-101

  Children
  - COMMUNITY-102  API利用画面を実装  対応中
  - COMMUNITY-103  テストを書く        未着手

前提:
  ・gitwork work で子ブランチを作成済みであること
  ・Backlog 設定があると課題タイトル・ステータスも表示される

オプション:
  --no-backlog   Backlog API を呼ばず、tree.json のローカル記録だけを表示

例:
  gitwork today
  gitwork today --no-backlog`)
	case "config":
		fmt.Fprintln(a.Stdout, `コマンド: config path

使い方:
  gitwork config path

説明:
  設定ファイル (.env) と tree.json の保存場所を表示します。
  初回セットアップ時や、設定ファイルの場所を確認したいときに使います。

表示例:
  config: /Users/you/Library/Application Support/gitwork/.env
  tree:   /Users/you/Library/Application Support/gitwork/tree.json

補足:
  ・設定は .env のみです（config.json は使いません）
  ・任意の .env を使う場合は GITWORK_ENV_FILE 環境変数を設定します

例:
  gitwork config path`)
	case "epic":
		fmt.Fprintln(a.Stdout, `コマンド: epic status

使い方:
  gitwork epic status [epic-key]

説明:
  指定したエピック配下のブランチ・課題を一覧表示します。
  epic-key を省略した場合は、現在のブランチ名から課題キーを取得して epic として使います。
  エピック全体の進捗を俯瞰するときに使います。

動作:
  1. epic-key 省略時は現在ブランチから課題キーを取得
  2. エピックキーのプレフィックス（COMMUNITY-100 → COMMUNITY）で絞り込み
  3. 同じリポジトリ内の該当課題キーを tree.json から検索
  4. 各課題の Backlog タイトル・ステータスを表示

表示例:
  Epic COMMUNITY-100

  - COMMUNITY-101  設計書を作成      完了
  - COMMUNITY-102  API利用画面を実装  対応中
  - COMMUNITY-103  テストを書く        未着手

例:
  gitwork epic status COMMUNITY-100
  gitwork epic status`)
	case "help":
		a.printGeneralHelp()
	default:
		fmt.Fprintf(a.Stdout, "不明なコマンド: %s\n\n", command)
		a.printGeneralHelp()
	}
}

func isHelpRequest(args []string) bool {
	switch args[0] {
	case "help", "-h", "--help":
		return true
	default:
		return false
	}
}

func helpCommandName(args []string) string {
	if args[0] == "help" && len(args) > 1 {
		return args[1]
	}
	return ""
}

func isSubcommandHelp(args []string) bool {
	if len(args) == 0 {
		return false
	}
	switch args[0] {
	case "-h", "--help":
		return true
	default:
		return false
	}
}
