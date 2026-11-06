package main

import (
	"math"
	"math/rand/v2"
	"strings"
	"testing"

	gpuorbit "github.com/janearc/celestial/gpu/orbit"
	"github.com/janearc/celestial/measure"
	"github.com/janearc/celestial/orbit"
	"go.uber.org/mock/gomock"
)

// The validation tests' own numbers, each defined.
const (
	seedOne     = 1    // defined: a seed for a test's draws
	stepRoom    = 1e-9 // defined: a share, a flight on its step to rounding
	placeRoom   = 1e-9 // defined: kpc, a place to rounding
	speedRoom   = 1e-9 // defined: a share, a speed to rounding
	pastC       = 4e5  // defined: km/s, a motion past light
	fortyThrown = 40   // defined: the tests' thrown stars
	passRoom    = 0.01 // defined: pc, a thrown star's true pass, a parsec out
)

// seeded is a source of draws from a fixed seed.
func seeded() *rand.Rand {
	return rand.New(rand.NewPCG(seedOne, seedOne))
}

// testNumbers is the tests' numbers, or the test stops.
func testNumbers(test *testing.T) picked {
	test.Helper()
	by, err := pickedFrom(numbersPath)
	if err != nil {
		test.Fatal(err)
	}
	return *by
}

// the test: each thrown star starts a parsec off Sgr A*, moving straight
// in at its speed, its motion reversed for the carry; its speed and its
// flight are in their ranges, its flight on a step.
func TestThrownStarsStartAParsecOffSgrAStar(test *testing.T) {
	by := testNumbers(test)
	stars := throw(by, seeded())
	if len(stars) != fortyThrown {
		test.Fatalf("threw %d stars", len(stars))
	}
	from := by.fromPc / parsecsInKpc
	for _, star := range stars {
		x, y, z := star.start.Numbers().At()
		vx, vy, vz := star.start.Numbers().Moving()
		away := math.Sqrt(x*x + y*y + z*z)
		speed := math.Sqrt(vx*vx + vy*vy + vz*vz)
		inward := (x*vx + y*vy + z*vz) / (away * speed)
		steps := star.flight / by.flightStep
		if !(math.Abs(away-from) <= placeRoom) ||
			!(math.Abs(speed-star.speed) <= speedRoom*speed) ||
			!(math.Abs(inward+whole) <= speedRoom) {
			test.Errorf("a star starts %g kpc out at %g km/s, %g inward",
				away, speed, inward)
		}
		if star.speed < by.least || star.speed > by.most ||
			star.flight < by.flightLeast || star.flight > by.flightMost ||
			!(math.Abs(steps-math.Round(steps)) <= stepRoom*steps) {
			test.Errorf("a star thrown at %g km/s flies %g Myr", star.speed,
				star.flight)
		}
	}
}

// carrying is a solver that answers as farFromCentre does and keeps the
// span and steps of each call, by the span's years.
type carrying struct {
	answers farFromCentre
	steps   map[float64]int
}

// back keeps the call's span and steps, then answers.
func (solver *carrying) back(stars []orbit.State, span measure.Duration,
	steps int) ([]orbit.Pass, []orbit.State, error) {
	solver.steps[span.InYears()] = steps
	return solver.answers.back(stars, span, steps)
}

