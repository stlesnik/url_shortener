package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var NoExitAnalyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "forbid direct calls to os.Exit in main.main function",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" {
				continue
			}
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				expr, ok := node.(*ast.ExprStmt)
				if !ok {
					return true
				}
				call, ok := expr.X.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				pkgIdent, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}
				if pkgIdent.Name == "os" && sel.Sel.Name == "Exit" {
					pass.Reportf(sel.Sel.Pos(), "direct call to os.Exit is forbidden in main.main")
				}
				return true
			})
			break
		}
	}
	return nil, nil
}
