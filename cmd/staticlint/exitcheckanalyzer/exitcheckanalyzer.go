package exitcheckanalyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// ExitCheckAnalyzer определяет анализатор, запрещающий прямой вызов os.Exit в функции main пакета main.
var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for direct os.Exit calls in main function",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// Проверяем только main пакеты
		if file.Name.Name != "main" {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			if funcDecl, ok := node.(*ast.FuncDecl); ok {
				// Проверяем только функцию main
				if funcDecl.Name.Name != "main" {
					return true
				}

				ast.Inspect(funcDecl, func(n ast.Node) bool {
					if callExpr, ok := n.(*ast.CallExpr); ok {
						if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
							if ident, ok := selExpr.X.(*ast.Ident); ok {
								if ident.Name == "os" && selExpr.Sel.Name == "Exit" {
									pass.Reportf(callExpr.Pos(), "direct call to os.Exit in main function is forbidden")
								}
							}
						}
					}
					return true
				})
			}
			return true
		})
	}
	return nil, nil
}
