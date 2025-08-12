package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"

	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/printf"

	"honnef.co/go/tools/staticcheck"

	"github.com/gostaticanalysis/forcetypeassert"
	"github.com/gostaticanalysis/nilerr"
)

func main() {
	analyzers := []*analysis.Analyzer{
		inspect.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
	}

	checks := map[string]bool{
		"ST1000": true,
		"S1000":  true,
		"QF1001": true,
	}
	for _, v := range staticcheck.Analyzers {
		if v.Analyzer.Name[:2] == "SA" || checks[v.Analyzer.Name] {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	analyzers = append(analyzers,
		forcetypeassert.Analyzer,
		nilerr.Analyzer,
	)

	analyzers = append(analyzers, NoExitAnalyzer)

	multichecker.Main(analyzers...)
}
