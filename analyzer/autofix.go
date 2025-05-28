package analyzer

import (
	"fmt"
	"go/format"
	"go/token"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// FileWriter interface for writing fixes to files
type FileWriter interface {
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

// RealFileWriter implements FileWriter for actual file system operations
type RealFileWriter struct{}

func (w *RealFileWriter) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}

// AutoFixer handles automatic code fixes based on AI suggestions
type AutoFixer struct {
	fset   *token.FileSet
	writer FileWriter
}

// NewAutoFixer creates a new AutoFixer instance
func NewAutoFixer(fset *token.FileSet) *AutoFixer {
	return &AutoFixer{
		fset:   fset,
		writer: &RealFileWriter{},
	}
}

// NewAutoFixerWithWriter creates a new AutoFixer instance with custom writer
func NewAutoFixerWithWriter(fset *token.FileSet, writer FileWriter) *AutoFixer {
	return &AutoFixer{
		fset:   fset,
		writer: writer,
	}
}

// ApplyFixesToFile applies all fixes to a file and writes the result back
func (af *AutoFixer) ApplyFixesToFile(filename string, fixes []analysis.TextEdit) error {
	// Read the original file
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	// Sort fixes by position (reverse order to apply from end to beginning)
	sort.Slice(fixes, func(i, j int) bool {
		return fixes[i].Pos > fixes[j].Pos
	})

	// Apply each fix
	result := content
	for _, fix := range fixes {
		result, err = af.applyTextEdit(result, fix)
		if err != nil {
			return err
		}
	}

	// Format the result
	formatted, err := format.Source(result)
	if err != nil {
		// If formatting fails, use unformatted result
		formatted = result
	}

	// Write back to file
	return af.writer.WriteFile(filename, formatted, 0644)
}

// applyTextEdit applies a single text edit to the content
func (af *AutoFixer) applyTextEdit(content []byte, edit analysis.TextEdit) ([]byte, error) {
	// Convert token positions to byte offsets
	startOffset := af.tokenPosToByteOffset(content, edit.Pos)
	endOffset := af.tokenPosToByteOffset(content, edit.End)

	if startOffset < 0 || endOffset < 0 || startOffset > len(content) || endOffset > len(content) {
		// Invalid positions, skip this edit
		return content, nil
	}

	// Apply the edit
	result := make([]byte, 0, len(content)+len(edit.NewText))
	result = append(result, content[:startOffset]...)
	result = append(result, edit.NewText...)
	result = append(result, content[endOffset:]...)

	return result, nil
}

// tokenPosToByteOffset converts a token.Pos to byte offset in the content
func (af *AutoFixer) tokenPosToByteOffset(content []byte, pos token.Pos) int {
	if pos == token.NoPos {
		return -1
	}

	position := af.fset.Position(pos)
	if position.Filename == "" {
		return -1
	}

	// Simple approach: count bytes to reach the line and column
	lines := strings.Split(string(content), "\n")
	if position.Line <= 0 || position.Line > len(lines) {
		return -1
	}

	offset := 0
	// Add bytes for all previous lines (including newlines)
	for i := 0; i < position.Line-1; i++ {
		offset += len(lines[i]) + 1 // +1 for newline
	}

	// Add column offset (1-based to 0-based)
	if position.Column > 0 {
		offset += position.Column - 1
	}

	return offset
}

// GenerateAutoFixes creates suggested fixes based on AI suggestions and issue analysis
func (af *AutoFixer) GenerateAutoFixes(issue Issue, aiSuggestion string) []analysis.SuggestedFix {
	var fixes []analysis.SuggestedFix

	// Parse the issue to understand what kind of fix is needed
	if strings.Contains(issue.Message, "new(T)") {
		// Try to generate a fix for new(T) allocations
		if fix := af.generateNewTFix(issue, aiSuggestion); fix != nil {
			fixes = append(fixes, *fix)
		}
	}

	return fixes
}

// generateNewTFix generates a fix for new(T) allocations
func (af *AutoFixer) generateNewTFix(issue Issue, aiSuggestion string) *analysis.SuggestedFix {
	// Read the source file to understand the context
	if issue.Pos.Filename == "" {
		return nil
	}

	content, err := ioutil.ReadFile(issue.Pos.Filename)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(content), "\n")
	if issue.Pos.Line < 1 || issue.Pos.Line > len(lines) {
		return nil
	}

	line := lines[issue.Pos.Line-1]

	// Look for patterns like "s := new(string)" or "i := new(int)"
	if strings.Contains(line, ":= new(") {
		// Find the exact position of "new(Type)" in the line
		newIndex := strings.Index(line, "new(")
		if newIndex == -1 {
			return nil
		}

		// Find the closing parenthesis
		closeIndex := strings.Index(line[newIndex:], ")")
		if closeIndex == -1 {
			return nil
		}
		closeIndex += newIndex + 1 // Adjust for the offset and include the closing paren

		// Extract the type
		newExpr := line[newIndex:closeIndex]
		if !strings.HasPrefix(newExpr, "new(") || !strings.HasSuffix(newExpr, ")") {
			return nil
		}

		typeName := newExpr[4 : len(newExpr)-1] // Remove "new(" and ")"

		// Generate the replacement value
		var replacement string
		switch typeName {
		case "string":
			replacement = `""`
		case "int":
			replacement = "0"
		case "bool":
			replacement = "false"
		case "float64":
			replacement = "0.0"
		default:
			// For other types, we can't easily provide a zero value without more context
			// So we'll skip this fix
			return nil
		}

		// Calculate the absolute positions in the file
		lineStart := 0
		for i := 0; i < issue.Pos.Line-1; i++ {
			lineStart += len(lines[i]) + 1 // +1 for newline
		}

		// Position of the new(Type) expression in the file
		newExprStart := lineStart + newIndex
		newExprEnd := lineStart + closeIndex

		return &analysis.SuggestedFix{
			Message: fmt.Sprintf("Replace new(%s) with zero value", typeName),
			TextEdits: []analysis.TextEdit{
				{
					Pos:     token.Pos(newExprStart),
					End:     token.Pos(newExprEnd),
					NewText: []byte(replacement),
				},
			},
		}
	}

	return nil
}

// FormatCode formats Go code using go/format
func (af *AutoFixer) FormatCode(code string) (string, error) {
	formatted, err := format.Source([]byte(code))
	if err != nil {
		return code, err
	}
	return string(formatted), nil
}

// SmartReplace performs intelligent code replacement with context awareness
func (af *AutoFixer) SmartReplace(issue Issue, oldPattern, newCode string) *analysis.SuggestedFix {
	// This would implement more sophisticated replacement logic
	// considering the AST context, variable scopes, etc.

	return &analysis.SuggestedFix{
		Message: "Smart replacement based on AI suggestion",
		TextEdits: []analysis.TextEdit{
			{
				Pos:     token.Pos(issue.Pos.Offset),
				End:     token.Pos(issue.Pos.Offset + len(oldPattern)),
				NewText: []byte(newCode),
			},
		},
	}
}
