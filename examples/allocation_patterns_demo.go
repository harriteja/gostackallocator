package examples

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// DemoAllocationPatterns demonstrates various allocation patterns
func DemoAllocationPatterns() {
	// 1. new(T) calls - already supported
	s := new(string)
	i := new(int)
	b := new(bool)
	f := new(float64)

	// 2. make() calls with various patterns
	// Small slice allocation
	smallSlice := make([]int, 5)

	// Large slice allocation
	largeSlice := make([]int, 10000)

	// Map without size hint
	m1 := make(map[string]int)

	// Map with small known size
	m2 := make(map[string]int, 3)

	// Channel patterns
	ch1 := make(chan int)    // Unbuffered
	ch2 := make(chan int, 1) // Small buffer

	// 3. Slice literals that could use arrays
	numbers := []int{1, 2, 3, 4}       // Small slice literal
	colors := []string{"red", "green"} // Another small slice

	// Complex slice with nested allocations
	complexSlice := [][]int{{1, 2}, {3, 4}}

	// 4. Map literals
	smallMap := map[string]int{
		"one": 1,
		"two": 2,
	}

	// 5. String concatenation patterns
	name := "John"
	greeting := "Hello " + name + "!"

	// Multiple concatenations (inefficient)
	result := ""
	for i := 0; i < 10; i++ {
		result = result + strconv.Itoa(i) + " "
	}

	// 6. append() patterns
	var nilSlice []int
	nilSlice = append(nilSlice, 1, 2, 3) // Append to nil

	// Append in loop (causes multiple reallocations)
	var loopSlice []int
	for i := 0; i < 100; i++ {
		loopSlice = append(loopSlice, i)
	}

	// 7. String formatting allocations
	formatted := fmt.Sprintf("Number: %d", 42)
	simple := fmt.Sprint("Simple")
	converted := strconv.Itoa(123)

	// 8. Reflection-based allocations
	typ := reflect.TypeOf(0)
	reflectValue := reflect.New(typ)
	reflectSlice := reflect.MakeSlice(reflect.TypeOf([]int{}), 10, 10)

	// 9. Interface boxing
	var iface interface{} = 42 // Boxing int to interface{}

	// 10. Type assertions
	if val, ok := iface.(int); ok {
		fmt.Println("Value:", val)
	}

	// 11. Closure patterns
	counter := 0
	increment := func() int {
		counter++ // Captures variable
		return counter
	}

	// Closure assigned to interface
	var fn interface{} = func() { fmt.Println("Hello") }

	// 12. Large struct literals
	type LargeStruct struct {
		Field1, Field2, Field3, Field4, Field5  string
		Field6, Field7, Field8, Field9, Field10 int
		Field11, Field12                        bool
	}

	large := LargeStruct{
		Field1: "a", Field2: "b", Field3: "c", Field4: "d", Field5: "e",
		Field6: 1, Field7: 2, Field8: 3, Field9: 4, Field10: 5,
		Field11: true, Field12: false,
	}

	// Use variables to avoid unused variable errors
	_ = s
	_ = i
	_ = b
	_ = f
	_ = smallSlice
	_ = largeSlice
	_ = m1
	_ = m2
	_ = ch1
	_ = ch2
	_ = numbers
	_ = colors
	_ = complexSlice
	_ = smallMap
	_ = greeting
	_ = result
	_ = nilSlice
	_ = loopSlice
	_ = formatted
	_ = simple
	_ = converted
	_ = reflectValue
	_ = reflectSlice
	_ = increment
	_ = fn
	_ = large
}

// DemoOptimizedPatterns shows better alternatives
func DemoOptimizedPatterns() {
	// 1. Use zero values instead of new()
	var s string  // Instead of new(string)
	var i int     // Instead of new(int)
	var b bool    // Instead of new(bool)
	var f float64 // Instead of new(float64)

	// 2. Use arrays for small, fixed-size collections
	numbers := [4]int{1, 2, 3, 4} // Instead of []int{1, 2, 3, 4}

	// 3. Pre-allocate slices with known capacity
	slice := make([]int, 0, 100) // Instead of appending to nil
	for i := 0; i < 100; i++ {
		slice = append(slice, i)
	}

	// 4. Use strings.Builder for multiple concatenations
	var builder strings.Builder
	for i := 0; i < 10; i++ {
		builder.WriteString(strconv.Itoa(i))
		builder.WriteString(" ")
	}
	result := builder.String()

	// 5. Provide map capacity hints
	m := make(map[string]int, 10) // Instead of make(map[string]int)

	// 6. Use concrete types instead of interfaces when possible
	var fn func() = func() { fmt.Println("Hello") } // Instead of interface{}

	// 7. Pass values as parameters instead of capturing
	increment := func(counter int) int {
		return counter + 1
	}

	// Use variables to avoid unused variable errors
	_ = s
	_ = i
	_ = b
	_ = f
	_ = numbers
	_ = slice
	_ = result
	_ = m
	_ = fn
	_ = increment
}

func main() {
	fmt.Println("=== Allocation Patterns Demo ===")
	DemoAllocationPatterns()

	fmt.Println("\n=== Optimized Patterns Demo ===")
	DemoOptimizedPatterns()
}
