# リリース手順

このドキュメントでは、istfprojの新しいバージョンをリリースする手順を説明します。

## 前提条件

1. GitHubリポジトリが作成されていること
2. GoReleaserがインストールされていること（`brew install goreleaser`）
3. GitHubのPersonal Access Tokenが設定されていること

## リリース手順

### 1. コードの準備

変更をコミットし、GitHubにプッシュします：

```bash
git add .
git commit -m "feat: 新機能の追加"
git push origin main
```

### 2. タグの作成

セマンティックバージョニングに従ってタグを作成します：

```bash
# 例: v1.0.0, v1.1.0, v2.0.0 など
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### 3. 自動リリース

タグをプッシュすると、GitHub Actionsが自動的に以下を実行します：

- 複数プラットフォーム向けのバイナリをビルド
- アーカイブファイルを作成
- チェックサムを生成
- GitHubリリースを作成
- Homebrewタップを更新（設定している場合）

### 4. リリースの確認

1. GitHubのReleasesページで新しいリリースを確認
2. バイナリがダウンロード可能か確認
3. changelogが正しく生成されているか確認

## ローカルでのテスト

リリース前にローカルでビルドをテストできます：

```bash
# スナップショットビルド（タグなしでテスト）
goreleaser build --snapshot --clean

# 特定のプラットフォームのみビルド
goreleaser build --snapshot --clean --single-target

# リリースのドライラン（実際にはリリースしない）
goreleaser release --snapshot --clean
```

## 環境変数の設定

GitHub Actionsで使用される環境変数：

- `GITHUB_TOKEN`: 自動的に提供される（設定不要）
- `GITHUB_OWNER`: リポジトリのオーナー名（自動設定）
- `GITHUB_REPO`: リポジトリ名（自動設定）
- `DOCKER_REGISTRY`: Dockerイメージのレジストリ（デフォルト: ghcr.io）

## Homebrewタップの設定（オプション）

Homebrewでのインストールを有効にするには：

1. `homebrew-tap`という名前のリポジトリを作成
2. GitHub Personal Access Tokenを作成（`repo`スコープが必要）
3. リポジトリのSecretsに`HOMEBREW_TAP_TOKEN`として追加

## トラブルシューティング

### ビルドが失敗する場合

```bash
# 設定ファイルの検証
goreleaser check

# 依存関係の更新
go mod tidy
```

### タグを間違えた場合

```bash
# ローカルのタグを削除
git tag -d v1.0.0

# リモートのタグを削除
git push origin :refs/tags/v1.0.0
```

## セマンティックバージョニング

- **MAJOR** (v2.0.0): 互換性のない変更
- **MINOR** (v1.1.0): 後方互換性のある機能追加
- **PATCH** (v1.0.1): 後方互換性のあるバグ修正

## コミットメッセージの規約

changelogの自動生成のため、以下の規約に従ってください：

- `feat:` - 新機能
- `fix:` - バグ修正
- `perf:` - パフォーマンス改善
- `docs:` - ドキュメントのみの変更
- `chore:` - ビルドプロセスやツールの変更
- `test:` - テストの追加・修正
- `ci:` - CI設定の変更

例：
```bash
git commit -m "feat: JSON出力機能を追加"
git commit -m "fix: providerブロックの解析エラーを修正"
git commit -m "docs: READMEの使用例を更新"