// the test: the carry runs the stars of one flight together, for the
// flight's years in its share of the span's steps, and a star now is
// its start with its motion turned forward: straight out from Sgr A*.
func TestTheCarryRunsEachFlightTogether(test *testing.T) {
	by := testNumbers(test)
	stars := throw(by, seeded())
	controller := gomock.NewController(test)
	made := NewMocksolver(controller)
	solver := &carrying{steps: map[float64]int{}}
	made.EXPECT().Back(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(solver.back).AnyTimes()
	at := setting{made: made, batch: fullBatch}
	nows, gone, err := at.carry(stars, by)
	if err != nil {
		test.Fatal(err)
	}
	flights := map[float64]bool{}
	for which, star := range stars {
		flights[star.flight*inAMillion] = true
		x, y, z := nows[which].Numbers().At()
		vx, vy, vz := nows[which].Numbers().Moving()
		sx, sy, sz := star.start.Numbers().At()
		svx, svy, svz := star.start.Numbers().Moving()
		if gone[which] || x != sx || y != sy || z != sz || vx != -svx ||
			vy != -svy || vz != -svz {
			test.Errorf("star %d is now %v, from %v", which, nows[which],
				star.start)
		}
	}
	if len(solver.steps) != len(flights) {
		test.Errorf("%d calls for %d flights", len(solver.steps),
			len(flights))
	}
	for years, steps := range solver.steps {
		share := float64(by.steps) * years / by.years
		want := int(math.Round(share))
		if steps != want {
			test.Errorf("a flight of %g years ran %d steps, want %d", years,
				steps, want)
		}
	}
}

// the anti-test: a thrown star the solver names as run away in the
// carry is gone; and a solver's other refusal stops the carry.
func TestARunawayInTheCarryIsGone(test *testing.T) {
	by := testNumbers(test)
	stars := throw(by, seeded())[:oneLine]
	controller := gomock.NewController(test)
	made := NewMocksolver(controller)
	made.EXPECT().Back(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil, gpuorbit.NewRunaway([]int{firstOfDraw}))
	at := setting{made: made, batch: fullBatch}
	_, gone, err := at.carry(stars, by)
	if err != nil || !gone[firstLine] {
		test.Errorf("a runaway in the carry gave %v, %v", gone, err)
	}
	refuser := NewMocksolver(controller)
	refuser.EXPECT().Back(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil, errDevice)
	stopped := setting{made: refuser, batch: fullBatch}
	if nows, _, err := stopped.carry(stars, by); err == nil || nows != nil {
		test.Errorf("a refused carry gave %v, %v", nows, err)
	}
}

// the test: the validation throws the stars, carries them out a parsec
// from Sgr A* as the mocked solver leaves them, sees them with the errors
// of the star at the centre, small enough that all are kept, and draws
// them: each passes within near-pc in every draw, and its true pass is a
// parsec.
func TestTheValidationSeesAndDrawsTheThrownStars(test *testing.T) {
	product := productAt(test, centreRow(centreID))
	into := intoA(test, "validate.tsv")
	code, _, errs := firangRan(running(test, &farFromCentre{}),
		published{}, "validate", "--numbers", numbersPath, "--gaia",
		product, "--screened", screenedAt(test, screened{row: firstRow,
			id: centreID}), "--into", into)
	if code != exitRan {
		test.Fatalf("exit %d: %s", code, errs)
	}
	table := writtenAt(test, into)
	if !strings.Contains(strings.Join(table.notes, " "), "40 thrown, 40 "+
		"kept by the floor, 0 lost") || len(table.lines) != fortyThrown {
		test.Fatalf("the validation wrote %q and %d lines", table.notes,
			len(table.lines))
	}
	for _, line := range table.lines {
		pass := table.number(test, line, "true_pass_pc")
		if table.number(test, line, "share_near") != whole ||
			table.number(test, line, "lender_id") != centreID ||
			!(math.Abs(pass-whole) <= passRoom) {
			test.Errorf("a thrown star wrote %v", line)
		}
	}
}

// the anti-test: a thrown star the validation cannot see is lost: one
// moving past light, one at the Sun's own place, one whose lender's
// errors cannot be drawn from, and one seen under the floor.
func TestWhatTheValidationCannotSeeIsLost(test *testing.T) {
	by := testNumbers(test)
	knotted := centreRow(centreID)
	knotted.Tied(tied, tied, -tied)
	poor := centreRow(poorID)
	poor.Distant(centreParallax(), poorError)
	product := productAt(test, centreRow(centreID), knotted, poor)
	catalogue, file, err := opened(product)
	if err != nil {
		test.Fatal(err)
	}
	defer file.Close()
	frame, err := published{}.Frame()
	if err != nil {
		test.Fatal(err)
	}
	at := setting{frame: frame}
	sun, _ := frame.Sun()
	near := orbit.StateOf(by.fromPc/parsecsInKpc, nothing, nothing,
		by.least, nothing, nothing)
	fast := orbit.StateOf(by.fromPc/parsecsInKpc, nothing, nothing, pastC,
		nothing, nothing)
	still, err := measure.KilometresPerSecond(nothing, nothing, nothing)
	if err != nil {
		test.Fatal(err)
	}
	cases := map[string]struct {
		now    orbit.State
		lender screened
	}{
		"past light":      {fast, screened{row: firstRow, id: centreID}},
		"at the sun":      {orbit.NewState(sun, *still), screened{row: firstRow}},
		"knotted":         {near, screened{row: secondRow, id: centreID}},
		"under the floor": {near, screened{row: thirdRow, id: poorID}},
	}
	for name, each := range cases {
		seen, truth, err := at.see(catalogue, []screened{each.lender},
			thrown{}, each.now, by, seeded())
		if err != nil || seen != nil || truth != nil {
			test.Errorf("%s gave %v, %v, %v", name, seen, truth, err)
		}
	}
}
