// Command firang runs firang, the search for stars thrown out of the
// galactic centre: it screens a Gaia product on the GPU, draws the
// screened stars across their errors, validates the method with stars
// thrown from Sgr A*, and writes the candidates out.
//
// Its interface is pzm's, which supplies --help, --agent-help, --age and
// --who.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/janearc/puzzlemouth/pzm"
	"github.com/janearc/puzzlemouth/verb"
)

// build, built, who and sessions are the stamps game build writes into a
// command's main package.
var build, built, who, sessions string

// afterName is where the arguments begin in os.Args. The program's own
// name comes first.
const afterName = 1 // defined: the arguments after the program's name

// main runs the command with process arguments and standard streams.
func main() {
	os.Exit(run(os.Args[afterName:], os.Stdin, os.Stdout, os.Stderr,
		build))
}

// run calls firang with the arguments, reading from in and writing to
// out and errs, for the build stamped as given. It returns the exit
// code.
func run(args []string, in io.Reader, out, errs io.Writer,
	stamp string) int {
	return tool(stamp).RunWith(args, in, out, errs)
}

// tool returns firang and its verbs, for the build stamped as given,
// running on the Metal kernel by celestial's ledger.
func tool(stamp string) *pzm.Tool {
	return toolWith(stamp, metal{}, published{})
}

// runner holds what a verb opens its solver with and reads its ledger
// from. Tests give their own.
type runner struct {
	open opener
	read ledger
}

// toolWith builds the command with the solver and ledger given.
func toolWith(stamp string, open opener, read ledger) *pzm.Tool {
	made := pzm.New("firang", "find stars thrown out of the galactic "+
		"centre", pzm.Build(stamp), pzm.Built(built), pzm.Who(who),
		pzm.Sessions(sessions))
	runs := runner{open: open, read: read}
	file := verb.NewFacts(verb.WritesTo("the file --into names"),
		verb.Changes, verb.ThisMachine, verb.NoOne, verb.FullGPU)
	folders := verb.NewFacts(verb.WritesTo("folders in --into"),
		verb.Changes, verb.ThisMachine, verb.NoOne, verb.Light)
	numbers := pzm.Value("numbers", 'n', "the search's numbers file")
	product := pzm.Value("gaia", 'g', "the gaia product, a fits binary "+
		"table")
	into := pzm.Value("into", 'i', "where to write")
	screenedFlag := pzm.Value("screened", 's', "the screen's table")
	made.Verb("screen", verb.NoShort, "run every kept star back with "+
		"both sun speeds", screenPage, "Example_screen", file, runs.screen,
		numbers, product, into)
	made.Verb("draws", verb.NoShort, "run the screened stars back across "+
		"their errors", drawsPage, "Example_draws", file, runs.draws,
		numbers, product, screenedFlag, into)
	made.Verb("validate", verb.NoShort, "throw stars from sgr a* and "+
		"draw them as gaia would see them", validatePage,
		"Example_validate", file, runs.validate, numbers, product,
		screenedFlag, into)
	made.Verb("candidates", verb.NoShort, "write a folder for each "+
		"candidate in the draws", candidatesPage, "Example_candidates",
		folders, runs.candidates, numbers,
		pzm.Value("drawn", 'd', "the draws' table"), into)
	return made
}

// asked is the numbers the call names, read; it refuses a call without
// --numbers or another flag the verb needs.
func (runs runner) asked(call pzm.Call, needs ...string) (*picked,
	error) {
	for _, name := range append([]string{"numbers"}, needs...) {
		if !call.Given(name) {
			return nil, fmt.Errorf("firang needs --%s", name)
		}
	}
	return pickedFrom(call.Value("numbers"))
}

const screenPage = `# screen

reads the gaia product named by --gaia and keeps each star whose
parallax is at least the quality floor times its error and that has a
radial velocity and a proper motion. each kept star runs back through
the milky way on the gpu twice, once with the sun's speed along the
rotation as published and once at the model's, and those passing within
screen-pc of sgr a* under either are written to --into: the row, the id
and both passes. the numbers are in the file --numbers names.

    firang screen --numbers numbers.yaml \
      --gaia data/gaia/gaia_source_0_0_r180.fits --into data/screen.tsv`

const drawsPage = `# draws

runs each star in the screen's table, --screened, back draws times,
each draw its five astrometric parameters drawn from their covariance,
its radial velocity from its error, and the sun's speed along the
rotation drawn between the two estimates. it writes to --into how often
each star's draws pass within near-pc and within screen-pc, its passes'
middle and spread, and where it is and how fast about the centre now,
the stars most often near first.

    firang draws --numbers numbers.yaml --screened data/screen.tsv \
      --gaia data/gaia/gaia_source_0_0_r180.fits --into data/draws.tsv`

const validatePage = `# validate

throws stars from a parsec off sgr a*, straight out, at speeds and for
flights drawn in the numbers' ranges, and carries them forward on the
gpu. each is seen from the sun as gaia would see it, with the errors of
a star from the screen's table, --screened, and one draw of them, kept
only if its parallax clears the floor. then each is drawn as the draws
verb draws a real star. what they score is what a star thrown from the
centre would score with errors like these.

    firang validate --numbers numbers.yaml --screened data/screen.tsv \
      --gaia data/gaia/gaia_source_0_0_r180.fits \
      --into data/validate.tsv`

const candidatesPage = `# candidates

reads the draws' table, --drawn, and writes a folder in --into for each
star moving at least candidate-speed through the galaxy, with at least
candidate-share of its draws within near-pc, and its radial velocity
good to candidate-rv-error: the folder named by its id, holding a
readme of its numbers. the ids go to standard output.

    firang candidates --numbers numbers.yaml --drawn data/draws.tsv \
      --into candidates`
