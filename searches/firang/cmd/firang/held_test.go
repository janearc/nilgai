package main

import (
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/janearc/celestial/galactocentric"
	"github.com/janearc/celestial/length"
	"github.com/janearc/celestial/measure"
	"github.com/janearc/celestial/orbit"
	"github.com/janearc/celestial/research"
	"github.com/janearc/celestial/si"
	"github.com/janearc/celestial/sky"
	"github.com/janearc/nilgai/internal/fitstest"
	"go.uber.org/mock/gomock"
)

// The held tests' own numbers, each defined.
const (
	speedRoomKms  = 1e-9                // defined: km/s, a speed to rounding
	shareRoom     = 1e-12               // defined: a share of a whole, to rounding
	quarterWay    = 0.25                // defined: a uniform draw a quarter of the way
	halfWay       = 0.5                 // defined: half way: the sphere's equator
	threeQuarters = 0.75                // defined: three quarters round: toward -y
	noDraw        = 0.0                 // defined: a uniform or normal draw of nothing
	sixNormals    = 6                   // defined: a source's draw takes six
	lowSun        = 230.0               // defined: km/s, a low sun speed for a test
	highSun       = 250.0               // defined: km/s, a high one
	oneKpc        = 1.0                 // defined: kpc, a star a kiloparsec out
	threeKpc      = 3.0                 // defined: kpc
	twoKpcInPc    = 2000.0              // defined: pc, between them
	oneKpcInPc    = 1000.0              // defined: pc, exactly the first
	sixStars      = 6                   // defined: stars over a batch and a half
	halfBatch     = 4                   // defined: two stars, each twice
	equalOrder    = 0                   // defined: what a sort reads as equal
	twoDraws      = 2                   // defined: draws a star, four stars a batch
	longID        = 4037334576111139712 // defined: an id past what a float holds to the unit
	folderMode    = 0o755               // defined: a folder its owner writes and all read
	fileMode      = 0o644               // defined: a file its owner writes and all read
	speedCut      = 400                 // defined: km/s, the tests' candidate speed
	shareCut      = 0.2                 // defined: the tests' candidate share
	rvCut         = 5                   // defined: km/s, the tests' radial velocity error
	distanceRoom  = 1e-3                // defined: a share, a distance read back from a parallax drawn once
	qualityRoom   = 1e-2                // defined: a share, a parallax's quality the same
	float32Room   = 1e-7                // defined: a share, a float32's rounding
	twoMas        = 2.0                 // defined: mas, a star half a kiloparsec off
	underLight    = 10                  // defined: km/s, under light with the model's Sun, over it with the frame's
	// rotationLongitude is the galactic longitude the Sun moves toward,
	// defined by the galactic frame: 90 degrees.
	rotationLongitude = 90 // defined: degrees
)

// pair is two numbers held against each other: one found and the one it
// is held to, or a number wanted and the room it is wanted within.
type pair struct{ one, other float64 }

// kmsOf is a state's speed, in kilometres per second.
func kmsOf(state orbit.State) float64 {
	vx, vy, vz := state.Numbers().Moving()
	return math.Sqrt(vx*vx + vy*vy + vz*vz)
}

