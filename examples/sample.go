package examples

import "fmt"

// Example 1: Pointer to local variable that escapes
func returnLocalPointer() *int {
	x := 42
	return &x // This should be detected - pointer escapes only once
}

// Example 2: new(T) that could be stack allocated
func useNew() *string {
	s := new(string)
	*s = "hello"
	return s // This should be detected - new() always allocates on heap
}

// Example 3: Object reuse - should NOT be flagged
func reuseObject() {
	data := make([]int, 100)

	// Object is reused multiple times
	processData(data)
	processData(data)
	processData(data)

	// This should NOT be flagged because the object is reused
}

// Example 4: Single-use allocation that escapes
func singleUseEscape() interface{} {
	temp := struct{ value int }{value: 123}
	return &temp // Should be detected - single use escape
}

// Example 5: Local allocation that doesn't escape - should NOT be flagged
func localAllocation() {
	x := 42
	y := &x
	fmt.Println(*y) // Local use only, doesn't escape
}

// Example 6: new() in assignment
func newInAssignment() {
	var ptr *int
	ptr = new(int) // Should be detected
	*ptr = 100
	fmt.Println(*ptr)
}

// Example 7: Complex case with multiple allocations
func complexCase() []*string {
	var results []*string

	for i := 0; i < 3; i++ {
		s := fmt.Sprintf("item-%d", i)
		results = append(results, &s) // Should be detected - escaping in loop
	}

	return results
}

// Example 8: Stack-friendly alternative (good pattern)
func stackFriendly() string {
	x := 42
	return fmt.Sprintf("value: %d", x) // Good - returns value, not pointer
}

// Helper function for examples
func processData(data []int) {
	for i := range data {
		data[i] = i * 2
	}
}
