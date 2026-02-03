package exectest_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/IlyasYOY/exectest"
)

func TestExecuteStdoutAfterFiles(t *testing.T) {
	exectest.Execute(t, "ls", `
--file:a.txt
--file:b.txt
--stdout
a.txt
b.txt
`)
}

func TestExecuteStdoutBeforeFiles(t *testing.T) {
	exectest.Execute(t, "ls", `
--stdout
a.txt
b.txt
--file:a.txt
--file:b.txt
`)
}

func TestExecuteStdoutBetweenFiles(t *testing.T) {
	exectest.Execute(t, "ls", `
--file:a.txt
--stdout
a.txt
b.txt
--file:b.txt
`)
}

func TestExecuteNoHiddenFilesShowed(t *testing.T) {
	exectest.Execute(t, "ls", `
--file:a.txt
--file:.b.txt
--stdout
a.txt
`)
}

func TestExecuteArgumentPassedToShowHiddenFile(t *testing.T) {
	exectest.Execute(t, "ls", `
--file:a.txt
--file:.b.txt
--arg:-a
--stdout
.
..
.b.txt
a.txt
`)
}

func TestExecuteStderrCapture(t *testing.T) {
	exectest.Execute(t, "sh", `
--arg:-c
--arg:echo err >&2; echo out
--stdout
out
--stderr
err
`)
}

func TestExecuteStdinFeedingCat(t *testing.T) {
	exectest.Execute(t, "cat", `
--stdin
hello world
--stdout
hello world
`)
}

func TestExecuteNonZeroReturnCode(t *testing.T) {
	exectest.Execute(t, "sh", `
--arg:-c
--arg:exit 42
--return-code: 42
`)
}

func TestExecuteMultipleArgsWithDirPlaceholder(t *testing.T) {
	exectest.Execute(t, "sh", fmt.Sprintf(`
--arg:-c
--arg:printf "%%s %%s\n" "$1" "$2"
--arg:sh
--arg:-p
--arg:{dir}
--stdout
-p %s
`, "{dir}"))
}

func TestExecuteNestedFileCreationAndReading(t *testing.T) {
	exectest.Execute(t, "cat", `
--file:sub/inner.txt
line1
line2

line4
--arg:sub/inner.txt
--stdout
line1
line2

line4
`)
}

func TestExecuteVariableSubstitutionInsideFileContent(t *testing.T) {
	exectest.Execute(t, "cat", `
--file:info.txt
Path: {dir}
--arg:info.txt
--stdout
Path: {dir}
`)
}

func TestExecuteEmptySchemeNoPrefixesBinaryThatSucceeds(t *testing.T) {
	exectest.Execute(t, "true", ``)
}

func TestExecuteUnknownPrefixIsIgnored(t *testing.T) {
	exectest.Execute(t, "true", `
--unknown:something
--stdout
`)
}

func TestExecuteCmdOptionSetEnvironmentVariable(t *testing.T) {
	exectest.Execute(t, "sh", `
--arg:-c
--arg:echo $TEST_VAR
--stdout
value123
`, func(c *exec.Cmd) {
		c.Env = append(os.Environ(), "TEST_VAR=value123")
	})
}

func TestExecuteParallelExecutions(t *testing.T) {
	for i := range 5 {
		t.Run(fmt.Sprintf("p-%d", i), func(t *testing.T) {
			t.Parallel()
			exectest.Execute(t, "sh", fmt.Sprintf(`
--arg:-c
--arg:printf "%%d\n" "$1"
--arg:sh
--arg:%d
--stdout
%d
`, i, i))
		})
	}
}

func TestExecuteEnvVariableIsSet(t *testing.T) {
	exectest.Execute(t, "sh", `
--env:FOO=hello_world
--arg:-c
--arg:printf "%s\n" "$FOO"
--stdout
hello_world
`)
}

func TestExecuteEnvVariableWithPlaceholder(t *testing.T) {
	exectest.Execute(t, "sh", fmt.Sprintf(`
--env:DIR_PATH=%s
--arg:-c
--arg:printf "dir=%s\n" "$DIR_PATH"
--stdout
dir=%s
`, "{dir}", "{dir}", "{dir}"))
}

func TestExecuteMultipleEnvVariables(t *testing.T) {
	exectest.Execute(t, "sh", `
--env:ONE=1
--env:TWO=2
--arg:-c
--arg:printf "%s:%s\n" "$ONE" "$TWO"
--stdout
1:2
`)
}
