# exectest - Declarative CLI testing

You need to test plain-old-cli application with integration tests! That's it!

Build you binary and run in with the tool:

```go
exectest.Execute(t, "ls", `
This is an arbitrary test description.
--file:a.txt
--file:.b.txt
--stdout
a.txt
`)
```

Or you might call it for the file path to the description using
`ExecuteForFile` function.

Check out the docs for more info. See the [tests](./executor_test.go) for
examples.
