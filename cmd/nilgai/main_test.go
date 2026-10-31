package main

import (
	"bytes"
	"fmt"
	"image/png"
	"os"
	"strings"
	"testing"
)

// pzm's exit codes. The verb ran, the verb failed, or nothing
// could be run.
const (
	exitRan = iota
	exitFailed
	exitNothingRan
)

// nothingToRead returns standard input for a verb that reads none.
func nothingToRead() *strings.Reader { return strings.NewReader("") }

// Example_show runs show on the real searches around Sgr A*. The picture
// is a png of the size asked for.
func Example_show() {
	var out, errs bytes.Buffer
	code := run([]string{"show",
		"--chandra", "testdata/chandra-sgra.tsv",
		"--allwise", "testdata/allwise-sgra.csv",
		"--ra", "266.4168371", "--dec", "-29.0078106",
		"--width", "12", "--pixels", "300"},
		nothingToRead(), &out, &errs, "")
	picture, err := png.Decode(&out)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(code, picture.Bounds(), errs.Len())
	// Output:
	// 0 (0,0)-(300,300) 0
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

// TestAnythingElseIsRefused checks that pzm refuses any verb but show.
// For example, version is not a verb, because pzm's --age says which
// build this is.
func TestAnythingElseIsRefused(test *testing.T) {
	for _, args := range [][]string{{"read"}, {"version"},
		{"--version"}, {"--flag"}} {
		var out, errs bytes.Buffer
		code := run(args, nothingToRead(), &out, &errs, "")
		if code != exitNothingRan ||
			strings.Contains(out.String(), "nilgai ") {
			test.Errorf("%v: code %d, out %q", args, code, out.String())
		}
	}
}

// closedWriter is a writer that refuses every write.
type closedWriter struct{}

// Write fails as a closed file does.
func (closedWriter) Write([]byte) (written int, err error) {
	return written, os.ErrClosed
}
