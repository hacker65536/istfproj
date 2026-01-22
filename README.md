# istfproj - Terraform Project Checker

Terraformプロジェクトの状態を判定するGoツールです。`github.com/hashicorp/hcl/v2`ライブラリを使用してHCLファイルを解析します。

**istfproj** = "is terraform project" の略

## 機能

- 指定されたディレクトリがTerraformプロジェクトかどうかを判定
- デフォルト: *.tfファイルの有無で判定（終了コード 0 or 1 で結果を返す）
- `--strict`オプション: providerブロックの有無も厳格にチェック
- `--json`オプション: providerブロックの詳細情報をJSON形式で出力

## Terraformプロジェクトの判定基準

### デフォルトモード
- *.tfファイルが存在すれば終了コード `0`（成功）
- Terraformは自動でproviderを検知してインストールするため、providerブロックがなくてもTerraformプロジェクトとして扱う

### Strictモード（`--strict`）
- *.tfファイルが存在し、かつproviderブロックが記述されている場合のみ終了コード `0`（成功）
- moduleかどうかの判断の参考になる（moduleは通常providerブロックを持たない）

## インストール

### バイナリをダウンロード（推奨）

[Releases](https://github.com/hacker65536/istfproj/releases)ページから、お使いのOSとアーキテクチャに合ったバイナリをダウンロードしてください。

### Homebrew (macOS/Linux)

```bash
brew install hacker65536/tap/istfproj
```

### Go install

```bash
go install github.com/hacker65536/istfproj@latest
```

### ソースからビルド

```bash
git clone https://github.com/hacker65536/istfproj.git
cd istfproj
make build
# または
go build -o istfproj
```

## 使い方

```bash
istfproj <directory> [--strict] [--json]
```

### パラメータ

- `<directory>`: 検査対象のディレクトリパス（`.` でカレントディレクトリ）

### オプション

- `--strict`: providerブロックの有無も厳格にチェック（providerがない場合は`false`）
- `--json`: providerブロックの詳細情報をJSON形式で出力

### 使用例

```bash
# デフォルト: *.tfファイルの有無のみチェック
istfproj .
# 終了コード: 0（成功）または 1（失敗）

# Strictモード: providerブロックも必須
istfproj . --strict
# 終了コード: 0（成功）または 1（失敗）

# JSON出力: providerの詳細情報を取得
istfproj . --json
# JSON形式の詳細情報を標準出力に出力

# Strictモード + JSON出力
istfproj . --strict --json
```

## 出力例

### デフォルトモード

```bash
$ istfproj .
$ echo $?
0
```

*.tfファイルが存在すれば終了コード `0`、存在しなければ `1` を返します。
標準出力には何も出力されません。

### Strictモード

```bash
# providerブロックがある場合
$ istfproj . --strict
$ echo $?
0

# *.tfファイルはあるがproviderブロックがない場合
$ istfproj ./module --strict
$ echo $?
1
```

### JSONモード

```bash
$ istfproj . --json
{
  "has_providers": true,
  "has_tf_files": true,
  "is_terraform_project": true,
  "providers": [
    {
      "name": "aws",
      "file": "example.tf",
      "line": 12,
      "attributes": {
        "region": "us-west-2"
      }
    }
  ],
  "providers_count": 1,
  "tf_files_count": 1
}
```

providerブロックがない場合：

```bash
$ istfproj ./module --json
{
  "has_providers": false,
  "has_tf_files": true,
  "is_terraform_project": true,
  "providers": null,
  "providers_count": 0,
  "tf_files_count": 3
}
```

## JSON出力フィールド

- `is_terraform_project` (boolean): Terraformプロジェクトかどうか（*.tfファイルが存在するか）
- `has_tf_files` (boolean): *.tfファイルが存在するか
- `has_providers` (boolean): providerブロックが存在するか
- `tf_files_count` (number): *.tfファイルの数
- `providers_count` (number): 検出されたproviderブロックの数
- `providers` (array): providerブロックの詳細情報の配列
  - `name` (string): provider名（例: "aws", "azurerm", "google"）
  - `file` (string): ファイルパス
  - `line` (number): ファイル内の行番号
  - `attributes` (object): provider設定の属性（region, project等）

## 終了コード

- `0`: 判定成功（Terraformプロジェクトである）
- `1`: 判定失敗（Terraformプロジェクトではない）、またはエラー発生時

**注意**: `--json`オプション使用時も、終了コードは同様に設定されます。

## ユースケース

### 1. CI/CDでのTerraformプロジェクト検証

```bash
#!/bin/bash

if istfproj .; then
    echo "✓ Terraformプロジェクトです"
    terraform init
    terraform plan
else
    echo "✗ Terraformプロジェクトではありません"
    exit 1
fi
```

### 2. moduleとrootプロジェクトの区別

```bash
#!/bin/bash

# Strictモードでproviderブロックの有無をチェック
if istfproj . --strict; then
    echo "Root Terraformプロジェクト（providerあり）"
    terraform init
    terraform apply
else
    if istfproj .; then
        echo "Terraform module（providerなし）"
        # moduleとしての処理
    else
        echo "Terraformプロジェクトではありません"
    fi
fi
```

### 3. provider情報の取得と利用

```bash
#!/bin/bash

result=$(istfproj . --json)

# jqを使用してprovider情報を解析
has_providers=$(echo "$result" | jq -r '.has_providers')

if [ "$has_providers" = "true" ]; then
    # AWSプロバイダーのリージョンを取得
    aws_region=$(echo "$result" | jq -r '.providers[] | select(.name=="aws") | .attributes.region')
    
    if [ -n "$aws_region" ]; then
        echo "AWSリージョン: $aws_region"
        # リージョン固有の処理
    fi
fi
```

### 4. 複数ディレクトリの一括チェック

```bash
#!/bin/bash

for dir in */; do
    if istfproj "$dir"; then
        echo "✓ $dir: Terraformプロジェクト"
        
        # Strictモードでproviderの有無も確認
        if istfproj "$dir" --strict; then
            echo "  → providerブロックあり（root project）"
        else
            echo "  → providerブロックなし（module）"
        fi
    fi
done
```

## 対応するprovider

すべてのTerraform providerに対応しています：

- `aws` - Amazon Web Services
- `azurerm` - Microsoft Azure
- `google` - Google Cloud Platform
- `kubernetes` - Kubernetes
- `helm` - Helm
- `docker` - Docker
- その他すべてのTerraform provider

## 開発

### Makefileコマンド

```bash
# ビルド
make build

# テスト実行
make test

# コードフォーマット
make fmt

# コード検証
make vet

# 依存関係の整理
make tidy

# スナップショットビルド（現在のプラットフォームのみ）
make snapshot

# スナップショットビルド（全プラットフォーム）
make snapshot-all

# GoReleaser設定の確認
make check

# クリーンアップ
make clean

# すべてのコマンドを表示
make help
```

## 依存関係

- `github.com/hashicorp/hcl/v2` - HCL v2パーサー
- `github.com/hashicorp/hcl/v2/hclparse` - HCLファイルパーサー
- `github.com/hashicorp/hcl/v2/hclsyntax` - HCL構文解析
- `github.com/zclconf/go-cty/cty` - HCL値の型システム

## ライセンス

このツールはMITライセンスの下で公開されています。
