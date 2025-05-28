package examples

import "fmt"

func autofixExample() {
	// This should trigger stackalloc detection
	s := new(string)
	*s = "hello"
	fmt.Println(*s)

	// Another allocation that could be optimized
	i := new(int)
	*i = 42
	fmt.Println(*i)
}
