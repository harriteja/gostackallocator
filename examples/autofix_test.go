package examples

import "fmt"

func autofixExample() {
	// This should trigger stackalloc detection
	s :="")
	*s = "hello"
	fmt.Println(*s)

	// Another allocation that could be optimized
	i :=0)
	*i = 42
	fmt.Println(*i)
}