// the test: the setting's two estimates of the Sun's speed along the
// rotation are the published frame's and the model's, worked by hand:
// the model's circular speed at the Sun's distance, the root of the
// distance times the pull toward the centre there, plus the Sun's own
// motion; the lower is low and the higher high; the motion toward the
// centre and toward the pole are the frame's; and the frames the
// estimates give go at each.
func TestTheSettingHoldsBothSunSpeeds(test *testing.T) {
	controller := gomock.NewController(test)
	open := NewMockopener(controller)
	open.EXPECT().Open(gomock.Any(), gomock.Any()).Return(
		NewMocksolver(controller), nil)
	at, err := settingFor(published{}, open, testNumbers(test))
	if err != nil {
		test.Fatal(err)
	}
	frame, err := published{}.Frame()
	if err != nil {
		test.Fatal(err)
	}
	galaxy, err := published{}.Galaxy()
	if err != nil {
		test.Fatal(err)
	}
	place, moving := frame.Sun()
	by := place.Minus(measure.Origin())
	distance := -by.AlongX().InMetres() / si.Kilo(length.Parsec)
	pull, _, _ := galaxy.Pull(orbit.PlaceOf(distance, nothing,
		nothing)).Along()
	model := math.Sqrt(distance*-pull) + research.Value("sun-motion-v")
	measured := moving.AlongY().InKilometresPerSecond()
	wants := map[string]pair{"low": {at.low, model},
		"high":      {at.high, measured},
		"to centre": {at.toCentre, moving.AlongX().InKilometresPerSecond()},
		"to pole":   {at.toPole, moving.AlongZ().InKilometresPerSecond()}}
	for name, pair := range wants {
		if !(math.Abs(pair.one-pair.other) <= speedRoomKms) {
			test.Errorf("%s is %g km/s, want %g", name, pair.one,
				pair.other)
		}
	}
	asPublished, asModel, err := at.estimates()
	if err != nil {
		test.Fatal(err)
	}
	_, published := asPublished.Sun()
	_, modelled := asModel.Sun()
	if published.AlongY().InKilometresPerSecond() != measured ||
		!(math.Abs(modelled.AlongY().InKilometresPerSecond()-model) <=
			speedRoomKms) {
		test.Errorf("the estimates go at %v and %v", published, modelled)
	}
}

// drawsOf is a mocked source of draws: six normal ones of nothing for a
// star's errors, then a uniform one as given.
func drawsOf(test *testing.T, uniform float64) *Mockdrawing {
	random := NewMockdrawing(gomock.NewController(test))
	random.EXPECT().NormFloat64().Return(noDraw).Times(sixNormals)
	random.EXPECT().Float64().Return(uniform)
	return random
}

// the test: a draw's Sun goes along the rotation the uniform draw's
// share of the way from the low estimate to the high: a quarter of the
// way, here, from 230 to 250 km/s, 235; the star is placed in that
// frame as the frame itself places it.
func TestADrawsSunGoesBetweenTheTwoEstimates(test *testing.T) {
	frame, err := published{}.Frame()
	if err != nil {
		test.Fatal(err)
	}
	_, moving := frame.Sun()
	at := setting{frame: frame, low: lowSun, high: highSun,
		toCentre: moving.AlongX().InKilometresPerSecond(),
		toPole:   moving.AlongZ().InKilometresPerSecond()}
	source := sourcesOf(test, centreRow(centreID))[firstLine]
	spread, err := source.Spread()
	if err != nil {
		test.Fatal(err)
	}
	states, unplaced, err := at.drawn(spread, oneDraw,
		drawsOf(test, quarterWay))
	if err != nil || unplaced != nothing || len(states) != oneDraw {
		test.Fatalf("%d drawn, %d unplaced, %v", len(states), unplaced, err)
	}
	along := lowSun + quarterWay*(highSun-lowSun)
	moved, err := frame.MovingSun(at.toCentre, along, at.toPole)
	if err != nil {
		test.Fatal(err)
	}
	want, err := source.State(*moved)
	if err != nil {
		test.Fatal(err)
	}
	if states[firstLine].Numbers() != want.Numbers() {
		test.Errorf("the draw is %v, want %v", states[firstLine], want)
	}
}

