# termtest

<p align="center">
  <a href="https://github.com/ActiveState/termtest/actions?query=workflow%3Aunit-tests"><img alt="GitHub Actions status" src="https://github.com/ActiveState/termtest/workflows/unit-tests/badge.svg" /></a>
</p>

An automatable terminal session with send/expect controls.

This package leverages the
[go-expect](https://github.com/ActiveState/termtest/expect) package to test
terminal applications on Linux, MacOS and Windows, which has been forked from
[Netflix/go-expect](https://github.com/Netflix/go-expect)

It has been developed for CI testing of the [ActiveState state
tool](https://www.activestate.com/products/platform/state-tool/)

## Example usage

```go

import (
    "testing"

    "github.com/ActiveState/termtest"
    "github.com/stretchr/testify/suite"
)

func TestBash(t *testing.T) {
    opts := termtest.Options{
        CmdName: "/bin/bash",
    }
    cp, err := termtest.NewTest(t, opts)
    require.NoError(t, err, "create console process")
    defer cp.Close()

    cp.SendLine("echo hello world")
    cp.Expect("hello world")
    cp.SendLine("exit")
    cp.ExpectExitCode(0)
}

```

## Multi-line matching

After each bytes `termtest` receives from the pseudo-terminal output, it updates the state of the virtual terminal like a terminal user would see it (including a scroll back buffer if necessary).  The `Expect()` look for matches in this processed output. Of course, the terminal wraps its output after text gets longer than 80 columns (or whatever width you have configured for your terminal). As this makes it more difficult to match long string, the default `Expect()` removes all these automatic wraps.

Consider the following examples, that all assume a terminal width of 10 columns.

### Programme sends a line with more than 10 characters

- Programme sends string "0123456789012345".
- Terminal output is "0123456789\n012345     \n".

```
cp.Expect("0123456789012345")  // this matches
```

### Programme sends several lines separated by `\n`

- Programme sends string "line 1\nline 2\n".
- Terminal output is "line 1    \nline 2    \n".
- The following does NOT match:
  ```
  cp.Expect("line 1\nline 2\n")  // this does NOT match
  ```
- The following does MATCH:
  ```
  cp.Expect("line 1")
  cp.Expect("line 2")
  ```
- The following does MATCH:
  ```
  cp.Expect("line 1    line 2    ")
  ```

### Custom matchers

Custom matchers that match against either the raw / or processed pseudo-terminal output can be specified in the `go-expect` package.  See `expect_opt.go` for examples.


