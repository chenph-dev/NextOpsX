package i18n

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestSourceKeysRegistered parses the UI packages and checks that every
// i18n.T / i18n.Tf key used in production code exists in both catalogs.
// It catches typos and forgotten translations before they reach the screen.
func TestSourceKeysRegistered(t *testing.T) {
	roots := []string{
		filepath.Join("..", "ui", "views"),
		filepath.Join("..", "ui", "workspace"),
		filepath.Join("..", "ui", "widgets"),
		filepath.Join("..", "cmd", "nextopsx", "main.go"),
	}
	fset := token.NewFileSet()
	checked := 0
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			file, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return err
			}
			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok || len(call.Args) == 0 {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				pkg, ok := sel.X.(*ast.Ident)
				if !ok || pkg.Name != "i18n" || (sel.Sel.Name != "T" && sel.Sel.Name != "Tf") {
					return true
				}
				lit, ok := call.Args[0].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				key, err := strconv.Unquote(lit.Value)
				if err != nil {
					return true
				}
				checked++
				pos := fset.Position(lit.Pos())
				if !Has(ZhCN, key) {
					t.Errorf("%s: key %q missing from zh-CN", pos, key)
				}
				if !Has(EnUS, key) {
					t.Errorf("%s: key %q missing from en-US", pos, key)
				}
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
	if checked == 0 {
		t.Fatal("no i18n.T/Tf calls found; key check did not run")
	}
}
