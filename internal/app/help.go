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
    ・config.json の branchPattern に従い子ブランチを作成
      （例: feature/member/backend/community-102）
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
    例: gitwork today

  epic status <epic-key>
    エピック配下のブランチ・課題を一覧表示します。
    ・課題キーのプレフィックス（例: COMMUNITY-100 → COMMUNITY）で絞り込み
    ・エピック全体の進捗を俯瞰する用途
    例: gitwork epic status COMMUNITY-100

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
  ~/Library/Application Support/gitwork/config.json

環境変数（config.json を上書き）:
  BACKLOG_SPACE_URL, BACKLOG_API_KEY, BACKLOG_DONE_STATUS_ID
  GITHUB_REPO, GITWORK_DEFAULT_BASE, GITWORK_BRANCH_PATTERN, GITWORK_PROJECT_KEY

詳細ヘルプ:
  gitwork help <command>`)
}

func (a App) printCommandHelp(command string) {
	switch command {
	case "work":
		fmt.Fprintln(a.Stdout, `コマンド: work

使い方:
  gitwork work <issue-key>

説明:
  現在 checkout しているブランチを親として、課題キーから子ブランチを作成します。
  エピックブランチの上に、個別タスク用のブランチを積み上げるときに使います。

動作:
  1. 現在のブランチ名を「親」として記録
  2. branchPattern の {issueKey} を課題キー（小文字）に置換してブランチ名を生成
  3. git switch -c で子ブランチを作成
  4. 親子関係・課題キーを tree.json に保存

前提:
  ・Git リポジトリ内で実行すること
  ・config.json に branchPattern を設定しておくこと
    （未設定時は feature/{issueKey}）

例:
  gitwork work COMMUNITY-102
  → feature/member/backend/community-102 を作成（設定による）`)
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
  ・config.json に Backlog API 設定と backlogDoneStatusId が設定されていること

例:
  gitwork pr
  gitwork pr --dry-run
  gitwork pr --yes`)
	case "today":
		fmt.Fprintln(a.Stdout, `コマンド: today

使い方:
  gitwork today

説明:
  現在のブランチから作成した子ブランチ（子タスク）を一覧表示します。
  「今日何をやるか」「エピック配下に何が残っているか」を確認する用途です。

動作:
  1. 現在のブランチ名を取得
  2. tree.json から、このブランチを親に持つ記録を検索
  3. 各子ブランチの課題キーで Backlog からタイトル・ステータスを取得して表示

表示例:
  Current branch
  feature/member/backend/community-101

  Children
  - COMMUNITY-102  API利用画面を実装  対応中
  - COMMUNITY-103  テストを書く        未着手

前提:
  ・gitwork work で子ブランチを作成済みであること
  ・Backlog 設定があると課題タイトル・ステータスも表示される

例:
  gitwork today`)
	case "epic":
		fmt.Fprintln(a.Stdout, `コマンド: epic status

使い方:
  gitwork epic status <epic-key>

説明:
  指定したエピック配下のブランチ・課題を一覧表示します。
  エピック全体の進捗を俯瞰するときに使います。

動作:
  1. エピックキーのプレフィックス（COMMUNITY-100 → COMMUNITY）で絞り込み
  2. 同じリポジトリ内の該当課題キーを tree.json から検索
  3. 各課題の Backlog タイトル・ステータスを表示

表示例:
  Epic COMMUNITY-100

  - COMMUNITY-101  設計書を作成      完了
  - COMMUNITY-102  API利用画面を実装  対応中
  - COMMUNITY-103  テストを書く        未着手

例:
  gitwork epic status COMMUNITY-100`)
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
