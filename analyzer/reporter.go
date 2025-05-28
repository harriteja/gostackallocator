package analyzer

import (
	"context"
	"fmt"
	"go/token"
	"io/ioutil"
	"strings"
	"sync"

	"golang.org/x/tools/go/analysis"
)

// FixTracker tracks fixes to be applied to files
type FixTracker struct {
	mu    sync.Mutex
	fixes map[string][]analysis.TextEdit // filename -> list of fixes
}

// NewFixTracker creates a new fix tracker
func NewFixTracker() *FixTracker {
	return &FixTracker{
		fixes: make(map[string][]analysis.TextEdit),
	}
}

// AddFix adds a fix for a specific file
func (ft *FixTracker) AddFix(filename string, edits []analysis.TextEdit) {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	// Deduplicate overlapping edits
	existingEdits := ft.fixes[filename]
	for _, newEdit := range edits {
		// Check if this edit overlaps with existing ones
		overlaps := false
		for i, existingEdit := range existingEdits {
			if newEdit.Pos <= existingEdit.End && newEdit.End >= existingEdit.Pos {
				// Overlapping edit found - replace if the new one is better
				if len(newEdit.NewText) > 0 && !strings.Contains(string(newEdit.NewText), "TODO") {
					existingEdits[i] = newEdit
				}
				overlaps = true
				break
			}
		}

		// If no overlap, add the new edit
		if !overlaps {
			existingEdits = append(existingEdits, newEdit)
		}
	}

	ft.fixes[filename] = existingEdits
}

// ApplyAllFixes applies all tracked fixes using the provided AutoFixer
func (ft *FixTracker) ApplyAllFixes(autoFixer *AutoFixer) error {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	for filename, edits := range ft.fixes {
		if len(edits) > 0 {
			err := autoFixer.ApplyFixesToFile(filename, edits)
			if err != nil {
				return fmt.Errorf("failed to apply fixes to %s: %v", filename, err)
			}
		}
	}
	return nil
}

// GetFilesWithFixes returns a list of files that have fixes
func (ft *FixTracker) GetFilesWithFixes() []string {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	var files []string
	for filename, edits := range ft.fixes {
		if len(edits) > 0 {
			files = append(files, filename)
		}
	}
	return files
}

// FormatIssue converts an Issue into an analysis.Diagnostic
func FormatIssue(issue Issue, aiClient AIClient, fset *token.FileSet, config *Config) analysis.Diagnostic {
	diagnostic := analysis.Diagnostic{
		Pos:      token.Pos(issue.Pos.Offset),
		Message:  issue.Message,
		Category: "stackalloc",
	}

	// Add AI-powered suggestion if enabled
	if !config.OpenAIDisable && aiClient != nil {
		if suggestion := getAISuggestion(issue, aiClient, fset, config); suggestion != "" {
			// Generate automatic fixes if enabled
			if config.AutoFix {
				if fixes := generateCodeFixes(issue, suggestion, fset); len(fixes) > 0 {
					diagnostic.SuggestedFixes = fixes
				} else {
					// Fallback to comment-based suggestion
					diagnostic.SuggestedFixes = []analysis.SuggestedFix{
						{
							Message: "AI-suggested improvement (enable -autofix for automatic fixes)",
							TextEdits: []analysis.TextEdit{
								{
									Pos:     token.Pos(issue.Pos.Offset),
									End:     token.Pos(issue.Pos.Offset),
									NewText: []byte(fmt.Sprintf("// AI suggestion: %s\n", suggestion)),
								},
							},
						},
					}
				}
			} else {
				// Just add suggestion as comment
				diagnostic.SuggestedFixes = []analysis.SuggestedFix{
					{
						Message: "AI-suggested improvement (enable -autofix for automatic fixes)",
						TextEdits: []analysis.TextEdit{
							{
								Pos:     token.Pos(issue.Pos.Offset),
								End:     token.Pos(issue.Pos.Offset),
								NewText: []byte(fmt.Sprintf("// AI suggestion: %s\n", suggestion)),
							},
						},
					},
				}
			}
		}
	}

	return diagnostic
}

