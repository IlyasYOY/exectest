// Package exectest is [os/exec] package testing facilities.
//
// The main goal of the package: declarative testing of any executable..
package exectest

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const (
	// This is the comment start in Lua, so there might be problems.
	filePrefix       = "--file:"
	stdoutPrefix     = "--stdout"
	stderrPrefix     = "--stderr"
	stdinPrefix      = "--stdin"
	envPrefix        = "--env:"
	argPrefix        = "--arg:"
	returnCodePrefix = "--return-code:"
)

type cmdOption func(*exec.Cmd)

// ExecuteForFile the same as the [Execute] but uses a file (path) with a scheme.
func ExecuteForFile(t *testing.T, binary string, file string, opts ...cmdOption) {
	t.Helper()
	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("Failed to read test file %s: %v", file, err)
	}
	Execute(t, binary, string(content), opts...)
}

// Execute is the main testing facility of the package.
//
// Function consists of the steps:
//
//   - [prepareScheme]: create directory, parse args and required stdout.
//   - execute given binary in the prepared conditions.
//   - assert results of the binary evaluation.
//
// Examples:
//
//	--file:a.txt
//	--file:.b.txt
//	--arg:-a
//	--stdout
//	.
//	..
//	.b.txt
//	a.txt
//
// This is a desciption of the command `ls -a` run in the
// directory with a.txt and .b.txt files.
func Execute(t *testing.T, binary, scheme string, opts ...cmdOption) {
	t.Helper()
	schemeResult := prepareScheme(t, scheme)

	executionResult := executeCommand(t, binary, schemeResult.Dir, schemeResult.Args, schemeResult.Stdin, schemeResult.Env, opts)

	assertReturnCode(t, schemeResult.ReturnCode, executionResult.ReturnCode)
	if assertNoDiff(t, "stdout", schemeResult.Stdout, executionResult.Stdout) {
		t.Logf("stdout:\n%s", executionResult.Stdout)
	}
	if assertNoDiff(t, "stderr", schemeResult.Stderr, executionResult.Stderr) {
		t.Logf("stderr:\n%s", executionResult.Stderr)
	}
}

func assertReturnCode(t *testing.T, want, got int) bool {
	t.Helper()
	if got != want {
		t.Errorf("Failed to match return code: want %d, got %d", want, got)
		return true
	}
	return false
}

func assertNoDiff(t *testing.T, name string, want string, got string) bool {
	t.Helper()
	wantLines := toLines(want)
	gotLines := toLines(got)
	if diff := cmp.Diff(wantLines, gotLines); diff != "" {
		t.Errorf("Failed matching %s (-missing line, +extra line): \n%s", name, diff)
		return true
	}
	return false
}

type executionResult struct {
	Stdout     string
	Stderr     string
	ReturnCode int
}

func executeCommand(t *testing.T, binary string, dir string, args []string, stdin string, env []string, opts []cmdOption) executionResult {
	t.Helper()

	cmd := exec.Command(binary)
	var stdoutBuilder strings.Builder
	cmd.Stdout = &stdoutBuilder
	var stderrBuilder strings.Builder
	cmd.Stderr = &stderrBuilder
	cmd.Dir = dir
	cmd.Args = append(cmd.Args, args...)
	cmd.Stdin = strings.NewReader(stdin)
	for _, opt := range opts {
		opt(cmd)
	}
	if len(env) > 0 {
		cmd.Env = append(cmd.Environ(), env...)
	}

	// this is intentional, we will assert exit code manually
	_ = cmd.Run()

	return executionResult{
		Stdout:     stdoutBuilder.String(),
		Stderr:     stderrBuilder.String(),
		ReturnCode: cmd.ProcessState.ExitCode(),
	}
}

type schemeResult struct {
	Stdout     string
	Stderr     string
	Stdin      string
	ReturnCode int
	Args       []string
	Env        []string
	Dir        string
}

