package analyzer

import (
	"go/format"
	"go/parser"
	"go/token"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// AutoFixer handles automatic code fixes based on AI suggestions
type AutoFixer struct {
	fset *token.FileSet
}

// NewAutoFixer creates a new AutoFixer instance
func NewAutoFixer(fset *token.FileSet) *AutoFixer {
	return &AutoFixer{fset: fset}
}

// GenerateAutoFixes creates automatic code fixes from AI suggestions
func (af *AutoFixer) GenerateAutoFixes(issue Issue, aiSuggestion string) []analysis.SuggestedFix {
	var fixes []analysis.SuggestedFix

	// Try different fix strategies based on the issue type
	if strings.Contains(issue.Message, "new(T)") {
		if fix := af.fixNewAllocation(issue, aiSuggestion); fix != nil {
			fixes = append(fixes, *fix)
		}
	}

	if strings.Contains(issue.Message, "pointer to") && strings.Contains(issue.Message, "escapes") {
		if fix := af.fixEscapingPointer(issue, aiSuggestion); fix != nil {
			fixes = append(fixes, *fix)
		}
	}

	return fixes
}

// fixNewAllocation handles fixes for new(T) allocations
func (af *AutoFixer) fixNewAllocation(issue Issue, aiSuggestion string) *analysis.SuggestedFix {
	// Extract the type from new(Type) pattern
	typePattern := regexp.MustCompile(`new\((\w+)\)`)

	// Read the source file to understand the context
	pos := af.fset.Position(token.Pos(issue.Pos.Offset))
	if pos.Filename == "" {
		return nil
	}

	// Parse AI suggestion for better replacement
	replacement := af.parseAISuggestionForReplacement(aiSuggestion, "new")
	if replacement == "" {
		// Default replacement strategy
		if matches := typePattern.FindStringSubmatch(issue.Message); len(matches) > 1 {
			typeName := matches[1]
			replacement = "var value " + typeName + "; &value"
		} else {
			replacement = "/* TODO: Replace new() with stack allocation */"
		}
	}

	return &analysis.SuggestedFix{
		Message: "Replace new(T) with stack allocation",
		TextEdits: []analysis.TextEdit{
			{
				Pos:     token.Pos(issue.Pos.Offset),
				End:     token.Pos(issue.Pos.Offset + 20), // Approximate range
				NewText: []byte(replacement),
			},
		},
	}
}

// fixEscapingPointer handles fixes for escaping pointer issues
func (af *AutoFixer) fixEscapingPointer(issue Issue, aiSuggestion string) *analysis.SuggestedFix {
	// Extract variable name from the message
	varPattern := regexp.MustCompile(`pointer to (\w+) escapes`)
	matches := varPattern.FindStringSubmatch(issue.Message)
	if len(matches) < 2 {
		return nil
	}

	varName := matches[1]

	// Parse AI suggestion for better replacement
	replacement := af.parseAISuggestionForReplacement(aiSuggestion, "return")
	if replacement == "" {
		// Default strategy: suggest returning value instead of pointer
		replacement = "/* Consider returning " + varName + " by value instead of pointer */"
	}

	return &analysis.SuggestedFix{
		Message: "Avoid pointer escape by returning value",
		TextEdits: []analysis.TextEdit{
			{
				Pos:     token.Pos(issue.Pos.Offset),
				End:     token.Pos(issue.Pos.Offset + 10),
				NewText: []byte(replacement),
			},
		},
	}
}

// parseAISuggestionForReplacement extracts actionable code from AI suggestions
func (af *AutoFixer) parseAISuggestionForReplacement(suggestion, context string) string {
	// Look for code blocks in the AI suggestion
	codeBlockPattern := regexp.MustCompile("```(?:go)?\n(.*?)\n```")
	matches := codeBlockPattern.FindAllStringSubmatch(suggestion, -1)

	for _, match := range matches {
		if len(match) > 1 {
			code := strings.TrimSpace(match[1])
			// Validate that this is reasonable Go code
			if af.isValidGoCode(code) {
				return code
			}
		}
	}

	// Look for "After:" patterns
	afterPattern := regexp.MustCompile(`(?i)after:\s*\n(.*?)(?:\n\n|\n//|$)`)
	if matches := afterPattern.FindStringSubmatch(suggestion); len(matches) > 1 {
		code := strings.TrimSpace(matches[1])
		if af.isValidGoCode(code) {
			return code
		}
	}

	// Look for direct suggestions
	suggestionPattern := regexp.MustCompile(`(?i)(?:replace|use|try):\s*(.*?)(?:\n|$)`)
	if matches := suggestionPattern.FindStringSubmatch(suggestion); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}

// isValidGoCode performs basic validation of Go code snippets
func (af *AutoFixer) isValidGoCode(code string) bool {
	// Wrap in a function to make it parseable
	wrappedCode := "package main\nfunc test() {\n" + code + "\n}"

	_, err := parser.ParseFile(af.fset, "", wrappedCode, parser.ParseComments)
	return err == nil
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
