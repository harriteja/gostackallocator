package examples

import "fmt"

// DemoAutoFix demonstrates various allocation patterns that can be auto-fixed
func DemoAutoFix() {
	// String allocation - will be fixed to empty string
	name :="")
	*name = "Alice"
	fmt.Printf("Hello, %s!\n", *name)

	// Integer allocation - will be fixed to zero
	count :=0)
	*count = 42
	fmt.Printf("Count: %d\n", *count)

	// Boolean allocation - will be fixed to false
	flag :=false)
	*flag = true
	fmt.Printf("Flag: %v\n", *flag)

	// Float allocation - will be fixed to 0.0
	price :=0.0)
	*price = 99.99
	fmt.Printf("Price: $%.2f\n", *price)
}
