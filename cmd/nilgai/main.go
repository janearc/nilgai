// Command nilgai is the command line for the nilgai library. It has one
// verb for now, show. The readers will get verbs as they are written.
// nilgai is built on pzm, which gives it --help, --agent-help, --age and
// --who.
package main

import (
	"io"
	"os"

	"github.com/janearc/nilgai/band"
	"github.com/janearc/puzzlemouth/pzm"
	"github.com/janearc/puzzlemouth/verb"
)

// build, built, who and sessions are the stamps game build writes into a
// command's main package.
var build, built, who, sessions string

// afterName is where the arguments begin in os.Args. The program's own
// name comes first.
const afterName = 1

// main runs nilgai on what it was asked.
func main() {
	os.Exit(run(os.Args[afterName:], os.Stdin, os.Stdout, os.Stderr,
		build))
}

// run calls nilgai with the arguments, reading from in and writing to out
// and errs, for the build stamped as given. It returns the exit code.
func run(args []string, in io.Reader, out, errs io.Writer,
	stamp string) int {
	return tool(stamp).RunWith(args, in, out, errs)
}

// tool returns nilgai and its verbs, for the build stamped as given.
func tool(stamp string) *pzm.Tool {
	return toolWith(stamp, realSearches(), band.NewEmbedded())
}

// toolWith returns nilgai with show reading the searches and the profiles
// given. The real tool passes the real ones, and a test passes others to
// check what show does with them.
func toolWith(stamp string, searches []search,
	profiles band.ProfileSource) *pzm.Tool {
	made := pzm.New("nilgai", "read radio, gamma, X-ray and infrared "+
		"catalogues as one record", pzm.Build(stamp), pzm.Built(built),
		pzm.Who(who), pzm.Sessions(sessions))
	reads := verb.NewFacts(verb.WritesNothing, verb.LeavesAlone,
		verb.ThisMachine, verb.NoOne, verb.Light)
	made.Verb("show", 's', "draw catalogue searches as a png", showPage,
		"Example_show", reads, showWith(profiles, searches...),
		pzm.Value("chandra", 'c', "a Chandra Source Catalog search"),
		pzm.Value("allwise", 'a', "an AllWISE search"),
		pzm.Value("fermi", 'f', "the Fermi-LAT 4FGL catalogue, FITS"),
		pzm.Value("ra", verb.NoShort, "the centre's RA, in degrees"),
		pzm.Value("dec", verb.NoShort, "the centre's Dec, in degrees"),
		pzm.List("width", 'w', "how much sky, in arcseconds; a list "+
			"gives a panel for each"),
		pzm.Value("pixels", 'p', "how many pixels across"))
	return made
}

const showPage = `# show

reads the searches named by --chandra and --allwise, as the CSC's and
IRSA's TAP services give them, and the 4FGL file named by --fermi, and
writes a png to standard output: the sky around --ra and --dec, --width
arcseconds and --pixels across. give --width a list, such as 12,6, for
panels side by side, one for each width, in the order given. giving the
flag twice is refused.

each source is its one-sigma error ellipse in its instrument's ink, with
a dot at its position, and a cross marks the centre. instruments are
told apart by how dark they are first and by hue second, and libreadme
holds every ink to the reader's contrast floor before anything is
drawn. rows that cannot be read are named after the picture is written.

    nilgai show --chandra chandra.tsv --allwise allwise.csv \
      --ra 266.4168371 --dec -29.0078106 --width 12 --pixels 600 \
      > sgra.png`