func prepareScheme(t *testing.T, scheme string) schemeResult {
	t.Helper()

	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("Test scheme: %s\n", scheme)
		}
	})

	var stdout strings.Builder
	var stderr strings.Builder
	var stdin strings.Builder
	var returnCode int
	var args []string
	var env []string
	files := make(map[string]string)
	dir := t.TempDir()

	// TODO: Make test fail if the same field defined twice.
	// TODO: Replace with enum. FSM is always good for readability.
	var isStdout bool
	var isStderr bool
	var isStdin bool
	var isFile bool

	var currentFileName string
	var currentFile strings.Builder

	saveFile := func(name string) {
		if isFile {
			resultPath := filepath.Join(dir, currentFileName)
			files[resultPath] = currentFile.String()
		}
		currentFileName = name
		currentFile.Reset()
	}

	for _, line := range toLines(scheme) {
		if strings.HasPrefix(line, stderrPrefix) {
			saveFile("")
			isFile = false
			isStderr = true
			isStdin = false
			isStdout = false
			continue
		}
		if strings.HasPrefix(line, stdoutPrefix) {
			saveFile("")
			isFile = false
			isStderr = false
			isStdin = false
			isStdout = true
			continue
		}
		if fileName, ok := strings.CutPrefix(line, filePrefix); ok {
			saveFile(strings.TrimSpace(fileName))
			isFile = true
			isStderr = false
			isStdin = false
			isStdout = false
			continue
		}
		if strings.HasPrefix(line, stdinPrefix) {
			saveFile("")
			isFile = false
			isStderr = false
			isStdin = true
			isStdout = false
			continue
		}

		if rtCodeText, ok := strings.CutPrefix(line, returnCodePrefix); ok {
			rtCodeText = strings.TrimSpace(rtCodeText)
			var err error
			returnCode, err = strconv.Atoi(rtCodeText)
			if err != nil {
				t.Fatalf("Failed to convert return code %q to int: %s", rtCodeText, err)
			}
			continue
		}
		if arg, ok := strings.CutPrefix(line, argPrefix); ok {
			arg = strings.TrimSpace(arg)
			arg = evaluateVariables(arg, dir)
			args = append(args, arg)
			continue
		}
		if kv, ok := strings.CutPrefix(line, envPrefix); ok {
			kv = strings.TrimSpace(kv)
			kv = evaluateVariables(kv, dir)
			if !strings.Contains(kv, "=") {
				t.Fatalf("Malformed --env entry %q, expected KEY=VALUE", kv)
			}
			env = append(env, kv)
			continue
		}

		if isStderr {
			line = evaluateVariables(line, dir)
			stderr.WriteString(line)
			continue
		}
		if isStdout {
			line = evaluateVariables(line, dir)
			stdout.WriteString(line)
			continue
		}
		if isFile {
			line = evaluateVariables(line, dir)
			currentFile.WriteString(line)
			continue
		}
		if isStdin {
			stdin.WriteString(line)
			continue
		}
	}
	if isFile {
		saveFile("")
	}

	for path, content := range files {
		fileDir := filepath.Dir(path)
		if err := os.MkdirAll(fileDir, 0o755); err != nil {
			t.Fatalf("Failed to create directory (%q) for test file: %s", fileDir, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("Failed to write file (%v): %s", path, err)
		}
	}

	return schemeResult{
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		Stdin:      stdin.String(),
		ReturnCode: returnCode,
		Args:       args,
		Env:        env,
		Dir:        dir,
	}
}

func evaluateVariables(data string, dir string) string {
	data = strings.ReplaceAll(data, "{dir}", dir)
	return data
}

// toLines splits strings to lines compatible with [strings.Lines].
func toLines(data string) []string {
	scanner := bufio.NewScanner(strings.NewReader(data))
	scanner.Buffer(nil, 1024*1024) // Set max token size to 1MB for long lines
	lines := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		line += "\n"
		lines = append(lines, line)
	}
	return lines
}
