package analyzer

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"
)

// AllocationPattern represents different types of allocation patterns
type AllocationPattern int

const (
	PatternNewCall AllocationPattern = iota
	PatternMakeSlice
	PatternMakeMap
	PatternMakeChan
	PatternSliceLiteral
	PatternMapLiteral
	PatternStructLiteral
	PatternInterfaceConversion
	PatternStringConcat
	PatternAppendGrowth
	PatternClosureCapture
	PatternReflectNew
	PatternBoxing
)

// PatternDetector detects various allocation patterns
type PatternDetector struct {
	info    *types.Info
	fset    *token.FileSet
	config  *Config
	tracker *usageTracker
}

// NewPatternDetector creates a new pattern detector
func NewPatternDetector(info *types.Info, fset *token.FileSet, config *Config, tracker *usageTracker) *PatternDetector {
	return &PatternDetector{
		info:    info,
		fset:    fset,
		config:  config,
		tracker: tracker,
	}
}

// DetectPattern analyzes a node and detects allocation patterns
func (pd *PatternDetector) DetectPattern(node ast.Node, report func(pos token.Pos, msg string)) {
	switch n := node.(type) {
	case *ast.CallExpr:
		pd.detectCallPatterns(n, report)
	case *ast.CompositeLit:
		pd.detectCompositeLiteralPatterns(n, report)
	case *ast.BinaryExpr:
		pd.detectBinaryExprPatterns(n, report)
	case *ast.TypeAssertExpr:
		pd.detectTypeAssertionPatterns(n, report)
	case *ast.FuncLit:
		pd.detectClosurePatterns(n, report)
	}
}

// detectCallPatterns detects allocation patterns in function calls
func (pd *PatternDetector) detectCallPatterns(call *ast.CallExpr, report func(pos token.Pos, msg string)) {
	// new(T) calls
	if pd.isNewCall(call) {
		report(call.Pos(), "new(T) always allocates on heap; consider using stack allocation if object doesn't escape")
		return
	}

	// make() calls
	if pd.isMakeCall(call) {
		pd.detectMakePatterns(call, report)
		return
	}

	// append() calls that may cause growth
	if pd.isAppendCall(call) {
		pd.detectAppendPatterns(call, report)
		return
	}

	// reflect.New() and similar reflection calls
	if pd.isReflectAllocation(call) {
		report(call.Pos(), "reflection-based allocation always uses heap; consider avoiding if performance critical")
		return
	}

	// String formatting functions that allocate
	if pd.isStringFormattingCall(call) {
		pd.detectStringFormattingPatterns(call, report)
		return
	}

	// Interface method calls that may box values
	if pd.isBoxingCall(call) {
		report(call.Pos(), "value may be boxed when passed to interface; consider using pointer receiver if appropriate")
		return
	}
}

// detectMakePatterns detects patterns in make() calls
func (pd *PatternDetector) detectMakePatterns(call *ast.CallExpr, report func(pos token.Pos, msg string)) {
	if len(call.Args) == 0 {
		return
	}

	// Get the type being made
	typeExpr := call.Args[0]

	switch pd.getTypeKind(typeExpr) {
	case "slice":
		if len(call.Args) >= 2 {
			// make([]T, size) or make([]T, size, capacity)
			if pd.isSmallConstantSize(call.Args[1]) {
				report(call.Pos(), "small slice allocation with make(); consider using array or stack allocation")
			} else if pd.isLargeSize(call.Args[1]) {
				report(call.Pos(), "large slice allocation may cause GC pressure; consider pre-allocation or streaming")
			}
		} else {
			report(call.Pos(), "make([]T) creates zero-length slice; consider using nil slice or array")
		}

	case "map":
		if len(call.Args) >= 2 {
			if pd.isSmallConstantSize(call.Args[1]) {
				report(call.Pos(), "small map with known size; consider using struct or array for better performance")
			}
		} else {
			report(call.Pos(), "make(map[K]V) without size hint; consider providing capacity for better performance")
		}

	case "chan":
		if len(call.Args) >= 2 {
			if pd.isZeroOrSmallSize(call.Args[1]) {
				report(call.Pos(), "unbuffered or small buffered channel; consider if synchronous communication is needed")
			}
		}
	}
}

// detectCompositeLiteralPatterns detects patterns in composite literals
func (pd *PatternDetector) detectCompositeLiteralPatterns(lit *ast.CompositeLit, report func(pos token.Pos, msg string)) {
	switch pd.getCompositeLiteralType(lit) {
	case "slice":
		if pd.isSmallSliceLiteral(lit) {
			report(lit.Pos(), "small slice literal; consider using array for stack allocation")
		}
		if pd.hasComplexElements(lit) {
			report(lit.Pos(), "slice literal with complex elements may cause multiple allocations")
		}

	case "map":
		if pd.isSmallMapLiteral(lit) {
			report(lit.Pos(), "small map literal; consider using struct or switch statement for better performance")
		}

	case "struct":
		if pd.isLargeStructLiteral(lit) {
			report(lit.Pos(), "large struct literal; consider using pointer or breaking into smaller structs")
		}
		if pd.hasEscapingStructLiteral(lit) {
			report(lit.Pos(), "struct literal address taken; consider stack allocation if lifetime allows")
		}
	}
}