// the test: a thrown star's direction, speed and flight are its draws':
// half way up the sphere is its equator, three quarters round is toward
// -y, a draw of nothing is the least speed and the shortest flight; so
// it starts a parsec out along -y, moving in along +y at the least
// speed.
func TestAThrownStarGoesWhereItsDrawsSay(test *testing.T) {
	by := testNumbers(test)
	by.thrown = oneLine
	random := NewMockdrawing(gomock.NewController(test))
	gomock.InOrder(random.EXPECT().Float64().Return(halfWay),
		random.EXPECT().Float64().Return(threeQuarters),
		random.EXPECT().Float64().Return(noDraw),
		random.EXPECT().Float64().Return(noDraw))
	star := throw(by, random)[firstLine]
	x, y, z := star.start.Numbers().At()
	vx, vy, vz := star.start.Numbers().Moving()
	from := by.fromPc / parsecsInKpc
	got := []float64{x, y + from, z, vx, vy - by.least, vz}
	for _, each := range got {
		if !(math.Abs(each) <= placeRoom*parsecsInKpc) {
			test.Errorf("the star starts at %v", star.start)
		}
	}
	if star.speed != by.least || star.flight != by.flightLeast {
		test.Errorf("thrown at %g km/s for %g Myr", star.speed, star.flight)
	}
}

