package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

// pzm's exit codes: the verb ran, the verb failed, nothing could be run.
const (
	exitRan = iota
	exitFailed
	exitNothingRan
)

// Version on a binary that game did not build.
func Example_version() {
	var out bytes.Buffer
	code := run([]string{"version"}, &out, &out, "")
	fmt.Print(out.String())
	fmt.Println(code)
	// Output:
	// nilgai unstamped
	// 0
}

func TestEveryVerbHasItsExample(test *testing.T) {
	if err := tool("").Examples(os.DirFS(".")); err != nil {
		test.Error(err)
	}
}

func TestToolIsDeclaredRight(test *testing.T) {
	if err := tool("").Check(); err != nil {
		test.Error(err)
	}
}

func TestVersionSaysAStampedBuild(test *testing.T) {
	for _, stamp := range []string{"abc1234", "deadbee"} {
		var out bytes.Buffer
		code := run([]string{"version"}, &out, &out, stamp)
		if code != exitRan || out.String() != "nilgai "+stamp+"\n" {
			test.Errorf("stamp %q: %d, %q", stamp, code, out.String())
		}
	}
}

func TestTheFlagFormOfAVerbIsTheVerb(test *testing.T) {
	var out bytes.Buffer
	if code := run([]string{"--version"}, &out, &out, ""); code != exitRan ||
		out.String() != "nilgai unstamped\n" {
		test.Errorf("--version: %d, %q", code, out.String())
	}
}

func TestAnythingElseIsRefused(test *testing.T) {
	for _, args := range [][]string{{"read"}, {"version", "--loud"},
		{"--flag"}} {
		var out, errs bytes.Buffer
		code := run(args, &out, &errs, "")
		if code != exitNothingRan ||
			strings.Contains(out.String(), "nilgai ") {
			test.Errorf("%v: code %d, out %q", args, code, out.String())
		}
	}
}

// closedWriter refuses every write.
type closedWriter struct{}

// Write fails as a closed file does.
func (closedWriter) Write([]byte) (written int, err error) {
	return written, os.ErrClosed
}

func TestVersionFailsWhenItCannotWrite(test *testing.T) {
	var errs bytes.Buffer
	if code := run([]string{"version"}, closedWriter{}, &errs, ""); code != exitFailed {
		test.Errorf("a closed output: code %d", code)
	}
}