// FormatIssueWithFixTracker converts an Issue into an analysis.Diagnostic and tracks fixes
func FormatIssueWithFixTracker(issue Issue, aiClient AIClient, fset *token.FileSet, config *Config, fixTracker *FixTracker) analysis.Diagnostic {
	diagnostic := FormatIssue(issue, aiClient, fset, config)

	// If autofix is enabled and we have suggested fixes, track them for later application
	if config.AutoFix && len(diagnostic.SuggestedFixes) > 0 {
		position := fset.Position(token.Pos(issue.Pos.Offset))
		if position.Filename != "" {
			// Collect all text edits from all suggested fixes
			var allEdits []analysis.TextEdit
			for _, fix := range diagnostic.SuggestedFixes {
				allEdits = append(allEdits, fix.TextEdits...)
			}
			fixTracker.AddFix(position.Filename, allEdits)
		}
	}

	return diagnostic
}

// getAISuggestion gets an AI-powered code suggestion for the issue
func getAISuggestion(issue Issue, aiClient AIClient, fset *token.FileSet, config *Config) string {
	ctx := context.Background()

	// Get code snippet around the issue
	snippet := getCodeSnippetFromPosition(issue.Pos, fset)
	if snippet == "" {
		return ""
	}

	// Get AI suggestion
	suggestion, err := aiClient.SuggestFix(ctx, snippet, issue.Message)
	if err != nil {
		// Log error but don't fail the analysis
		return ""
	}

	return suggestion
}

// getCodeSnippetFromPosition extracts code snippet from file position
func getCodeSnippetFromPosition(pos token.Position, fset *token.FileSet) string {
	if pos.Filename == "" {
		return ""
	}

	// Read the source file
	src, err := ioutil.ReadFile(pos.Filename)
	if err != nil {
		return ""
	}

	// Convert position to token.Pos for GetCodeSnippet
	tokenPos := fset.Position(token.Pos(pos.Offset)).Offset
	if tokenPos == 0 {
		// Fallback: try to find position by line/column
		lines := splitLines(src)
		if pos.Line > 0 && pos.Line <= len(lines) {
			start := max(0, pos.Line-3)
			end := min(len(lines), pos.Line+2)

			snippet := ""
			for i := start; i < end; i++ {
				if i == pos.Line-1 {
					snippet += fmt.Sprintf(">>> %s\n", lines[i])
				} else {
					snippet += fmt.Sprintf("    %s\n", lines[i])
				}
			}
			return snippet
		}
	}

	return GetCodeSnippet(fset, token.Pos(tokenPos), src)
}

// generateCodeFixes attempts to generate actual code fixes based on AI suggestions
func generateCodeFixes(issue Issue, suggestion string, fset *token.FileSet) []analysis.SuggestedFix {
	// Use the new AutoFixer for more sophisticated fixes
	autoFixer := NewAutoFixer(fset)
	fixes := autoFixer.GenerateAutoFixes(issue, suggestion)

	if len(fixes) > 0 {
		return fixes
	}

	// Fallback to simple fixes if AutoFixer doesn't handle the case
	var fallbackFixes []analysis.SuggestedFix

	if strings.Contains(issue.Message, "new(T)") {
		fallbackFixes = append(fallbackFixes, analysis.SuggestedFix{
			Message: "Replace new(T) with stack allocation",
			TextEdits: []analysis.TextEdit{
				{
					Pos:     token.Pos(issue.Pos.Offset),
					End:     token.Pos(issue.Pos.Offset + 10),
					NewText: []byte("/* TODO: Replace with stack allocation */"),
				},
			},
		})
	}

	return fallbackFixes
}

// ReportIssue is a helper function to report an issue with proper formatting
func ReportIssue(pass *analysis.Pass, issue Issue, aiClient AIClient, config *Config) {
	diagnostic := FormatIssue(issue, aiClient, pass.Fset, config)
	pass.Report(diagnostic)
}

// ReportIssueWithAutoFix reports an issue and applies fixes automatically if enabled
func ReportIssueWithAutoFix(pass *analysis.Pass, issue Issue, aiClient AIClient, config *Config, fixTracker *FixTracker) {
	diagnostic := FormatIssueWithFixTracker(issue, aiClient, pass.Fset, config, fixTracker)
	pass.Report(diagnostic)
}