// the test: a flight too short for a step of the span's is carried in
// one step, not none.
func TestTheShortestFlightIsCarriedInOneStep(test *testing.T) {
	by := testNumbers(test)
	by.steps = oneLine
	made := NewMocksolver(gomock.NewController(test))
	solver := &carrying{steps: map[float64]int{}}
	made.EXPECT().Back(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(solver.back)
	at := setting{made: made, batch: fullBatch}
	stars := throw(testNumbers(test), seeded())[:oneLine]
	stars[firstLine].flight = by.flightLeast
	if _, _, err := at.carry(stars, by); err != nil {
		test.Fatal(err)
	}
	for _, steps := range solver.steps {
		if steps != oneLine {
			test.Errorf("the shortest flight ran %d steps", steps)
		}
	}
}

// bySpeed answers the solver with each star's speed for its pass: a
// thousandth of a parsec for each kilometre a second, so the two
// estimates of the Sun give a star two passes.
type bySpeed struct{}

// back answers a batch so.
func (solver *bySpeed) back(stars []orbit.State, _ measure.Duration,
	_ int) ([]orbit.Pass, []orbit.State, error) {
	var passes []orbit.Pass
	for _, star := range stars {
		near := kmsOf(star) / parsecsInKpc / parsecsInKpc
		passes = append(passes, orbit.NewPass(near, nothing, nothing,
			nothing))
	}
	return passes, stars, nil
}

// the test: the screen writes each star's own pass under each estimate
// of the Sun: with passes the stars' speeds, each line's two passes are
// that star's speed in each frame, worked by placing it there.
func TestTheScreenWritesEachStarsPassUnderEachSun(test *testing.T) {
	controller := gomock.NewController(test)
	open, made := NewMockopener(controller), NewMocksolver(controller)
	open.EXPECT().Open(gomock.Any(), gomock.Any()).Return(made, nil)
	made.EXPECT().Close()
	solver := &bySpeed{}
	made.EXPECT().Back(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(solver.back).AnyTimes()
	rows := []*fitstest.GaiaStar{centreRow(centreID), farRow(farID)}
	into := intoA(test, "screen.tsv")
	code, _, errs := firangRan(open, published{}, "screen", "--numbers",
		numbersPath, "--gaia", productAt(test, rows...), "--into", into)
	if code != exitRan {
		test.Fatalf("exit %d: %s", code, errs)
	}
	table := writtenAt(test, into)
	frame, model := estimatesOf(test)
	sources := sourcesOf(test, rows...)
	for which, line := range table.lines {
		for column, placing := range map[string]*galactocentric.Frame{
			"near_frame_pc": frame, "near_model_pc": model} {
			state, err := sources[which].State(*placing)
			if err != nil {
				test.Fatal(err)
			}
			want := kmsOf(*state) / parsecsInKpc
			if got := table.number(test, line, column); !(math.Abs(got-want) <=
				speedRoomKms) {
				test.Errorf("star %d's %s is %g, want %g", which, column, got,
					want)
			}
		}
	}
}

// estimatesOf is the published frame and the frame at the model's
// speed, as a setting from the ledger gives them, or the test stops.
func estimatesOf(test *testing.T) (*galactocentric.Frame,
	*galactocentric.Frame) {
	test.Helper()
	controller := gomock.NewController(test)
	open := NewMockopener(controller)
	open.EXPECT().Open(gomock.Any(), gomock.Any()).Return(
		NewMocksolver(controller), nil)
	at, err := settingFor(published{}, open, testNumbers(test))
	if err != nil {
		test.Fatal(err)
	}
	frame, model, err := at.estimates()
	if err != nil {
		test.Fatal(err)
	}
	return frame, model
}

// the test and the anti-test: a star is found when it passes within the
// radius under either estimate, not both, and not at the radius itself.
func TestAStarIsFoundWithinTheRadiusUnderEitherSun(test *testing.T) {
	made := NewMocksolver(gomock.NewController(test))
	made.EXPECT().Back(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn((&farFromCentre{}).back).AnyTimes()
	at := setting{made: made, batch: fullBatch}
	near := orbit.StateOf(oneKpc, nothing, nothing, nothing, nothing,
		nothing)
	far := orbit.StateOf(threeKpc, nothing, nothing, nothing, nothing,
		nothing)
	cases := []struct {
		framed, modelled orbit.State
		radius           float64
		found            int
	}{{near, far, twoKpcInPc, oneLine}, {far, near, twoKpcInPc, oneLine},
		{near, far, oneKpcInPc, nothing}, {far, near, oneKpcInPc, nothing}}
	for _, each := range cases {
		found := &screening{}
		batch := screenBatch{pending: []screened{{id: centreID}},
			framed:   []orbit.State{each.framed},
			modelled: []orbit.State{each.modelled}}
		if _, err := at.flush(batch, found, each.radius); err != nil ||
			len(found.found) != each.found {
			test.Errorf("within %g pc found %d, %v", each.radius,
				len(found.found), err)
		}
	}
}

// refusingOnce is a solver that refuses its first call, not as a
// runaway, and answers as farFromCentre does after.
type refusingOnce struct {
	answers farFromCentre
	refused bool
}

// back refuses once, then answers.
func (solver *refusingOnce) back(stars []orbit.State, span measure.Duration,
	steps int) ([]orbit.Pass, []orbit.State, error) {
	if !solver.refused {
		solver.refused = true
		return nil, nil, errDevice
	}
	return solver.answers.back(stars, span, steps)
}

// the anti-test: a full batch the solver refuses stops the screen, even
// when it would answer the next.
func TestAFullBatchRefusedStopsTheScreen(test *testing.T) {
	made := NewMocksolver(gomock.NewController(test))
	made.EXPECT().Back(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn((&refusingOnce{}).back).AnyTimes()
	frame, err := published{}.Frame()
	if err != nil {
		test.Fatal(err)
	}
	at := setting{frame: frame, low: lowSun, high: highSun, made: made,
		batch: fullBatch}
	var rows []*fitstest.GaiaStar
	for each := range fiveCentres {
		id := int64(each)
		rows = append(rows, centreRow(id))
	}
	catalogue, file, err := opened(productAt(test, rows...))
	if err != nil {
		test.Fatal(err)
	}
	defer file.Close()
	if found, err := at.screen(catalogue, testNumbers(test)); err == nil ||
		found != nil {
		test.Errorf("a refused full batch gave %v, %v", found, err)
	}
}

// the test: six stars, each twice, fill one batch and half another: the
// screen sends the first four when the batch is full and the last two at
// the end.
func TestSixStarsGoInAFullBatchAndAHalf(test *testing.T) {
	answers := &farFromCentre{}
	var rows []*fitstest.GaiaStar
	for each := range sixStars {
		id := int64(each)
		rows = append(rows, centreRow(id))
	}
	code, _, errs := firangRan(running(test, answers), published{}, "screen",
		"--numbers", numbersPath, "--gaia", productAt(test, rows...),
		"--into", intoA(test, "screen.tsv"))
	if code != exitRan || !slices.Equal(answers.batches, []int{fullBatch,
		halfBatch}) {
		test.Errorf("exit %d, batches %v: %s", code, answers.batches, errs)
	}
}

// the test: the draws hold a whole star's draws together in a batch:
// two stars of ten draws in batches of eight go as eight and two each,
// not as their twenty together; and the validation's draws the same.
func TestDrawsGoAStarAtATime(test *testing.T) {
	product := productAt(test, centreRow(centreID), farRow(farID))
	catalogue, file, err := opened(product)
	if err != nil {
		test.Fatal(err)
	}
	defer file.Close()
	answers := &farFromCentre{}
	made := NewMocksolver(gomock.NewController(test))
	made.EXPECT().Back(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(answers.back).AnyTimes()
	frame, err := published{}.Frame()
	if err != nil {
		test.Fatal(err)
	}
	at := setting{frame: frame, low: lowSun, high: highSun, made: made,
		batch: fullBatch}
	stars := []screened{{row: firstRow, id: centreID},
		{row: secondRow, id: farID}}
	by := testNumbers(test)
	if _, _, err := at.drawAll(catalogue, stars, by, seeded()); err != nil {
		test.Fatal(err)
	}
	starAtATime := []int{fullBatch, restOfThem, fullBatch, restOfThem}
	if !slices.Equal(answers.batches, starAtATime) {
		test.Errorf("the draws went in batches of %v", answers.batches)
	}
	answers.batches = nil
	sources := sourcesOf(test, centreRow(centreID), farRow(farID))
	found := []validated{{drawn: drawnStar{source: sources[firstLine]}},
		{drawn: drawnStar{source: sources[secondLine]}}}
	if err := at.drawValidated(found, by, seeded()); err != nil {
		test.Fatal(err)
	}
	if !slices.Equal(answers.batches, starAtATime) {
		test.Errorf("the validation's draws went in batches of %v",
			answers.batches)
	}
}

// the test: with four stars to a batch of draws, each validated star
// keeps its own draws: the source drawn is the one it was seen as.
func TestEachValidatedStarKeepsItsOwnDraws(test *testing.T) {
	by := testNumbers(test)
	by.draws = twoDraws
	made := NewMocksolver(gomock.NewController(test))
	made.EXPECT().Back(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn((&farFromCentre{}).back).AnyTimes()
	frame, err := published{}.Frame()
	if err != nil {
		test.Fatal(err)
	}
	at := setting{frame: frame, low: lowSun, high: highSun, made: made,
		batch: fullBatch}
	var found []validated
	for _, source := range sourcesOf(test, centreRow(centreID),
		farRow(farID), farRow(poorID), centreRow(fastID), farRow(fiveCentres)) {
		found = append(found, validated{lender: source.ID(),
			drawn: drawnStar{source: source}})
	}
	if err := at.drawValidated(found, by, seeded()); err != nil {
		test.Fatal(err)
	}
	for _, each := range found {
		if each.drawn.source.ID() != each.lender ||
			len(each.drawn.passes) != twoDraws {
			test.Errorf("star %d holds %d's draws", each.lender,
				each.drawn.source.ID())
		}
	}
}

// the test: the draws write each star's numbers as they are: a star at
// two milliarcseconds is half a kiloparsec off, its parallax two million
// times its error, and its speed and outward speed about the centre are
// as the frame places it.
func TestTheDrawsWriteEachStarsNumbers(test *testing.T) {
	nearer := farRow(farID)
	nearer.Distant(twoMas, tinyError)
	product := productAt(test, nearer)
	table := drawsRan(test, running(test, &farFromCentre{}), product,
		screened{row: firstRow, id: farID})
	line := table.lines[firstLine]
	frame, err := published{}.Frame()
	if err != nil {
		test.Fatal(err)
	}
	state, err := sourcesOf(test, nearer)[firstLine].State(*frame)
	if err != nil {
		test.Fatal(err)
	}
	x, y, z := state.Numbers().At()
	vx, vy, vz := state.Numbers().Moving()
	away := math.Sqrt(x*x + y*y + z*z)
	quality := twoMas / tinyError
	wants := map[string]pair{
		"distance_kpc":        {oneKpc / twoMas, shareRoom},
		"parallax_over_error": {quality, float32Room * quality},
		"v_gc_kms":            {kmsOf(*state), speedRoomKms},
		"v_out_kms":           {(x*vx + y*vy + z*vz) / away, speedRoomKms}}
	for column, want := range wants {
		if got := table.number(test, line, column); !(math.Abs(got-
			want.one) <= want.other) {
			test.Errorf("%s is %g, want %g", column, got, want.one)
		}
	}
}

// the test: the validation writes each thrown star's numbers as they
// are: carried nowhere by the mocked solver, it is a parsec from the
// centre moving at its thrown speed, as far from the Sun as Sgr A*, its
// parallax as well measured as the centre star's that lent it errors.
func TestTheValidationWritesEachStarsNumbers(test *testing.T) {
	into := intoA(test, "validate.tsv")
	product := productAt(test, centreRow(centreID))
	code, _, errs := firangRan(running(test, &farFromCentre{}),
		published{}, "validate", "--numbers", numbersPath, "--gaia",
		product, "--screened",
		screenedAt(test, screened{row: firstRow, id: centreID}), "--into",
		into)
	if code != exitRan {
		test.Fatalf("exit %d: %s", code, errs)
	}
	table := writtenAt(test, into)
	distance := research.Value("sgr-a-distance") / parsecsInKpc
	quality := centreParallax() / tinyError
	for _, line := range table.lines {
		wants := map[string]pair{
			"r_gc_kpc":            {testNumbers(test).fromPc / parsecsInKpc, placeRoom},
			"v_gc_kms":            {table.number(test, line, "thrown_kms"), speedRoomKms},
			"distance_kpc":        {distance, distanceRoom * distance},
			"parallax_over_error": {quality, qualityRoom * quality}}
		for column, want := range wants {
			if got := table.number(test, line, column); !(math.Abs(got-
				want.one) <= want.other) {
				test.Errorf("%s is %g, want %g", column, got, want.one)
			}
		}
	}
}

// the test: a star near more often comes first, whatever its middle
// pass; and two stars alike order as equal.
func TestAShareNearOrdersBeforeAMiddlePass(test *testing.T) {
	by := testNumbers(test)
	often := drawnStar{passes: []float64{nearEnough, nearEnough,
		parsecsInKpc, parsecsInKpc, parsecsInKpc}}
	nearer := drawnStar{passes: []float64{nearEnough, twoKpcInPc / sixStars,
		twoKpcInPc / sixStars}}
	if by.nearerFirst(often, nearer) >= equalOrder ||
		by.nearerFirst(nearer, often) <= equalOrder ||
		by.nearerFirst(often, often) != equalOrder {
		test.Errorf("the orders are %d, %d, %d", by.nearerFirst(often,
			nearer), by.nearerFirst(nearer, often), by.nearerFirst(often,
			often))
	}
}

// the test: numbers at each check's edge run: a quality floor of none,
// shares of none and of the whole, speeds and flights each of one
// value, and the longest flight the whole span.
func TestNumbersAtTheirEdgesRun(test *testing.T) {
	base := testNumbers(test)
	floor, none, all, speeds, flights, span := base, base, base, base,
		base, base
	floor.quality, none.share, all.share = nothing, nothing, whole
	speeds.most = speeds.least
	flights.flightMost = flights.flightLeast
	span.flightMost = span.years / inAMillion
	for name, by := range map[string]picked{"floor": floor, "none": none,
		"all": all, "speeds": speeds, "flights": flights, "span": span} {
		if err := by.check(); err != nil {
			test.Errorf("%s: %v", name, err)
		}
	}
}

// the test and the anti-test: a pass at the radius is not within it; a
// candidate exactly at each cut is chosen; a star of no parallax and no
// error is not kept; a star named at the batch's end is refused; an id
// past what a float holds is read whole.
func TestTheEdgesFallTheirWay(test *testing.T) {
	by := testNumbers(test)
	if share := (drawnStar{passes: []float64{by.nearPc}}).within(by.nearPc,
		by.draws); share != nothing {
		test.Errorf("a pass at the radius is within it, %g", share)
	}
	line := drawsLine(chosenID, shareCut, speedCut, rvCut)
	table, err := readFrom(drawnAt(test, line), candidateColumns...)
	if err != nil {
		test.Fatal(err)
	}
	if chosen, err := choose(*table, by); err != nil ||
		len(chosen) != oneLine {
		test.Errorf("a candidate at the cuts gave %v, %v", chosen, err)
	}
	none := centreRow(centreID)
	none.Distant(nothing, nothing)
	if kept(sourcesOf(test, none)[firstLine], by.quality) {
		test.Error("a star of no parallax and no error was kept")
	}
	places := []int{firstRow, secondRow, thirdRow, halfBatch - oneLine}
	_, _, err = without(fourOut(), places, []int{halfBatch},
		make([]outcome, halfBatch))
	if err == nil || !strings.Contains(err.Error(), "named star 4 of a "+
		"batch of 4") {
		test.Errorf("a star at the batch's end gave %v", err)
	}
	path := tableOf(test, strings.Join(screenColumns, "\t")+"\n0\t"+
		strconv.FormatInt(longID, decimal)+"\t1\t1\n")
	stars, err := screenedFrom(path)
	if err != nil || stars[firstLine].id != longID {
		test.Errorf("the long id read back as %v, %v", stars, err)
	}
}

// the test: a candidate's folder and its readme are made with the modes
// a folder and a file for others to read have.
func TestACandidatesFolderIsMadeForOthersToRead(test *testing.T) {
	into := test.TempDir()
	each := candidate{id: "101", numbers: map[string]float64{}}
	if err := each.write(into, testNumbers(test)); err != nil {
		test.Fatal(err)
	}
	folder, err := os.Stat(filepath.Join(into, "101"))
	if err != nil {
		test.Fatal(err)
	}
	page, err := os.Stat(filepath.Join(into, "101", readme))
	if err != nil {
		test.Fatal(err)
	}
	if folder.Mode().Perm() != folderMode || page.Mode().Perm() != fileMode {
		test.Errorf("the folder is %v and the readme %v", folder.Mode(),
			page.Mode())
	}
}

// the anti-test: a star moving away along the rotation at all but the
// speed of light, less the model's Sun, would outrun light with the
// published Sun, 20 km/s faster: the frame refuses it under one
// estimate only, and the screen counts it refused, not placed.
func TestAStarOneEstimateCannotPlaceIsRefused(test *testing.T) {
	along, err := sky.NewGalactic(measure.Degrees(rotationLongitude),
		measure.Degrees(nothing))
	if err != nil {
		test.Fatal(err)
	}
	place := along.Equatorial()
	_, model := estimatesOf(test)
	_, moving := model.Sun()
	light := si.Speed / si.Kilo(length.Metre)
	edge := centreRow(centreID)
	edge.At(place.Ascension().InDegrees(), place.Declination().InDegrees())
	edge.Receding(light-moving.AlongY().InKilometresPerSecond()-underLight,
		radialError)
	into := intoA(test, "screen.tsv")
	code, _, errs := firangRan(running(test, &farFromCentre{}), published{},
		"screen", "--numbers", numbersPath, "--gaia", productAt(test, edge),
		"--into", into)
	notes := strings.Join(writtenAt(test, into).notes, " ")
	if code != exitRan || !strings.Contains(notes, "1 refused by the frame") {
		test.Errorf("exit %d, %q: %s", code, notes, errs)
	}
}
