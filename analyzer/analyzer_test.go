package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func TestInspectFile(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []string
	}{
		{
			name: "new allocation",
			code: `
package main

func useNew() *string {
	return new(string)
}
`,
			expected: []string{"new(T) in return/assignment", "new(T) always allocates on heap"},
		},
		{
			name: "local allocation no escape",
			code: `
package main

import "fmt"

func localUse() {
	x := 42
	y := &x
	fmt.Println(*y)
}
`,
			expected: []string{}, // Should not detect anything
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			// Create type info
			info := &types.Info{
				Types: make(map[ast.Expr]types.TypeAndValue),
				Defs:  make(map[*ast.Ident]types.Object),
				Uses:  make(map[*ast.Ident]types.Object),
			}

			// Type check the file
			config := &types.Config{}
			pkg, err := config.Check("test", fset, []*ast.File{file}, info)
			if err != nil {
				t.Logf("Type checking failed (this may be expected): %v", err)
			}
			_ = pkg

			var issues []string
			InspectFile(file, info, fset, func(pos token.Pos, msg string) {
				issues = append(issues, msg)
			})

			if len(tt.expected) == 0 && len(issues) == 0 {
				return // Both empty, test passes
			}

			if len(tt.expected) == 0 && len(issues) > 0 {
				t.Errorf("Expected no issues, got %d: %v", len(issues), issues)
				return
			}

			// Check that we have at least the expected number of issues
			if len(issues) < len(tt.expected) {
				t.Errorf("Expected at least %d issues, got %d: %v", len(tt.expected), len(issues), issues)
				return
			}

			// Check that each expected substring appears in at least one issue
			for _, expected := range tt.expected {
				found := false
				for _, issue := range issues {
					if contains(issue, expected) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find issue containing %q, but didn't find it in: %v", expected, issues)
				}
			}
		})
	}
}

func TestConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MaxAllocSize != 32 {
		t.Errorf("Expected MaxAllocSize to be 32, got %d", config.MaxAllocSize)
	}

	if config.OpenAIModel != "gpt-4" {
		t.Errorf("Expected OpenAIModel to be gpt-4, got %s", config.OpenAIModel)
	}

	if config.OpenAIMaxTokens != 512 {
		t.Errorf("Expected OpenAIMaxTokens to be 512, got %d", config.OpenAIMaxTokens)
	}
}

func TestConfigPatternDisabled(t *testing.T) {
	config := &Config{
		DisablePatterns: []string{"pointer-escape", "new-allocation"},
	}

	if !config.IsPatternDisabled("pointer-escape") {
		t.Error("Expected pointer-escape to be disabled")
	}

	if !config.IsPatternDisabled("new-allocation") {
		t.Error("Expected new-allocation to be disabled")
	}

	if config.IsPatternDisabled("other-pattern") {
		t.Error("Expected other-pattern to not be disabled")
	}
}

func TestAnalyzeFile(t *testing.T) {
	code := `
package main

func test() *int {
	return new(int)
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	// Type check the file
	config := &types.Config{}
	pkg, err := config.Check("test", fset, []*ast.File{file}, info)
	if err != nil {
		t.Logf("Type checking failed (this may be expected): %v", err)
	}
	_ = pkg

	analyzerConfig := DefaultConfig()
	issues := analyzeFile(file, info, fset, analyzerConfig)

	if len(issues) == 0 {
		t.Error("Expected at least one issue to be found")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