// detectBinaryExprPatterns detects allocation patterns in binary expressions
func (pd *PatternDetector) detectBinaryExprPatterns(expr *ast.BinaryExpr, report func(pos token.Pos, msg string)) {
	if expr.Op == token.ADD {
		// String concatenation
		if pd.isStringType(expr.X) && pd.isStringType(expr.Y) {
			report(expr.Pos(), "string concatenation with + operator allocates; consider using strings.Builder for multiple concatenations")
		}
	}
}

// detectTypeAssertionPatterns detects allocation patterns in type assertions
func (pd *PatternDetector) detectTypeAssertionPatterns(assert *ast.TypeAssertExpr, report func(pos token.Pos, msg string)) {
	if pd.isInterfaceToConcreteAssertion(assert) {
		report(assert.Pos(), "type assertion may cause allocation if value was boxed; consider avoiding interface{} when possible")
	}
}

// detectAppendPatterns detects allocation patterns in append calls
func (pd *PatternDetector) detectAppendPatterns(call *ast.CallExpr, report func(pos token.Pos, msg string)) {
	if len(call.Args) < 2 {
		return
	}

	// Check if appending to nil or small slice
	if pd.isNilSlice(call.Args[0]) {
		report(call.Pos(), "appending to nil slice causes allocation; consider pre-allocating with make()")
	}

	// Check if appending many elements at once
	if len(call.Args) > 3 {
		report(call.Pos(), "appending multiple elements may cause multiple reallocations; consider pre-allocating capacity")
	}

	// Check for append in loop (common performance issue)
	if pd.isInLoop(call) {
		report(call.Pos(), "append in loop may cause multiple reallocations; consider pre-allocating slice capacity")
	}
}

// detectClosurePatterns detects allocation patterns in closures
func (pd *PatternDetector) detectClosurePatterns(fn *ast.FuncLit, report func(pos token.Pos, msg string)) {
	// Check if closure captures variables (may cause allocation)
	if pd.capturesVariables(fn) {
		report(fn.Pos(), "closure captures variables and may allocate; consider passing values as parameters")
	}

	// Check if closure is assigned to interface
	if pd.isClosureToInterface(fn) {
		report(fn.Pos(), "closure assigned to interface causes allocation; consider using concrete function type")
	}
}

// detectStringFormattingPatterns detects allocation patterns in string formatting
func (pd *PatternDetector) detectStringFormattingPatterns(call *ast.CallExpr, report func(pos token.Pos, msg string)) {
	funcName := pd.getFunctionName(call)

	switch funcName {
	case "fmt.Sprintf", "fmt.Errorf":
		if pd.isSimpleStringFormatting(call) {
			report(call.Pos(), "simple string formatting; consider using string concatenation or strings.Builder")
		}
	case "fmt.Sprint", "fmt.Sprintln":
		report(call.Pos(), "fmt.Sprint family functions allocate; consider using strings.Builder or direct conversion")
	case "strconv.Itoa":
		if pd.isInHotPath(call) {
			report(call.Pos(), "strconv.Itoa allocates; consider using strconv.AppendInt with pre-allocated buffer")
		}
	}
}

// Helper methods for pattern detection

func (pd *PatternDetector) isNewCall(call *ast.CallExpr) bool {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		if obj := pd.info.ObjectOf(ident); obj != nil {
			if builtin, ok := obj.(*types.Builtin); ok {
				return builtin.Name() == "new"
			}
		}
	}
	return false
}

func (pd *PatternDetector) isMakeCall(call *ast.CallExpr) bool {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		if obj := pd.info.ObjectOf(ident); obj != nil {
			if builtin, ok := obj.(*types.Builtin); ok {
				return builtin.Name() == "make"
			}
		}
	}
	return false
}

func (pd *PatternDetector) isAppendCall(call *ast.CallExpr) bool {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		if obj := pd.info.ObjectOf(ident); obj != nil {
			if builtin, ok := obj.(*types.Builtin); ok {
				return builtin.Name() == "append"
			}
		}
	}
	return false
}

func (pd *PatternDetector) isReflectAllocation(call *ast.CallExpr) bool {
	funcName := pd.getFunctionName(call)
	return strings.HasPrefix(funcName, "reflect.New") ||
		strings.HasPrefix(funcName, "reflect.MakeSlice") ||
		strings.HasPrefix(funcName, "reflect.MakeMap") ||
		strings.HasPrefix(funcName, "reflect.MakeChan")
}

func (pd *PatternDetector) isStringFormattingCall(call *ast.CallExpr) bool {
	funcName := pd.getFunctionName(call)
	return strings.HasPrefix(funcName, "fmt.") ||
		strings.HasPrefix(funcName, "strconv.")
}

func (pd *PatternDetector) isBoxingCall(call *ast.CallExpr) bool {
	// Check if passing value type to interface parameter
	if len(call.Args) == 0 {
		return false
	}

	// This is a simplified check - in practice, you'd need more sophisticated analysis
	for _, arg := range call.Args {
		if pd.isValueTypeToInterface(arg) {
			return true
		}
	}
	return false
}

