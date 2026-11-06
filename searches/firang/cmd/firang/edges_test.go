package main

import (
	"bytes"
	"errors"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/janearc/celestial-data/gaia"
	gpuorbit "github.com/janearc/celestial/gpu/orbit"
	"github.com/janearc/celestial/measure"
	"github.com/janearc/celestial/orbit"
	"github.com/janearc/celestial/research"
	"github.com/janearc/nilgai/internal/fitstest"
	"go.uber.org/mock/gomock"
)

// The edges tests' own numbers, each defined.
const (
	nearThePole = 89.9999999 // defined: degrees, 0.36 mas short of it
	wideDec     = 10.0       // defined: mas, past the pole in one step
	longLine    = 70000      // defined: characters, past a scanner's line
	noSuns      = -1         // defined: a hole of less than no mass
	aStamp      = "a stamp"  // defined: a build's stamp, for run
	fiveTimes   = 5          // defined: the tests' quality floor
	belowNone   = -1         // defined: a floor under none
	pastWhole   = 2          // defined: a share past the whole
	oneDraw     = 1          // defined: a draw
	draws       = 10         // defined: the tests' draws a star
)

// sourcesOf is the sources of a product of the rows given, in order, or
// the test stops.
func sourcesOf(test *testing.T, rows ...*fitstest.GaiaStar) []*gaia.Source {
	test.Helper()
	catalogue, file, err := opened(productAt(test, rows...))
	if err != nil {
		test.Fatal(err)
	}
	defer file.Close()
	var sources []*gaia.Source
	for row := range catalogue.Rows() {
		source, err := catalogue.Source(row)
		if err != nil {
			test.Fatal(err)
		}
		sources = append(sources, source)
	}
	return sources
}

// the test and the anti-test: a star is kept with a positive parallax
// at least the floor times its error, and a radial velocity and a proper
// motion each a number; at the floor exactly it is kept; with any of
// them missing, or its parallax under the floor or none, it is not.
func TestAStarIsKeptOnlyWithEveryNumber(test *testing.T) {
	good, floor := centreRow(centreID), centreRow(centreID)
	floor.Distant(poorError*fiveTimes, poorError)
	under, none := centreRow(centreID), centreRow(centreID)
	under.Distant(poorMas, poorError)
	none.Distant(nothing, tinyError)
	noRadial, noRA, noDec := centreRow(centreID), centreRow(centreID),
		centreRow(centreID)
	noRadial.Receding(math.NaN(), radialError)
	noRA.Moving(math.NaN(), nothing)
	noDec.Moving(nothing, math.NaN())
	sources := sourcesOf(test, good, floor, under, none, noRadial, noRA,
		noDec)
	wants := []bool{true, true, false, false, false, false, false}
	quality := testNumbers(test).quality
	for which, source := range sources {
		if got := kept(source, quality); got != wants[which] {
			test.Errorf("star %d is kept %v, want %v", which, got,
				wants[which])
		}
	}
}

// pickedFromText writes the numbers' text to a file and reads them back.
func pickedFromText(test *testing.T, text string) (*picked, error) {
	test.Helper()
	path := intoA(test, "numbers.yaml")
	if err := os.WriteFile(path, []byte(text), fileBit); err != nil {
		test.Fatal(err)
	}
	return pickedFrom(path)
}

// renamed is the tests' numbers with one renamed, so it is missing.
func renamed(test *testing.T, name string) string {
	test.Helper()
	raw, err := os.ReadFile(numbersPath)
	if err != nil {
		test.Fatal(err)
	}
	return strings.Replace(string(raw), "name: "+name+"\n", "name: gone-"+
		name+"\n", oneLine)
}

// the anti-test: numbers that cannot be read are refused with their
// words: not yaml, and a number of each kind missing.
func TestNumbersThatCannotBeReadAreRefused(test *testing.T) {
	for said, text := range map[string]string{
		"numbers.yaml":    "numbers: [",
		"near-pc":         renamed(test, "near-pc"),
		"draws":           renamed(test, "draws"),
		"validation-seed": renamed(test, "validation-seed"),
	} {
		if by, err := pickedFromText(test, text); err == nil || by != nil ||
			!strings.Contains(err.Error(), said) {
			test.Errorf("%s: %v, %v", said, by, err)
		}
	}
}

