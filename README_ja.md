# AI-ターミナル

AI-ターミナルは、AI駆動の自動化と最適化を通じてターミナルワークフローを強化する高度なAI駆動のCLIです。ファイル管理、データ処理、システム診断などのタスクを効率的に管理します。

## 主な機能

- **コンテキストアシスタンス:** ユーザーのコマンドから学習し、構文の提案を提供します。
- **自動化タスク:** 繰り返しのタスクパターンを認識し、ショートカットを作成します。
- **インテリジェント検索:** ファイル、ディレクトリ、および特定のファイルタイプ内で検索を行います。
- **エラー修正:** 不正確なコマンドを修正し、代替案を提案します。
- **カスタム統合:** プラグインやAPIを介してさまざまなツールやサービスとの統合をサポートします。

## はじめに

### 前提条件

- Goバージョンv1.22.0以上。

### インストール

Homebrewを使用してインストール：

```bash
brew install coding-hui/tap/ai-terminal
```

または直接ダウンロード：

- [パッケージ][releases] Debian および RPM 形式で提供
- [バイナリ][releases] Linux、macOS、Windows 用

[releases]: https://github.com/coding-hui/ai-terminal/releases

またはソースからビルド（Go 1.22+ が必要）：

```sh
make build
```

設定を初期化：
```sh
ai configure
```

<details>
<summary>シェル補完</summary>

すべてのパッケージとアーカイブには、Bash、ZSH、Fish、PowerShell 用の事前生成された補完ファイルが含まれています。

ソースからビルドした場合、以下のコマンドで生成できます：

```bash
ai completion bash -h
ai completion zsh -h
ai completion fish -h
ai completion powershell -h
```

パッケージ（Homebrew、Debs など）を使用する場合、シェルの設定が正しければ補完は自動的に設定されます。

</details>

### 使用方法

AI-ターミナルの使用例を機能別に紹介します：

#### チャット

- **チャットを開始する：**
  ```sh
  ai ask "Dockerコンテナを管理する最良の方法は何ですか？"
  ```

- **プロンプトファイルを使用する：**
  ```sh
  ai ask --file /path/to/prompt_file.txt
  ```

- **パイプ入力：**
  ```sh
  cat some_script.go | ai ask generate unit tests
  ```

#### コード生成

- **対話型コード生成：**
  ```sh
  ai coder
  ```
  プロンプトに基づいて対話的にコードを生成します。

- **CLIベースのコード生成：**
  ```sh
  ai ctx load /path/to/context_file
  ai coder -c session_id -p "コメントを改善し、単体テストを追加"
  ```
  コンテキストファイルを読み込み、セッションIDを指定してバッチ処理を実行。以下をサポート：
  - コード改善
  - コメントの強化
  - 単体テストの生成
  - コードリファクタリング

- **コンテキスト付きでコードを生成する：**
  ```sh
  ai ctx load /path/to/context_file
  ai coder "feature X を実装"
  ```
  コード生成のための追加情報を提供するために、まずコンテキストファイルを読み込みます。

#### コードレビュー

- **コード変更をレビューする：**
  ```sh
  ai review --exclude-list "*.md,*.txt"
  ```

#### コマンド実行

- **AIによるシェルコマンド実行：**
  ```sh
  ai exec "過去7日間に変更されたファイルをすべて検索"
  ```
  AIを使用して指示を解釈し、適切なシェルコマンドを実行します。

- **確認なしで自動実行：**
  ```sh
  ai exec --yes "すべてのDockerコンテナを一覧表示"
  ```
  推論されたコマンドを確認なしで自動的に実行します。

- **対話型コマンドモード：**
  ```sh
  ai exec --interactive
  ```
  コマンドを改良して実行するための対話型ダイアログを開始します。

#### コミットメッセージ

- **コミットメッセージを生成する：**
  ```sh
  ai commit --diff-unified 3 --lang ja
  ```

## 貢献

貢献を歓迎します！詳細については、[貢献ガイドライン](CONTRIBUTING.md)をご覧ください。

### 変更履歴

プロジェクトの詳細な更新と変更については、[CHANGELOG.md](CHANGELOG.md)をご覧ください。

### ライセンス

2024年 coding-hui。 [MITライセンス](LICENSE)の下でライセンスされています。
