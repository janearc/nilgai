package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/astrogo/fitsio"
	"github.com/janearc/nilgai/internal/fitstest"
	"go.uber.org/mock/gomock"
)

// Example_screen says what a screen would do, and reads nothing.
func Example_screen() {
	code, out, _ := firangRan(metal{}, published{}, "screen", "--numbers",
		numbersPath, "--gaia", "gaia.fits", "--into", "screen.tsv",
		"--dry-run")
	fmt.Print(code, " ", out)
	// Output:
	// 0 would read gaia.fits, run every star kept at a parallax of 5 times its error back 1e+07 years with each estimate of the sun's speed, and write those passing within 1000 pc to screen.tsv
}

// Example_draws says what the draws would do, and reads nothing.
func Example_draws() {
	code, out, _ := firangRan(metal{}, published{}, "draws", "--numbers",
		numbersPath, "--gaia", "gaia.fits", "--screened", "screen.tsv",
		"--into", "draws.tsv", "--dry-run")
	fmt.Print(code, " ", out)
	// Output:
	// 0 would run each star in screen.tsv back 10 times across its errors and the sun's speed, from gaia.fits, and write how often each passes within 100 pc to draws.tsv
}

// Example_validate says what the validation would do, and reads
// nothing.
func Example_validate() {
	code, out, _ := firangRan(metal{}, published{}, "validate",
		"--numbers", numbersPath, "--gaia", "gaia.fits", "--screened",
		"screen.tsv", "--into", "validate.tsv", "--dry-run")
	fmt.Print(code, " ", out)
	// Output:
	// 0 would throw 40 stars from sgr a* at 500 to 1500 km/s, 0.5 to 9 Myr ago, see them with the errors of stars in screen.tsv from gaia.fits, draw each 10 times, and write them to validate.tsv
}

// Example_candidates says what the candidates would do, and reads
// nothing.
func Example_candidates() {
	code, out, _ := firangRan(metal{}, published{}, "candidates",
		"--numbers", numbersPath, "--drawn", "draws.tsv", "--into",
		"candidates", "--dry-run")
	fmt.Print(code, " ", out)
	// Output:
	// 0 would write a folder in candidates for each star in draws.tsv at 400 km/s or more, at least 0.2 of its draws within 100 pc, its radial velocity good to 5 km/s
}

// refused is a call that cannot run: what it opens with, what it reads
// the ledger from, its arguments, and the words its refusal says.
type refused struct {
	open opener
	read ledger
	args []string
	said string
}

// refusing is what the refusal tests share: a product, a table that is
// not the query's, a screen's table and an empty one, a path nothing can
// be written to, ledgers without a galaxy or a frame, and openers that
// cannot open, whose solver refuses, or that must not be called.
type refusing struct {
	product, sizes, screened, empty, blocked string
	noGalaxy, noFrame                        ledger
	noSolver, refuses, unused                opener
}

// refusals is what the refusal tests share, made for the test given.
func refusals(test *testing.T) refusing {
	controller := gomock.NewController(test)
	noGalaxy, noFrame := NewMockledger(controller), NewMockledger(controller)
	noGalaxy.EXPECT().Galaxy().Return(nil, errors.New("no galaxy")).
		AnyTimes()
	noFrame.EXPECT().Galaxy().DoAndReturn(published{}.Galaxy).AnyTimes()
	noFrame.EXPECT().Frame().Return(nil, errors.New("no frame")).
		AnyTimes()
	noSolver := NewMockopener(controller)
	noSolver.EXPECT().Open(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("no gpu here")).AnyTimes()
	return refusing{product: productAt(test, centreRow(centreID)),
		sizes:    sizesAt(test),
		screened: screenedAt(test, screened{row: firstRow, id: centreID}),
		empty:    tableOf(test, strings.Join(screenColumns, "\t")+"\n"),
		blocked:  filepath.Join(test.TempDir(), "none", "out.tsv"),
		noGalaxy: noGalaxy, noFrame: noFrame, noSolver: noSolver,
		refuses: refusingOpener(test), unused: NewMockopener(controller)}
}

// refusedAll runs each call, which must be refused with its words and
// write nothing to standard output.
func refusedAll(test *testing.T, cases map[string]refused) {
	test.Helper()
	for name, each := range cases {
		code, out, errs := firangRan(each.open, each.read, each.args...)
		if code == exitRan || out != "" || !strings.Contains(errs, each.said) {
			test.Errorf("%s: exit %d, %q, %s", name, code, out, errs)
		}
	}
}

