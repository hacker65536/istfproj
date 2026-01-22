package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// ProviderInfo はproviderブロックの情報を保持します
type ProviderInfo struct {
	Name       string                 `json:"name"`
	File       string                 `json:"file"`
	Line       int                    `json:"line"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// ProviderChecker はTerraformファイル内のproviderブロックをチェックするツールです
type ProviderChecker struct {
	parser *hclparse.Parser
}

// NewProviderChecker は新しいProviderCheckerを作成します
func NewProviderChecker() *ProviderChecker {
	return &ProviderChecker{
		parser: hclparse.NewParser(),
	}
}

// GetProviderInfoFromFile は指定されたファイルから特定のproviderブロックの情報を取得します
func (pc *ProviderChecker) GetProviderInfoFromFile(filename, providerName string) (*ProviderInfo, error) {
	file, diags := pc.parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, fmt.Errorf("parse error in %s: %s", filename, diags.Error())
	}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil, fmt.Errorf("unexpected body type in %s", filename)
	}

	return pc.findProviderInBlocks(body.Blocks, providerName, filename), nil
}

// findProviderInBlocks はブロックリスト内から指定されたproviderを検索し、情報を返します
func (pc *ProviderChecker) findProviderInBlocks(blocks hclsyntax.Blocks, providerName, filename string) *ProviderInfo {
	for _, block := range blocks {
		if block.Type == "provider" && len(block.Labels) > 0 {
			if block.Labels[0] == providerName {
				info := &ProviderInfo{
					Name:       providerName,
					File:       filename,
					Line:       block.DefRange().Start.Line,
					Attributes: make(map[string]interface{}),
				}

				// 属性を抽出
				body := block.Body
				for name, attr := range body.Attributes {
					// 属性値を取得（簡易的な実装）
					value := pc.extractAttributeValue(attr)
					if value != nil {
						info.Attributes[name] = value
					}
				}

				return info
			}
		}
	}
	return nil
}

// extractAttributeValue は属性値を抽出します
func (pc *ProviderChecker) extractAttributeValue(attr *hclsyntax.Attribute) interface{} {
	// 式の種類に応じて値を抽出
	switch expr := attr.Expr.(type) {
	case *hclsyntax.LiteralValueExpr:
		// リテラル値（文字列、数値、ブール値など）
		val := expr.Val
		return ctyValueToInterface(val)
	case *hclsyntax.TemplateExpr:
		// テンプレート式（文字列補間など）
		if len(expr.Parts) == 1 {
			if lit, ok := expr.Parts[0].(*hclsyntax.LiteralValueExpr); ok {
				return ctyValueToInterface(lit.Val)
			}
		}
		return "<template>"
	case *hclsyntax.ScopeTraversalExpr:
		// 変数参照など
		return fmt.Sprintf("${%s}", traversalToString(expr.Traversal))
	case *hclsyntax.ObjectConsExpr:
		// オブジェクト
		obj := make(map[string]interface{})
		for _, item := range expr.Items {
			if keyExpr, ok := item.KeyExpr.(*hclsyntax.ObjectConsKeyExpr); ok {
				if !keyExpr.ForceNonLiteral {
					if lit, ok := keyExpr.Wrapped.(*hclsyntax.LiteralValueExpr); ok {
						key := ctyValueToInterface(lit.Val)
						if keyStr, ok := key.(string); ok {
							obj[keyStr] = pc.extractExprValue(item.ValueExpr)
						}
					}
				}
			}
		}
		return obj
	case *hclsyntax.TupleConsExpr:
		// 配列
		arr := make([]interface{}, 0, len(expr.Exprs))
		for _, e := range expr.Exprs {
			arr = append(arr, pc.extractExprValue(e))
		}
		return arr
	default:
		return "<complex>"
	}
}

// extractExprValue は式から値を抽出します
func (pc *ProviderChecker) extractExprValue(expr hclsyntax.Expression) interface{} {
	switch e := expr.(type) {
	case *hclsyntax.LiteralValueExpr:
		return ctyValueToInterface(e.Val)
	case *hclsyntax.TemplateExpr:
		if len(e.Parts) == 1 {
			if lit, ok := e.Parts[0].(*hclsyntax.LiteralValueExpr); ok {
				return ctyValueToInterface(lit.Val)
			}
		}
		return "<template>"
	case *hclsyntax.ScopeTraversalExpr:
		return fmt.Sprintf("${%s}", traversalToString(e.Traversal))
	default:
		return "<complex>"
	}
}

// ctyValueToInterface はcty.ValueをGo標準の型に変換します
func ctyValueToInterface(val cty.Value) interface{} {
	if val.IsNull() {
		return nil
	}

	valType := val.Type()
	switch {
	case valType == cty.String:
		var s string
		gocty.FromCtyValue(val, &s)
		return s
	case valType == cty.Number:
		var f float64
		gocty.FromCtyValue(val, &f)
		// 整数かどうかチェック
		if f == float64(int64(f)) {
			return int64(f)
		}
		return f
	case valType == cty.Bool:
		var b bool
		gocty.FromCtyValue(val, &b)
		return b
	default:
		return val.GoString()
	}
}

// traversalToString はTraversalを文字列に変換します
func traversalToString(traversal hcl.Traversal) string {
	parts := make([]string, 0, len(traversal))
	for i, t := range traversal {
		switch tt := t.(type) {
		case hcl.TraverseRoot:
			parts = append(parts, tt.Name)
		case hcl.TraverseAttr:
			if i > 0 {
				parts = append(parts, "."+tt.Name)
			} else {
				parts = append(parts, tt.Name)
			}
		}
	}
	return strings.Join(parts, "")
}

// CheckProviderInDirectory はディレクトリ内のすべての*.tfファイルをチェックします
func (pc *ProviderChecker) CheckProviderInDirectory(dir, providerName string) ([]*ProviderInfo, error) {
	var providers []*ProviderInfo

	// ディレクトリ内の*.tfファイルを検索
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// ディレクトリはスキップ
		if info.IsDir() {
			return nil
		}

		// .tfファイルのみ処理
		if !strings.HasSuffix(path, ".tf") {
			return nil
		}

		// ファイル内にproviderブロックが存在するかチェック
		providerInfo, err := pc.GetProviderInfoFromFile(path, providerName)
		if err != nil {
			// エラーは警告として表示するが、処理は継続
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			return nil
		}

		if providerInfo != nil {
			providers = append(providers, providerInfo)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %v", err)
	}

	return providers, nil
}

// hasTerraformFiles はディレクトリ内に*.tfファイルが存在するかチェックします
func hasTerraformFiles(dir string) (bool, []string, error) {
	var tfFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".tf") {
			tfFiles = append(tfFiles, path)
		}

		return nil
	})

	if err != nil {
		return false, nil, err
	}

	return len(tfFiles) > 0, tfFiles, nil
}

// getAllProviders はディレクトリ内のすべてのproviderブロックを取得します
func (pc *ProviderChecker) getAllProviders(dir string) ([]*ProviderInfo, error) {
	var allProviders []*ProviderInfo

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".tf") {
			return nil
		}

		// ファイルを解析
		file, diags := pc.parser.ParseHCLFile(path)
		if diags.HasErrors() {
			fmt.Fprintf(os.Stderr, "Warning: parse error in %s: %s\n", path, diags.Error())
			return nil
		}

		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			return nil
		}

		// すべてのproviderブロックを検索
		for _, block := range body.Blocks {
			if block.Type == "provider" && len(block.Labels) > 0 {
				info := &ProviderInfo{
					Name:       block.Labels[0],
					File:       path,
					Line:       block.DefRange().Start.Line,
					Attributes: make(map[string]interface{}),
				}

				// 属性を抽出
				blockBody := block.Body
				for name, attr := range blockBody.Attributes {
					value := pc.extractAttributeValue(attr)
					if value != nil {
						info.Attributes[name] = value
					}
				}

				allProviders = append(allProviders, info)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return allProviders, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: istfproj <directory> [--strict] [--json]")
		fmt.Println("\nDescription:")
		fmt.Println("  指定されたディレクトリがTerraformプロジェクトかどうかを判定します")
		fmt.Println("\nOptions:")
		fmt.Println("  --strict  providerブロックの有無も厳格にチェック（providerがない場合はfalse）")
		fmt.Println("  --json    providerブロックの詳細情報をJSON形式で出力")
		fmt.Println("\nExamples:")
		fmt.Println("  istfproj .              # カレントディレクトリをチェック（*.tfの有無のみ）")
		fmt.Println("  istfproj . --strict     # providerブロックも必須としてチェック")
		fmt.Println("  istfproj . --json       # providerの詳細情報をJSON出力")
		fmt.Println("  istfproj . --strict --json  # 厳格チェック + JSON出力")
		os.Exit(1)
	}

	directory := os.Args[1]
	strictMode := false
	jsonOutput := false

	// オプションの解析
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--strict":
			strictMode = true
		case "--json":
			jsonOutput = true
		}
	}

	// ディレクトリの存在確認
	info, err := os.Stat(directory)
	if os.IsNotExist(err) {
		if !jsonOutput {
			fmt.Fprintf(os.Stderr, "Error: Directory '%s' does not exist\n", directory)
		}
		os.Exit(1)
	}
	if !info.IsDir() {
		if !jsonOutput {
			fmt.Fprintf(os.Stderr, "Error: '%s' is not a directory\n", directory)
		}
		os.Exit(1)
	}

	// *.tfファイルの存在確認
	hasTfFiles, tfFiles, err := hasTerraformFiles(directory)
	if err != nil {
		if !jsonOutput {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}

	// *.tfファイルがない場合
	if !hasTfFiles {
		os.Exit(1)
	}

	checker := NewProviderChecker()

	// すべてのproviderブロックを取得
	providers, err := checker.getAllProviders(directory)
	if err != nil {
		if !jsonOutput {
			fmt.Fprintf(os.Stderr, "Warning: Error parsing files: %v\n", err)
		}
		// パースエラーがあっても*.tfファイルは存在するので、strictモードでなければ成功
		if strictMode {
			os.Exit(1)
		}
		os.Exit(0)
	}

	// JSON出力モード
	if jsonOutput {
		output := map[string]interface{}{
			"is_terraform_project": true,
			"has_tf_files":         true,
			"has_providers":        len(providers) > 0,
			"tf_files_count":       len(tfFiles),
			"providers_count":      len(providers),
			"providers":            providers,
		}

		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to generate JSON: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(string(jsonData))
		os.Exit(0)
	}

	// デフォルトモード: *.tfファイルがあれば成功（終了コード0）
	if !strictMode {
		os.Exit(0)
	}

	// strictモード: providerブロックも必要
	if len(providers) > 0 {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
