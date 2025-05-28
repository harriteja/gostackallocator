package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

// usageTracker tracks allocation sites and their usage patterns
type usageTracker struct {
	allocSites map[types.Object]token.Pos
	useCounts  map[types.Object]int
	escapes    map[types.Object]bool
}

// newUsageTracker creates a new usage tracker
func newUsageTracker() *usageTracker {
	return &usageTracker{
		allocSites: make(map[types.Object]token.Pos),
		useCounts:  make(map[types.Object]int),
		escapes:    make(map[types.Object]bool),
	}
}

// InspectFile walks the AST and detects allocation patterns
func InspectFile(f *ast.File, info *types.Info, fset *token.FileSet, report func(pos token.Pos, msg string)) {
	tracker := newUsageTracker()

	// Create pattern detector with config (we'll need to add config parameter later)
	config := &Config{} // Default config for now
	detector := NewPatternDetector(info, fset, config, tracker)

	// First pass: collect allocation sites and usage counts using enhanced pattern detection
	ast.Inspect(f, func(n ast.Node) bool {
		// Use the new pattern detector for comprehensive analysis
		detector.DetectPattern(n, report)

		// Keep existing logic for compatibility
		switch expr := n.(type) {
		case *ast.UnaryExpr:
			if expr.Op == token.AND {
				if ident, ok := expr.X.(*ast.Ident); ok {
					if obj := info.ObjectOf(ident); obj != nil && isLocalVar(obj) {
						tracker.allocSites[obj] = expr.Pos()
						tracker.useCounts[obj]++
					}
				}
			}
		case *ast.CallExpr:
			// Enhanced new() call detection is now handled by PatternDetector
			// Keep usage counting for tracked objects
			for _, arg := range expr.Args {
				if ident, ok := arg.(*ast.Ident); ok {
					if obj := info.ObjectOf(ident); obj != nil {
						if _, exists := tracker.allocSites[obj]; exists {
							tracker.useCounts[obj]++
						}
					}
				}
			}
		case *ast.Ident:
			// Count general usage
			if obj := info.ObjectOf(expr); obj != nil {
				if _, exists := tracker.allocSites[obj]; exists {
					tracker.useCounts[obj]++
				}
			}
		case *ast.ReturnStmt:
			// Check for escaping allocations in return statements
			for _, res := range expr.Results {
				checkEscapingAllocation(res, info, tracker, report)
			}
		case *ast.AssignStmt:
			// Check for escaping allocations in assignments
			for _, rhs := range expr.Rhs {
				checkEscapingAllocation(rhs, info, tracker, report)
			}
		}
		return true
	})

	// Second pass: report single-use escaping allocations
	for obj, pos := range tracker.allocSites {
		if tracker.useCounts[obj] <= 1 && tracker.escapes[obj] {
			report(pos, fmt.Sprintf("pointer to %s escapes only once; consider using stack allocation", obj.Name()))
		}
	}
}

// checkEscapingAllocation checks if an expression contains escaping allocations
func checkEscapingAllocation(expr ast.Expr, info *types.Info, tracker *usageTracker, report func(pos token.Pos, msg string)) {
	switch e := expr.(type) {
	case *ast.UnaryExpr:
		if e.Op == token.AND {
			if ident, ok := e.X.(*ast.Ident); ok {
				if obj := info.ObjectOf(ident); obj != nil {
					if _, exists := tracker.allocSites[obj]; exists {
						tracker.escapes[obj] = true
					}
				}
			}
		}
	case *ast.CallExpr:
		// Check if this is a new() call in return/assignment
		if isNewCall(e, info) {
			report(e.Pos(), "new(T) in return/assignment always allocates on heap; consider stack allocation")
		}
	}
}

// isLocalVar checks if an object is a local variable
func isLocalVar(obj types.Object) bool {
	if obj == nil {
		return false
	}
	// Check if it's a variable and not a package-level declaration
	if v, ok := obj.(*types.Var); ok {
		return v.Parent() != v.Pkg().Scope()
	}
	return false
}

// isNewCall checks if a call expression is a call to new()
func isNewCall(call *ast.CallExpr, info *types.Info) bool {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		if obj := info.ObjectOf(ident); obj != nil {
			if builtin, ok := obj.(*types.Builtin); ok {
				return builtin.Name() == "new"
			}
		}
	}
	return false
}

// GetCodeSnippet extracts a code snippet around the given position
func GetCodeSnippet(fset *token.FileSet, pos token.Pos, src []byte) string {
	position := fset.Position(pos)
	if position.Line <= 0 || len(src) == 0 {
		return ""
	}

	lines := splitLines(src)
	if position.Line > len(lines) {
		return ""
	}

	// Get context around the line (±2 lines)
	start := max(0, position.Line-3)
	end := min(len(lines), position.Line+2)

	snippet := ""
	for i := start; i < end; i++ {
		if i == position.Line-1 {
			snippet += fmt.Sprintf(">>> %s\n", lines[i])
		} else {
			snippet += fmt.Sprintf("    %s\n", lines[i])
		}
	}

	return snippet
}

// splitLines splits source code into lines
func splitLines(src []byte) []string {
	var lines []string
	var line []byte

	for _, b := range src {
		if b == '\n' {
			lines = append(lines, string(line))
			line = line[:0]
		} else {
			line = append(line, b)
		}
	}

	if len(line) > 0 {
		lines = append(lines, string(line))
	}

	return lines
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
