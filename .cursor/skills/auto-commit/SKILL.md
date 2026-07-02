---
name: auto-commit
description: 現在の git 差分を分析し Conventional Commits 形式のメッセージでコミットし GitHub へ push する。ユーザーが「コミットして」「自動コミット」「push して」「変更を GitHub に上げて」と依頼したとき、または /auto-commit を指定したときに使用する。
---

# Auto Commit

現在の作業ツリーの差分をもとに、コミットメッセージを生成して GitHub へ push する。

## 前提

- コミットメッセージは `.cursor/rules/conventional-commits.mdc` に従う
- `subject` と `body` は**日本語**で記述する（`type` / `scope` のみ英語 lowerCase）
- ユーザーが明示的に依頼した場合のみ実行する（勝手にコミットしない）
- 秘密情報（`.env`、認証情報など）はコミットしない

## ワークフロー

### 1. 状態確認（並列実行）

```bash
git status
git diff
git diff --staged
git log --oneline -10
git branch -vv
```

- 変更がなければ空コミットは作らず、ユーザーに報告して終了
- ステージ済みと未ステージの両方を確認する

### 2. コミットメッセージ作成

差分から以下を判断する:

| 変更内容 | type の目安 |
|----------|-------------|
| 新機能・新エンドポイント | `feat` |
| バグ修正 | `fix` |
| 構造変更・命名変更 | `refactor` |
| ドキュメントのみ | `docs` |
| テストのみ | `test` |
| CI / ビルド設定 | `ci` / `build` |
| その他雑多 | `chore` |

- `scope` は変更の主な領域（例: `domain`, `infra`, `member`, `like`）
- `subject` は日本語・末尾に `.` なし
- `body` がある場合も日本語で記述する
- 破壊的変更がある場合は footer に `BREAKING CHANGE:` を記載（説明文は日本語）

メッセージ案をユーザーに提示してからコミットしてもよい。依頼が「自動で」系の場合は提示せず実行してよい。

### 3. ステージ・コミット（順次実行）

```bash
# 関連ファイルのみ add（秘密ファイルは除外）
git add <paths>

git commit -m "$(cat <<'EOF'
<type>(<scope>): <日本語の subject>

<日本語の body（任意）>

EOF
)"
```

```bash
git status
```

- フックで失敗した場合は amend せず、修正して新規コミットする
- `--no-verify` は使わない
- `git config` は変更しない

### 4. GitHub へ push

```bash
git push -u origin HEAD
```

- リモート未設定の場合は `-u origin HEAD` で upstream を設定
- `main` / `master` への force push は禁止
- push 前に現在ブランチ名を確認する

### 5. 結果報告

以下を報告する:

- コミットメッセージ
- コミットしたファイル概要
- push 先（ブランチ名・リモート URL）
- 失敗時はエラー内容と次のアクション

## Git 安全ルール

- 破壊的操作（`push --force`, `reset --hard` など）はユーザー明示指示がない限り禁止
- amend は次の条件をすべて満たす場合のみ: ユーザー明示依頼、直前コミットがこのセッションで作成、未 push
- リモートに push 済みのコミットは amend しない

## 例

**差分**: `BaseValueObject` 追加と各 VO の継承変更

```
refactor(domain): 共通 VO 用の基底クラスを導入
```

**差分**: Like 送信 API のバグ修正

```
fix(like): 同一メンバーからの重複いいねを防止
```

**差分**: README 更新のみ

```
docs: readme のセットアップ手順を更新
```
