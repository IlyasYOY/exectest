# exectest - Go Package for Testing Executables

## Project Overview

exectest is a Go package that provides testing facilities for executables, similar to the `os/exec` package. The main goal of the package is to enable declarative testing of any executable by defining test schemes that specify input files, arguments, environment variables, expected stdout/stderr output, and return codes.

### Key Features

- **Declarative Testing**: Define test cases using a scheme-based approach with prefixes like `--file:`, `--stdout`, `--stderr`, `--arg:`, `--env:`, etc.
- **File System Setup**: Automatically creates temporary directories with specified files for testing
- **Flexible Assertions**: Compare actual vs expected stdout, stderr, return codes, and environment variables
- **Variable Substitution**: Support for `{dir}` placeholder that gets replaced with the temporary test directory
- **Custom Command Options**: Ability to pass custom options to the underlying `exec.Cmd`

### Architecture

The package consists of:
- `executor.go`: Main implementation with functions for parsing schemes, executing commands, and asserting results
- `executor_test.go`: Comprehensive test suite demonstrating various use cases
- Supporting files: `go.mod`, `go.sum`, `Makefile`, CI workflow

## Building and Running

### Prerequisites
- Go 1.25.6 or later

### Commands

```bash
# Run all tests with shuffling enabled
make test

# Run tests with coverage
make test-cover

# Generate and open coverage report in browser
make test-cover-open

# Run go vet for static analysis
make vet

# Clean coverage output file
make clean
```

Alternatively, you can run the commands directly:
```bash
# Run tests
go test ./... -test.shuffle=on -test.fullpath

# Run tests with coverage
go test -cover ./... -test.shuffle=on -test.fullpath

# Run vet
go vet ./...
```

## Development Conventions

### Testing Approach
- Tests use the declarative scheme format to define test cases
- Each test creates isolated temporary directories to avoid interference
- Parallel tests are supported using `t.Parallel()`
- Comprehensive test coverage is maintained

### Scheme Format
The test scheme supports the following prefixes:
- `--file:<filename>`: Creates a file with the following content until the next prefix
- `--stdout`: Defines expected stdout content
- `--stderr`: Defines expected stderr content  
- `--stdin`: Provides input to the command's stdin
- `--arg:<argument>`: Adds an argument to the command
- `--env:<KEY=VALUE>`: Sets an environment variable
- `--return-code:<code>`: Specifies the expected return code

### Code Style
- Follows Go idioms and best practices
- Uses helper functions for repetitive testing logic
- Includes comprehensive error handling with `t.Fatalf` and `t.Errorf`
- Uses `t.Helper()` for wrapper functions to attribute failures to calling code

## Usage Example

From the test file, here's a typical usage example:

```go
exectest.Execute(t, "ls", `
--file:a.txt
--file:b.txt
--arg:-a
--stdout
.
..
a.txt
b.txt
`)
```

This creates a temporary directory with `a.txt` and `b.txt` files, runs `ls -a` in that directory, and asserts that the stdout matches the expected output.

## Dependencies

- `github.com/google/go-cmp`: Used for comparing expected vs actual output with detailed diff reporting

## Project Structure

- `executor.go`: Main package implementation
- `executor_test.go`: Comprehensive test suite
- `go.mod` / `go.sum`: Module definition and dependency tracking
- `Makefile`: Build and test commands
- `.github/workflows/ci.yml`: Continuous integration setup
- `LICENSE`: MIT License terms

