// Command nilgai is the command line for the nilgai library. It has one
// verb for now, version. The readers will get verbs as they are written.
// It is on pzm, which gives it --help, --agent-help, --age and --who.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/janearc/puzzlemouth/pzm"
	"github.com/janearc/puzzlemouth/verb"
)

// the four stamps game build writes into a command's main package
var build, built, who, sessions string

// afterName is where the arguments begin in os.Args, as go has defined
// it: the program's own name comes first.
const afterName = 1

// unstamped is what version says of a binary that game did not build.
const unstamped = "unstamped"

// main runs nilgai on what it was asked.
func main() {
	os.Exit(run(os.Args[afterName:], os.Stdout, os.Stderr, build))
}

// run is nilgai called with the arguments, writing to out and errs, as the
// build stamped as given. It returns the exit code.
func run(args []string, out, errs io.Writer, stamp string) int {
	return tool(stamp).Run(args, out, errs)
}

// tool is nilgai and its verbs, for the build stamped as given.
func tool(stamp string) *pzm.Tool {
	made := pzm.New("nilgai", "read radio, gamma, X-ray and infrared "+
		"catalogues as one record", pzm.Build(stamp), pzm.Built(built),
		pzm.Who(who), pzm.Sessions(sessions))
	reads := verb.NewFacts(verb.WritesNothing, verb.LeavesAlone,
		verb.ThisMachine, verb.NoOne, verb.Light)
	made.Verb("version", 'v', "say which build this is", versionPage,
		"Example_version", reads, stampedAs(stamp).version)
	return made
}

const versionPage = `# version

says the commit this binary was built from, or "unstamped" when game
did not build it. it reads nothing and writes only to standard output.`

// stampedAs is the commit game stamped a build with, or empty.
type stampedAs string

// version says the build.
func (stamp stampedAs) version(call pzm.Call) error {
	said := string(stamp)
	if said == "" {
		said = unstamped
	}
	_, err := fmt.Fprintf(call.Out(), "nilgai %s\n", said)
	return err
}