// the anti-test: numbers that cannot run are refused with their words:
// each count and length of none, a floor below none, a share past the
// whole, speeds and flights the wrong way round, a flight past the span.
func TestNumbersThatCannotRunAreRefused(test *testing.T) {
	base := testNumbers(test)
	steps, batch, drawn, thrownNone := base, base, base, base
	steps.steps, batch.batch, drawn.draws, thrownNone.thrown = nothing,
		nothing, nothing, nothing
	span, screen, near, from, step := base, base, base, base, base
	span.years, screen.screenPc, near.nearPc = nothing, nothing, nothing
	from.fromPc, step.flightStep = nothing, nothing
	floor, share := base, base
	floor.quality, share.share = belowNone, pastWhole
	speeds, flights, past := base, base, base
	speeds.least = speeds.most + whole
	flights.flightLeast = flights.flightMost + whole
	past.flightMost = past.years
	cases := map[string]picked{"steps is 0": steps, "batch is 0": batch,
		"draws is 0": drawn, "thrown is 0": thrownNone,
		"span-years is 0": span, "screen-pc is 0": screen,
		"near-pc is 0": near, "thrown-from-pc is 0": from,
		"flight-step-myr is 0": step, "quality-floor -1": floor,
		"candidate-share 2": share, "thrown 1501 to 1500": speeds,
		"flights 10 to 9": flights, "flights 0.5 to 1e+07": past}
	for said, by := range cases {
		if err := by.check(); err == nil || !strings.Contains(err.Error(),
			said) {
			test.Errorf("%s: %v", said, err)
		}
	}
	if err := base.check(); err != nil {
		test.Errorf("the tests' numbers were refused: %v", err)
	}
}

// the test: run gives the exit of the call it makes, the real tool
// given a stamp: a dry run ran.
func TestRunGivesTheExitOfTheCall(test *testing.T) {
	var out, errs bytes.Buffer
	code := run([]string{"screen", "--numbers", numbersPath, "--gaia", "g",
		"--into", "i", "--dry-run"}, strings.NewReader(""), &out, &errs,
		aStamp)
	if code != exitRan || !strings.HasPrefix(out.String(), "would read g") {
		test.Errorf("exit %d: %q %s", code, out.String(), errs.String())
	}
}

// the anti-test: a hole of less than no mass makes no galaxy.
func TestAHoleOfLessThanNoMassMakesNoGalaxy(test *testing.T) {
	if galaxy, err := milkyWay(noSuns); err == nil || galaxy != nil {
		test.Errorf("a hole of less than no mass gave %v, %v", galaxy, err)
	}
}

// pastLight is a setting whose Sun would go past light at either
// estimate, over the published frame.
func pastLightSetting(test *testing.T) setting {
	test.Helper()
	frame, err := published{}.Frame()
	if err != nil {
		test.Fatal(err)
	}
	return setting{frame: frame, low: pastLight, high: pastLight,
		batch: fullBatch}
}

// the anti-test: a setting whose Sun would go past light has no frames
// to screen in or draw in: the screen and the draws each stop with the
// frame's refusal.
func TestASunPastLightStopsTheScreenAndTheDraws(test *testing.T) {
	at := pastLightSetting(test)
	product := NewMockproduct(gomock.NewController(test))
	by := testNumbers(test)
	if found, err := at.screen(product, by); err == nil || found != nil {
		test.Errorf("the screen gave %v, %v", found, err)
	}
	sources := sourcesOf(test, centreRow(centreID))
	spread, err := sources[firstLine].Spread()
	if err != nil {
		test.Fatal(err)
	}
	if states, _, err := at.drawn(spread, oneDraw, seeded()); err == nil ||
		states != nil {
		test.Errorf("the draws gave %v, %v", states, err)
	}
	stars := []drawnStar{{source: sources[firstLine]}}
	by = testNumbers(test)
	if drawn, _, err := at.drawRun(stars, by, seeded()); err == nil ||
		drawn != nil {
		test.Errorf("the drawn run gave %v, %v", drawn, err)
	}
}

// the anti-test: a row the catalogue cannot give stops the screen, the
// draws and the validation's sight of a star.
func TestARowTheCatalogueCannotGiveStopsTheRun(test *testing.T) {
	controller := gomock.NewController(test)
	product := NewMockproduct(controller)
	product.EXPECT().Rows().Return(int64(oneLine)).AnyTimes()
	product.EXPECT().Source(gomock.Any()).Return(nil,
		errors.New("no row there")).AnyTimes()
	frame, err := published{}.Frame()
	if err != nil {
		test.Fatal(err)
	}
	at := setting{frame: frame, low: whole, high: whole, batch: fullBatch}
	by := testNumbers(test)
	if found, err := at.screen(product, by); err == nil || found != nil {
		test.Errorf("the screen gave %v, %v", found, err)
	}
	stars := []screened{{row: firstRow, id: centreID}}
	if drawn, _, err := at.drawAll(product, stars, by, seeded()); err == nil ||
		drawn != nil {
		test.Errorf("the draws gave %v, %v", drawn, err)
	}
	near := orbit.StateOf(by.fromPc/parsecsInKpc, nothing, nothing,
		by.least, nothing, nothing)
	seen, truth, err := at.see(product, stars, thrown{}, near, by, seeded())
	if err == nil || seen != nil || truth != nil {
		test.Errorf("the validation saw %v, %v, %v", seen, truth, err)
	}
	answering := refusingSetting(test, by.thrown)
	if found, _, err := answering.validateAll(product, stars, by,
		seeded()); err == nil || found != nil {
		test.Errorf("the validation gave %v, %v", found, err)
	}
}

