package main

import (
	"bytes"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/janearc/celestial/measure"
	"github.com/janearc/celestial/orbit"
	"github.com/janearc/celestial/research"
	"github.com/janearc/nilgai/internal/fitstest"
	"go.uber.org/mock/gomock"
)

// The tests' own numbers, each defined.
const (
	numbersPath = "testdata/numbers.yaml" // defined: the tests' numbers
	smallError  = 0.01                    // defined: mas, mas a year: a draw moves little
	tinyError   = 1e-6                    // defined: mas, a parallax at the centre known to a parsec
	radialError = 1.0                     // defined: km/s
	farRA       = 10.0                    // defined: degrees, a place far from the centre
	farDec      = 10.0                    // defined: degrees
	farParallax = 1.0                     // defined: mas, a kiloparsec from the sun
	pastLight   = 5e5                     // defined: km/s, which the frame refuses
	poorError   = 1.0                     // defined: mas, a parallax that is noise
	poorMas     = 0.1                     // defined: mas, under five times its error
	centreID    = 1                       // defined: the star at the centre
	farID       = 2                       // defined: the star far from it
	poorID      = 3                       // defined: the star whose parallax is noise
	fastID      = 4                       // defined: the star past light
	exitRan     = 0                       // defined: pzm's exit when a verb ran
	firstLine   = 0                       // defined: a table's first star
	firstRow    = 0                       // defined: a product's first row
	secondRow   = 1                       // defined: its second
	thirdRow    = 2                       // defined: its third
)

// errDevice is a solver's refusal that is not a runaway.
var errDevice = errors.New("the device is gone")

// centreParallax is Sgr A*'s parallax in milliarcseconds, at the
// ledger's distance.
func centreParallax() float64 {
	return parsecsInKpc / research.Value("sgr-a-distance")
}

// centreRow is a made-up star at Sgr A*'s place on the sky and its
// distance, as the ledger has them: a star at the centre, at rest along
// the line of sight, every error small.
func centreRow(id int64) *fitstest.GaiaStar {
	star := fitstest.NewGaiaStar(id)
	star.At(research.Value("sgr-a-ascension"),
		research.Value("sgr-a-declination"))
	star.Distant(centreParallax(), tinyError)
	star.Erring(smallError, smallError, smallError, smallError)
	star.Receding(nothing, radialError)
	return star
}

// farRow is a made-up star a kiloparsec from the Sun, far from the
// centre.
func farRow(id int64) *fitstest.GaiaStar {
	star := centreRow(id)
	star.At(farRA, farDec)
	star.Distant(farParallax, tinyError)
	return star
}

// productAt writes a product of the rows given into a folder and gives
// its path, or the test stops.
func productAt(test *testing.T, rows ...*fitstest.GaiaStar) string {
	test.Helper()
	path := filepath.Join(test.TempDir(), "gaia.fits")
	file, err := os.Create(path)
	if err != nil {
		test.Fatal(err)
	}
	defer file.Close()
	if err := fitstest.WriteGaia(file, rows...); err != nil {
		test.Fatal(err)
	}
	return path
}

// farFromCentre answers the solver: each star's nearest pass is where it
// is now, its distance from the centre, and it ends where it began. It
// keeps the sizes of the batches it was given, in order.
type farFromCentre struct{ batches []int }

// back answers a batch.
func (answers *farFromCentre) back(stars []orbit.State, _ measure.Duration,
	_ int) ([]orbit.Pass, []orbit.State, error) {
	answers.batches = append(answers.batches, len(stars))
	var passes []orbit.Pass
	for _, star := range stars {
		x, y, z := star.Numbers().At()
		away := math.Sqrt(x*x + y*y + z*z)
		passes = append(passes, orbit.NewPass(away, nothing, nothing,
			nothing))
	}
	return passes, stars, nil
}

// running is a mocked opener whose solver answers as the answers given,
// opened once and closed once.
func running(test *testing.T, answers *farFromCentre) *Mockopener {
	controller := gomock.NewController(test)
	open, made := NewMockopener(controller), NewMocksolver(controller)
	open.EXPECT().Open(gomock.Any(), gomock.Any()).Return(made, nil)
	made.EXPECT().Close()
	made.EXPECT().Back(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(answers.back).AnyTimes()
	return open
}

// firangRan runs firang with the arguments, opening its solver and
// reading its ledger as given, and gives the exit code and what it
// wrote and said.
func firangRan(open opener, read ledger, args ...string) (int, string,
	string) {
	var out, errs bytes.Buffer
	code := toolWith("", open, read).RunWith(args, strings.NewReader(""),
		&out, &errs)
	return code, out.String(), errs.String()
}

// written is a table a verb wrote, read back by hand: its notes, its
// columns and its lines, each split at tabs.
type written struct {
	notes   []string
	columns []string
	lines   [][]string
}

// writtenAt reads the table at the path given, or the test stops.
func writtenAt(test *testing.T, path string) written {
	test.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		test.Fatal(err)
	}
	var found written
	text := strings.TrimSpace(string(raw))
	for _, line := range strings.Split(text, "\n") {
		switch {
		case strings.HasPrefix(line, "# "):
			found.notes = append(found.notes, strings.TrimPrefix(line, "# "))
		case found.columns == nil:
			found.columns = strings.Split(line, "\t")
		default:
			found.lines = append(found.lines, strings.Split(line, "\t"))
		}
	}
	return found
}

// field is a line's field under a column, or the test stops.
func (table written) field(test *testing.T, line []string,
	column string) string {
	test.Helper()
	for at, name := range table.columns {
		if name == column {
			return line[at]
		}
	}
	test.Fatalf("no column %s in %v", column, table.columns)
	return ""
}

// intoA is a path in a new folder for a verb to write to.
func intoA(test *testing.T, name string) string {
	return filepath.Join(test.TempDir(), name)
}
