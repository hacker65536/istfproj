# Examples Directory

このディレクトリには、`istfproj` ツールをテストするための様々なモックTerraformプロジェクトが含まれています。

## ディレクトリ構成

### 1. simple-aws/
シンプルなAWSプロバイダーを使用した基本的なTerraformプロジェクト。
- **プロバイダー**: AWS (1つ)
- **リソース**: S3バケット
- **用途**: 基本的な動作確認

**テスト例:**
```bash
# 基本チェック（.tfファイルの存在確認）
./istfproj examples/simple-aws

# 厳格モード（providerブロックも確認）
./istfproj examples/simple-aws --strict

# JSON出力
./istfproj examples/simple-aws --json
```

### 2. no-provider/
providerブロックが定義されていないTerraformプロジェクト。
- **プロバイダー**: なし（required_providersのみ）
- **リソース**: S3バケット
- **用途**: `--strict` モードのテスト

**テスト例:**
```bash
# 基本チェック（成功: .tfファイルが存在）
./istfproj examples/no-provider

# 厳格モード（失敗: providerブロックが存在しない）
./istfproj examples/no-provider --strict
```

### 3. multi-provider/
複数のクラウドプロバイダーを使用するマルチクラウドプロジェクト。
- **プロバイダー**: AWS, Google Cloud, Azure (3つ)
- **リソース**: S3バケット、GCSバケット、Azureリソースグループ
- **用途**: 複数プロバイダーの検出テスト

**テスト例:**
```bash
# JSON出力で全プロバイダー情報を確認
./istfproj examples/multi-provider --json

# 厳格モード（成功: 複数のproviderブロックが存在）
./istfproj examples/multi-provider --strict
```

### 4. complex-config/
複雑な設定を持つAWSプロバイダープロジェクト。
- **プロバイダー**: AWS (3つ - デフォルト、tokyo、backup)
- **特徴**: 
  - 変数参照
  - assume_role設定
  - default_tags
  - プロバイダーエイリアス
- **用途**: 複雑な属性の抽出テスト

**テスト例:**
```bash
# JSON出力で詳細な属性情報を確認
./istfproj examples/complex-config --json
```

### 5. gcp-project/
Google Cloud Platform専用のプロジェクト。
- **プロバイダー**: Google Cloud (1つ)
- **リソース**: VPC、サブネット、Compute Instance
- **用途**: GCPプロバイダーの検出テスト

**テスト例:**
```bash
./istfproj examples/gcp-project --json
./istfproj examples/gcp-project --strict
```

### 6. azure-project/
Microsoft Azure専用のプロジェクト。
- **プロバイダー**: Azure (1つ)
- **リソース**: リソースグループ、VNet、ストレージアカウント
- **用途**: Azureプロバイダーの検出テスト

**テスト例:**
```bash
./istfproj examples/azure-project --json
./istfproj examples/azure-project --strict
```

### 7. empty-dir/
.tfファイルが存在しない空のディレクトリ。
- **プロバイダー**: なし
- **リソース**: なし
- **用途**: エラーハンドリングのテスト

**テスト例:**
```bash
# 失敗: .tfファイルが存在しない
./istfproj examples/empty-dir

# 終了コードは1
echo $?
```

## テストシナリオ

### シナリオ1: 基本的な動作確認
```bash
# すべてのサンプルディレクトリをチェック
for dir in examples/*/; do
  echo "Testing: $dir"
  ./istfproj "$dir"
  echo "Exit code: $?"
  echo "---"
done
```

### シナリオ2: 厳格モードのテスト
```bash
# providerブロックの有無を確認
./istfproj examples/simple-aws --strict      # 成功
./istfproj examples/no-provider --strict     # 失敗
./istfproj examples/multi-provider --strict  # 成功
./istfproj examples/empty-dir --strict       # 失敗
```

### シナリオ3: JSON出力の確認
```bash
# 各プロジェクトのプロバイダー情報をJSON形式で取得
./istfproj examples/simple-aws --json | jq .
./istfproj examples/multi-provider --json | jq '.providers[] | {name, file, line}'
./istfproj examples/complex-config --json | jq '.providers[] | {name, attributes}'
```

### シナリオ4: プロバイダー数のカウント
```bash
# 各プロジェクトのプロバイダー数を確認
for dir in examples/*/; do
  if [ -f "$dir/main.tf" ]; then
    count=$(./istfproj "$dir" --json 2>/dev/null | jq -r '.providers_count // 0')
    echo "$dir: $count providers"
  fi
done
```

## 期待される結果

| ディレクトリ | 基本モード | --strict | プロバイダー数 |
|------------|----------|----------|--------------|
| simple-aws | ✅ (0) | ✅ (0) | 1 |
| no-provider | ✅ (0) | ❌ (1) | 0 |
| multi-provider | ✅ (0) | ✅ (0) | 3 |
| complex-config | ✅ (0) | ✅ (0) | 3 |
| gcp-project | ✅ (0) | ✅ (0) | 1 |
| azure-project | ✅ (0) | ✅ (0) | 1 |
| empty-dir | ❌ (1) | ❌ (1) | 0 |

注: ✅ = 成功（終了コード0）、❌ = 失敗（終了コード1）

## 追加のテストケース

必要に応じて、以下のようなテストケースを追加できます：

- **invalid-syntax/**: 構文エラーのある.tfファイル
- **nested-modules/**: モジュールを使用したプロジェクト
- **multiple-files/**: 複数の.tfファイルに分割されたプロジェクト
- **backend-config/**: バックエンド設定を含むプロジェクト
