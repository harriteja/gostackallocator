# Go Stack Allocator Analyzer

A static analysis tool for Go that detects heap allocations and suggests stack allocation optimizations with AI-powered autofix capabilities.

## Features

### Comprehensive Allocation Pattern Detection

The analyzer detects 13+ different types of allocation patterns:

#### 1. **new(T) Calls**
```go
// Detected
s := new(string)
i := new(int)

// Suggested Fix
var s string    // or s := ""
var i int       // or i := 0
```

#### 2. **make() Patterns**
```go
// Small slice allocations
smallSlice := make([]int, 5)  // → Consider using [5]int{} array

// Large slice allocations  
largeSlice := make([]int, 10000)  // → May cause GC pressure

// Maps without size hints
m := make(map[string]int)  // → Consider make(map[string]int, expectedSize)
```

#### 3. **Slice Literals**
```go
// Small slice literals
numbers := []int{1, 2, 3, 4}  // → Consider [4]int{1, 2, 3, 4}

// Complex nested slices
nested := [][]int{{1, 2}, {3, 4}}  // → Multiple allocations detected
```

#### 4. **Map Literals**
```go
// Small maps
config := map[string]int{"a": 1, "b": 2}  // → Consider struct or switch
```

#### 5. **String Concatenation**
```go
// Multiple concatenations
result := "Hello " + name + "!"  // → Consider strings.Builder
```

#### 6. **append() Patterns**
```go
// Append to nil
var slice []int
slice = append(slice, 1, 2, 3)  // → Pre-allocate with make()

// Append in loops
for i := 0; i < 100; i++ {
    slice = append(slice, i)  // → Pre-allocate capacity
}
```

#### 7. **String Formatting**
```go
// Simple formatting
s := fmt.Sprintf("Number: %d", 42)  // → Consider direct conversion
converted := strconv.Itoa(123)      // → Consider strconv.AppendInt
```

#### 8. **Reflection Allocations**
```go
// Reflection-based allocations
val := reflect.New(typ)                    // → Always heap allocated
slice := reflect.MakeSlice(sliceType, 10, 10)  // → Consider alternatives
```

#### 9. **Interface Boxing**
```go
// Value boxing
var iface interface{} = 42  // → Value boxed to interface{}
```

#### 10. **Type Assertions**
```go
// Interface to concrete
if val, ok := iface.(int); ok {  // → May cause allocation if boxed
    // ...
}
```

#### 11. **Closure Captures**
```go
// Variable capture
counter := 0
fn := func() { counter++ }  // → Captures variable, may allocate
```

#### 12. **Large Struct Literals**
```go
// Large structs
large := LargeStruct{
    Field1: "a", Field2: "b", /* ... 10+ fields */
}  // → Consider using pointer or smaller structs
```

#### 13. **Channel Patterns**
```go
// Unbuffered channels
ch := make(chan int)     // → Consider if synchronous needed
ch2 := make(chan int, 1) // → Small buffer detected
```

## Installation

```bash
go install github.com/harriteja/gostackallocator/cmd@latest
```

## Usage

### Basic Analysis
```bash
# Analyze a single file
go vet -vettool=stackalloc main.go

# Analyze a package
go vet -vettool=stackalloc ./...

# Analyze with detailed output
go vet -vettool=stackalloc -stackalloc.verbose=true ./...
```

### AI-Powered Autofix
```bash
# Enable automatic code fixes
go vet -vettool=stackalloc -stackalloc.autofix=true main.go

# Generate reports with suggestions
go vet -vettool=stackalloc -stackalloc.report=true ./...
```

### Configuration Options

- `-stackalloc.autofix=true`: Enable automatic code fixes
- `-stackalloc.report=true`: Generate detailed reports
- `-stackalloc.verbose=true`: Enable verbose output
- `-stackalloc.ai-key=<key>`: OpenAI API key for enhanced suggestions

## Example Output

```bash
$ go vet -vettool=stackalloc examples/demo.go

examples/demo.go:10:6: new(T) always allocates on heap; consider using stack allocation if object doesn't escape
examples/demo.go:15:12: small slice literal; consider using array for stack allocation  
examples/demo.go:20:15: small slice allocation with make(); consider using array or stack allocation
examples/demo.go:25:13: string concatenation with + operator allocates; consider using strings.Builder
examples/demo.go:30:12: appending to nil slice causes allocation; consider pre-allocating with make()
examples/demo.go:35:14: simple string formatting; consider using string concatenation or strings.Builder
examples/demo.go:40:17: reflection-based allocation always uses heap; consider avoiding if performance critical
```

## Autofix Examples

### Before Autofix:
```go
func example() {
    s := new(string)
    numbers := []int{1, 2, 3}
    result := "Hello " + "World"
}
```

### After Autofix:
```go
func example() {
    s := ""
    numbers := [3]int{1, 2, 3}
    result := "Hello World"  // or use strings.Builder for multiple concatenations
}
```

## Performance Impact

The analyzer helps identify allocations that can impact performance:

- **Heap vs Stack**: Stack allocations are ~10x faster than heap
- **GC Pressure**: Fewer heap allocations = less garbage collection overhead  
- **Memory Efficiency**: Arrays use less memory than slices for fixed-size data
- **String Operations**: strings.Builder is much faster for multiple concatenations

## Integration

### CI/CD Pipeline
```yaml
- name: Run Stack Allocation Analysis
  run: |
    go vet -vettool=stackalloc ./...
    if [ $? -ne 0 ]; then
      echo "Stack allocation issues found"
      exit 1
    fi
```

### Pre-commit Hook
```bash
#!/bin/sh
go vet -vettool=stackalloc ./...
```

## Advanced Features

### Pattern-Specific Analysis
The tool provides context-aware suggestions based on usage patterns:

- **Hot Path Detection**: Identifies allocations in performance-critical code
- **Loop Analysis**: Detects allocations inside loops that may cause performance issues
- **Escape Analysis Integration**: Works with Go's escape analysis for better suggestions
- **Type-Aware Fixes**: Provides appropriate zero values for different types

### AI Integration
When configured with an OpenAI API key, the tool provides:

- **Context-Aware Suggestions**: AI analyzes code context for better recommendations
- **Performance Explanations**: Detailed explanations of why changes improve performance
- **Alternative Implementations**: Multiple optimization strategies for complex cases

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new allocation patterns
4. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Changelog

### v2.0.0 - Enhanced Pattern Detection
- Added 12 new allocation pattern types
- Implemented AI-powered autofix capabilities  
- Enhanced string concatenation detection
- Added reflection and interface boxing detection
- Improved closure capture analysis
- Added comprehensive test suite

### v1.0.0 - Initial Release
- Basic new(T) call detection
- Simple autofix for primitive types
- Command-line interface 