// the anti-test: a screen that cannot run is refused with its words:
// without its numbers or its product, its numbers not there, its
// product not there or not the query's, a ledger without a galaxy or a
// frame, a solver that cannot open or that refuses, and a file it
// cannot write.
func TestAScreenThatCannotRunIsRefused(test *testing.T) {
	at := refusals(test)
	noYaml, noFits := intoA(test, "none.yaml"), intoA(test, "none.fits")
	screen := []string{"screen", "--numbers", numbersPath, "--into",
		at.blocked}
	refusedAll(test, map[string]refused{
		"no numbers": {at.unused, published{}, []string{"screen", "--gaia",
			at.product, "--into", at.blocked}, "firang needs --numbers"},
		"no gaia": {at.unused, published{}, screen, "firang needs --gaia"},
		"numbers not there": {at.unused, published{}, []string{"screen",
			"--numbers", noYaml, "--gaia", at.product,
			"--into", at.blocked}, "no such file"},
		"not there": {at.unused, published{}, append(screen, "--gaia",
			noFits), "no such file"},
		"sizes": {at.unused, published{}, append(screen, "--gaia", at.sizes),
			"no source_id"},
		"galaxy": {at.unused, at.noGalaxy, append(screen, "--gaia",
			at.product), "no galaxy"},
		"frame": {at.unused, at.noFrame, append(screen, "--gaia",
			at.product), "no frame"},
		"solver": {at.noSolver, published{}, append(screen, "--gaia",
			at.product), "no gpu here"},
		"refused": {at.refuses, published{}, append(screen, "--gaia",
			at.product), "the device is gone"},
		"into": {running(test, &farFromCentre{}), published{},
			append(screen, "--gaia", at.product), "no such file"}})
}

// the anti-test: draws that cannot run are refused with their words:
// without their product, the product not the query's, a ledger without
// a galaxy, a solver that refuses, and a file they cannot write.
func TestDrawsThatCannotRunAreRefused(test *testing.T) {
	at := refusals(test)
	draws := []string{"draws", "--numbers", numbersPath, "--screened",
		at.screened, "--into", at.blocked}
	refusedAll(test, map[string]refused{
		"no gaia": {at.unused, published{}, draws, "firang needs --gaia"},
		"sizes": {at.unused, published{}, append(draws, "--gaia", at.sizes),
			"no source_id"},
		"galaxy": {at.unused, at.noGalaxy, append(draws, "--gaia",
			at.product), "no galaxy"},
		"refused": {at.refuses, published{}, append(draws, "--gaia",
			at.product), "the device is gone"},
		"into": {running(test, &farFromCentre{}), published{},
			append(draws, "--gaia", at.product), "no such file"}})
}

// the anti-test: a validation that cannot run is refused with its
// words: without its product, its screen's table not there or with no
// stars to lend their errors, the product not the query's, a ledger
// without a galaxy, a solver that refuses, and a file it cannot write.
func TestAValidationThatCannotRunIsRefused(test *testing.T) {
	at := refusals(test)
	noTable := intoA(test, "none.tsv")
	validate := []string{"validate", "--numbers", numbersPath,
		"--screened", at.screened, "--into", at.blocked}
	without := []string{"validate", "--numbers", numbersPath, "--gaia",
		at.product, "--into", at.blocked, "--screened"}
	refusedAll(test, map[string]refused{
		"no gaia": {at.unused, published{}, validate, "firang needs --gaia"},
		"screened": {at.unused, published{}, append(without, noTable),
			"no such file"},
		"lenders": {at.unused, published{}, append(without, at.empty),
			"has no stars to lend"},
		"sizes": {at.unused, published{}, append(validate, "--gaia",
			at.sizes), "no source_id"},
		"galaxy": {at.unused, at.noGalaxy, append(validate, "--gaia",
			at.product), "no galaxy"},
		"refused": {at.refuses, published{}, append(validate, "--gaia",
			at.product), "the device is gone"},
		"into": {running(test, &farFromCentre{}), published{},
			append(validate, "--gaia", at.product), "no such file"}})
}

// the anti-test: candidates without their numbers are refused.
func TestCandidatesWithoutTheirNumbersAreRefused(test *testing.T) {
	at := refusals(test)
	refusedAll(test, map[string]refused{"no numbers": {at.unused,
		published{}, []string{"candidates", "--drawn", at.empty, "--into",
			at.blocked}, "firang needs --numbers"}})
}

// sizesAt writes a fits table that is not the query's, one column of
// sizes, and gives its path.
func sizesAt(test *testing.T) string {
	test.Helper()
	path := intoA(test, "sizes.fits")
	file, err := os.Create(path)
	if err != nil {
		test.Fatal(err)
	}
	defer file.Close()
	columns := []fitsio.Column{{Name: "size", Format: "E"}}
	if err := fitstest.Write(file, "sizes", columns, &sized{}); err != nil {
		test.Fatal(err)
	}
	return path
}

// sized is a row of a table that is not the query's.
type sized struct {
	Size float32 `fits:"size"`
}

// refusingOpener is a mocked opener whose solver refuses every call,
// not as a runaway; it may be opened and closed any number of times.
func refusingOpener(test *testing.T) *Mockopener {
	controller := gomock.NewController(test)
	open, made := NewMockopener(controller), NewMocksolver(controller)
	open.EXPECT().Open(gomock.Any(), gomock.Any()).Return(made, nil).
		AnyTimes()
	made.EXPECT().Close().AnyTimes()
	made.EXPECT().Back(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil, errDevice).AnyTimes()
	return open
}