// refusingAfter is a solver that answers as farFromCentre does for so
// many calls, then refuses every call after, not as a runaway.
type refusingAfter struct {
	answers farFromCentre
	left    int
}

// back answers while calls are left, then refuses.
func (solver *refusingAfter) back(stars []orbit.State, span measure.Duration,
	steps int) ([]orbit.Pass, []orbit.State, error) {
	if solver.left <= nothing {
		return nil, nil, errDevice
	}
	solver.left--
	return solver.answers.back(stars, span, steps)
}

// refusingSetting is a setting over the published frame whose solver
// answers so many calls, then refuses.
func refusingSetting(test *testing.T, calls int) setting {
	test.Helper()
	made := NewMocksolver(gomock.NewController(test))
	solver := &refusingAfter{left: calls}
	made.EXPECT().Back(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(solver.back).AnyTimes()
	frame, err := published{}.Frame()
	if err != nil {
		test.Fatal(err)
	}
	return setting{frame: frame, low: whole, high: whole, made: made,
		batch: fullBatch, steps: draws, span: measure.Years(inAMillion)}
}

// callsOf is how many calls the validation of the tests' throw makes of
// the solver before it draws: one a batch for each flight's stars, then
// one a batch for the true passes of every star thrown.
func callsOf(by picked) (int, int) {
	flights := map[float64]int{}
	for _, star := range throw(by, seeded()) {
		flights[star.flight]++
	}
	carried := nothing
	for _, count := range flights {
		carried += batchesOf(count, by.batch)
	}
	return carried, batchesOf(by.thrown, by.batch)
}

// batchesOf is how many batches so many stars take.
func batchesOf(count, batch int) int {
	return (count + batch - oneLine) / batch
}

// the anti-test: a solver that refuses partway stops the screen in its
// last batch and in a full one, the draws, and the validation in its
// carry, its true passes and its draws, each with the solver's words.
func TestASolverThatRefusesPartwayStopsTheRun(test *testing.T) {
	by := testNumbers(test)
	product := productAt(test, centreRow(centreID), centreRow(farID),
		centreRow(poorID), centreRow(fastID), centreRow(fiveCentres))
	catalogue, file, err := opened(product)
	if err != nil {
		test.Fatal(err)
	}
	defer file.Close()
	lenders := []screened{{row: firstRow, id: centreID}}
	carried, truths := callsOf(by)
	for name, calls := range map[string]int{"carry": nothing,
		"true passes": carried, "draws": carried + truths} {
		at := refusingSetting(test, calls)
		found, _, err := at.validateAll(catalogue, lenders, by, seeded())
		if !errors.Is(err, errDevice) || found != nil {
			test.Errorf("the validation's %s gave %v, %v", name, found, err)
		}
	}
	for name, calls := range map[string]int{"full batch": nothing,
		"last batch": oneLine} {
		at := refusingSetting(test, calls)
		if found, err := at.screen(catalogue, by); !errors.Is(err,
			errDevice) || found != nil {
			test.Errorf("the screen's %s gave %v, %v", name, found, err)
		}
	}
	at := refusingSetting(test, nothing)
	if drawn, _, err := at.drawAll(catalogue, lenders, by,
		seeded()); !errors.Is(err, errDevice) || drawn != nil {
		test.Errorf("the draws gave %v, %v", drawn, err)
	}
}

// namingFirst is a solver that names the first star of its first call
// as run away, and answers as farFromCentre does after.
type namingFirst struct {
	answers farFromCentre
	named   bool
}

// back names the first star once, then answers.
func (solver *namingFirst) back(stars []orbit.State, span measure.Duration,
	steps int) ([]orbit.Pass, []orbit.State, error) {
	if !solver.named {
		solver.named = true
		return nil, nil, gpuorbit.NewRunaway([]int{firstOfDraw})
	}
	return solver.answers.back(stars, span, steps)
}

// the anti-test: a thrown star that runs away in the carry is lost, and
// so is every star whose lender's errors cannot be drawn from.
func TestThrownStarsLostOnTheWayAreCounted(test *testing.T) {
	by := testNumbers(test)
	knotted := centreRow(centreID)
	knotted.Tied(tied, tied, -tied)
	product := productAt(test, centreRow(centreID), knotted)
	catalogue, file, err := opened(product)
	if err != nil {
		test.Fatal(err)
	}
	defer file.Close()
	made := NewMocksolver(gomock.NewController(test))
	solver := &namingFirst{}
	made.EXPECT().Back(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(solver.back).AnyTimes()
	frame, err := published{}.Frame()
	if err != nil {
		test.Fatal(err)
	}
	at := setting{frame: frame, low: whole, high: whole, made: made,
		batch: fullBatch, steps: draws, span: measure.Years(inAMillion)}
	found, lost, err := at.validateAll(catalogue, []screened{{row: firstRow,
		id: centreID}}, by, seeded())
	if err != nil || lost != oneLine || len(found) != by.thrown-oneLine {
		test.Errorf("one gone: %d found, %d lost, %v", len(found), lost, err)
	}
	found, lost, err = at.validateAll(catalogue, []screened{{row: secondRow,
		id: centreID}}, by, seeded())
	if err != nil || len(found) != nothing || lost != by.thrown {
		test.Errorf("knotted: %d found, %d lost, %v", len(found), lost, err)
	}
}

// the test and the anti-test: a star near a pole whose draws cross it
// has those draws unplaced; and a star the frame refuses, past light, is
// placed nowhere now, its distance, speed and outward speed not
// numbers.
func TestDrawsPastAPoleAreUnplacedAndAStarPastLightIsNowhere(
	test *testing.T) {
	polar := centreRow(centreID)
	polar.At(research.Value("sgr-a-ascension"), nearThePole)
	polar.Erring(smallError, wideDec, smallError, smallError)
	fast := centreRow(fastID)
	fast.Receding(pastLight, radialError)
	sources := sourcesOf(test, polar, fast)
	frame, err := published{}.Frame()
	if err != nil {
		test.Fatal(err)
	}
	at := setting{frame: frame, low: whole, high: whole}
	spread, err := sources[firstLine].Spread()
	if err != nil {
		test.Fatal(err)
	}
	states, unplaced, err := at.drawn(spread, draws, seeded())
	if err != nil || unplaced == nothing || len(states)+unplaced != draws {
		test.Errorf("%d placed, %d unplaced, %v", len(states), unplaced, err)
	}
	star := at.central(drawnStar{source: sources[secondLine]})
	if !math.IsNaN(star.away) || !math.IsNaN(star.speed) ||
		!math.IsNaN(star.outward) {
		test.Errorf("a star past light is at %g, %g, %g", star.away,
			star.speed, star.outward)
	}
}

// failingWriter is a writer that takes nothing.
type failingWriter struct{}

// Write refuses every write.
func (failingWriter) Write([]byte) (int, error) {
	return nothing, errors.New("nothing written")
}

// the anti-test and the test: a table is not written to a writer that
// takes nothing; a line too long to scan stops a read; a short line has
// no number and no text past its end; an id or a pass not a number
// stops the screen's table being read.
func TestTablesThatCannotBeWrittenOrReadAreRefused(test *testing.T) {
	if err := writeTable(failingWriter{}, table{notes: []string{"a"}}); err ==
		nil {
		test.Error("a table was written to a writer that takes nothing")
	}
	long := "a\n" + strings.Repeat("x", longLine) + "\n"
	if found, err := readTable(strings.NewReader(long), "long",
		nil); err == nil || found != nil {
		test.Errorf("a long line gave %v, %v", found, err)
	}
	found, err := readTable(strings.NewReader("a\tb\n1\n"), "short", nil)
	if err != nil {
		test.Fatal(err)
	}
	short := found.lines[firstLine]
	if _, err := found.number(short, "b"); err == nil ||
		!strings.Contains(err.Error(), "has no b") ||
		found.text(short, "b") != "" {
		test.Errorf("a short line gave %v and %q", err,
			found.text(short, "b"))
	}
	for said, text := range map[string]string{
		"parsing \"one\"": "0\tone\t1\t1\n", "parsing \"far\"": "0\t1\t1\tfar\n",
	} {
		path := tableOf(test, strings.Join(screenColumns, "\t")+"\n"+text)
		if stars, err := screenedFrom(path); err == nil || stars != nil ||
			!strings.Contains(err.Error(), said) {
			test.Errorf("%s: %v, %v", said, stars, err)
		}
	}
}