func (pd *PatternDetector) getFunctionName(call *ast.CallExpr) string {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		return fun.Name
	case *ast.SelectorExpr:
		if pkg, ok := fun.X.(*ast.Ident); ok {
			return pkg.Name + "." + fun.Sel.Name
		}
		return fun.Sel.Name
	}
	return ""
}

func (pd *PatternDetector) getTypeKind(expr ast.Expr) string {
	if t := pd.info.TypeOf(expr); t != nil {
		switch t.Underlying().(type) {
		case *types.Slice:
			return "slice"
		case *types.Map:
			return "map"
		case *types.Chan:
			return "chan"
		}
	}
	return "unknown"
}

func (pd *PatternDetector) getCompositeLiteralType(lit *ast.CompositeLit) string {
	if t := pd.info.TypeOf(lit); t != nil {
		switch t.Underlying().(type) {
		case *types.Slice:
			return "slice"
		case *types.Map:
			return "map"
		case *types.Struct:
			return "struct"
		case *types.Array:
			return "array"
		}
	}
	return "unknown"
}

func (pd *PatternDetector) isSmallConstantSize(expr ast.Expr) bool {
	if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.INT {
		// Simple heuristic: consider sizes <= 32 as small
		return len(lit.Value) <= 2 // "32" or smaller
	}
	return false
}

func (pd *PatternDetector) isLargeSize(expr ast.Expr) bool {
	if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.INT {
		// Consider sizes > 1000 as large
		return len(lit.Value) >= 4 // "1000" or larger
	}
	return false
}

func (pd *PatternDetector) isZeroOrSmallSize(expr ast.Expr) bool {
	if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.INT {
		return lit.Value == "0" || pd.isSmallConstantSize(expr)
	}
	return false
}

func (pd *PatternDetector) isSmallSliceLiteral(lit *ast.CompositeLit) bool {
	return len(lit.Elts) <= 4 && len(lit.Elts) > 0
}

func (pd *PatternDetector) isSmallMapLiteral(lit *ast.CompositeLit) bool {
	return len(lit.Elts) <= 3 && len(lit.Elts) > 0
}

func (pd *PatternDetector) isLargeStructLiteral(lit *ast.CompositeLit) bool {
	return len(lit.Elts) > 10
}

func (pd *PatternDetector) hasComplexElements(lit *ast.CompositeLit) bool {
	for _, elt := range lit.Elts {
		if _, ok := elt.(*ast.CompositeLit); ok {
			return true // Nested composite literal
		}
		if _, ok := elt.(*ast.CallExpr); ok {
			return true // Function call in element
		}
	}
	return false
}

func (pd *PatternDetector) hasEscapingStructLiteral(lit *ast.CompositeLit) bool {
	// This would need more sophisticated escape analysis
	// For now, just check if it's in a return statement or assignment to interface
	return false // Simplified
}

func (pd *PatternDetector) isStringType(expr ast.Expr) bool {
	if t := pd.info.TypeOf(expr); t != nil {
		if basic, ok := t.Underlying().(*types.Basic); ok {
			return basic.Kind() == types.String
		}
	}
	return false
}

func (pd *PatternDetector) isInterfaceToConcreteAssertion(assert *ast.TypeAssertExpr) bool {
	// Check if asserting from interface{} to concrete type
	if t := pd.info.TypeOf(assert.X); t != nil {
		if iface, ok := t.Underlying().(*types.Interface); ok {
			return iface.Empty() // interface{}
		}
	}
	return false
}

func (pd *PatternDetector) isNilSlice(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name == "nil"
	}
	return false
}

func (pd *PatternDetector) isInLoop(call *ast.CallExpr) bool {
	// This would need parent node tracking to determine if we're in a loop
	// Simplified implementation
	return false
}

func (pd *PatternDetector) capturesVariables(fn *ast.FuncLit) bool {
	// This would need sophisticated analysis to determine captured variables
	// Simplified: assume any closure captures variables
	return true
}

func (pd *PatternDetector) isClosureToInterface(fn *ast.FuncLit) bool {
	// This would need context analysis
	return false
}

func (pd *PatternDetector) isSimpleStringFormatting(call *ast.CallExpr) bool {
	// Check if format string is simple (no complex formatting)
	if len(call.Args) >= 1 {
		if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
			// Simple heuristic: if format string has only %s or %d
			formatStr := lit.Value
			return strings.Count(formatStr, "%") <= 2
		}
	}
	return false
}

func (pd *PatternDetector) isInHotPath(call *ast.CallExpr) bool {
	// This would need profiling data or heuristics
	// For now, assume it's in hot path if in a loop
	return pd.isInLoop(call)
}

func (pd *PatternDetector) isValueTypeToInterface(expr ast.Expr) bool {
	// Check if passing value type to interface parameter
	if t := pd.info.TypeOf(expr); t != nil {
		// Check if it's a basic type or struct (value types)
		switch t.Underlying().(type) {
		case *types.Basic, *types.Struct:
			return true
		}
	}
	return false
}
