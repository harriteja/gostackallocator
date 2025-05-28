package examples

import "fmt"

func TestEnhancedPatterns() {
	// Basic new() patterns
	s := ""
	i := 0

	// Make patterns
	smallSlice := // AI suggestion: Consider optimizing this allocation
		make([]int, 5)
	m := // AI suggestion: Consider optimizing this allocation
		make(map[string]int)

	// Slice literals
	numbers := // AI suggestion: Consider optimizing this allocation
		[]int{1, 2, 3}

	// String concatenation
	greeting := "Hello " + "World"

	// Use variables
	fmt.Println(s, i, smallSlice, m, numbers, greeting)
}
