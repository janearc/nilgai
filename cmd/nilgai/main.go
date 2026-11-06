// Command nilgai draws catalogue searches with the show verb. Its command
// interface uses pzm, which supplies --help, --agent-help, --age, and --who.
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

// main runs the command with process arguments and standard streams.
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
	return toolWith(stamp, realSearches(), band.NewEmbedded(), metal{},
		published{}, orbitsNumbers)
}

// toolWith builds the command with supplied searches, colour profiles
// and a way to open the orbit solver. Tests substitute these
// dependencies to check show and orbits without real files or a gpu.
func toolWith(stamp string, searches []search,
	profiles band.ProfileSource, open opener, read ledger,
	raw []byte) *pzm.Tool {
	made := pzm.New("nilgai", "read radio, gamma, X-ray and infrared "+
		"catalogues as one record", pzm.Build(stamp), pzm.Built(built),
		pzm.Who(who), pzm.Sessions(sessions))
	reads := verb.NewFacts(verb.WritesNothing, verb.LeavesAlone,
		verb.ThisMachine, verb.NoOne, verb.Light)
	runs := verb.NewFacts(verb.WritesTo("standard output"), verb.LeavesAlone,
		verb.ThisMachine, verb.NoOne, verb.FullGPU)
	made.Verb("orbits", 'o', "run the fastest gaia stars back to their "+
		"nearest pass by sgr a*", orbitsPage, "Example_orbits", runs,
		orbitsWith(open, read, raw),
		pzm.Value("gaia", 'g', "the gaia product, a fits binary table"),
		pzm.Value("sun", verb.NoShort, "frame or model: whose speed the "+
			"frame's sun goes at"),
		pzm.Value("fastest", verb.NoShort, "the share of the kept stars "+
			"run, the fastest first"),
		pzm.Value("quality", verb.NoShort, "the least parallax over its "+
			"error a star is kept at"),
		pzm.Value("span", verb.NoShort, "how many years back"),
		pzm.Value("steps", verb.NoShort, "in how many steps"),
		pzm.Value("batch", verb.NoShort, "how many stars at once"))
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

const orbitsPage = `# orbits

reads the gaia product named by --gaia, the rows of gaia_source with a
radial velocity as celestial-data fetched them, keeps the stars whose
parallax is at least --quality times its error, selects the fastest
--fastest of them by their speed against the sun, and runs those back
--span years in --steps steps through the milky way, --batch at a time,
on the gpu. it writes one line a star to standard output, sorted by the
star's nearest pass by sgr a*: the id, the pass in parsecs and how long
ago in millions of years, the pass's step shares, the speed, and the
star as the catalogue had it. a star the frame refuses or whose orbit
ran away is left out and named in the header.

--sun is a choice and has no default: the frame's sun goes 251.5 km/s
along the rotation as published, the model's at the sun's distance
about 231; frame runs the frame as published, model gives the frame's
sun the model's speed. the numbers the other flags default to are in
orbits.yaml, each with its reason.

    nilgai orbits --gaia ~/sky/gaia/gaia_source_0_0_r180.fits \
      --sun frame > candidates.tsv`
