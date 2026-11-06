package main

import (
	"errors"
	"math"
	"slices"
	"strconv"
	"strings"
	"testing"

	gpuorbit "github.com/janearc/celestial/gpu/orbit"
	"github.com/janearc/celestial/measure"
	"github.com/janearc/celestial/orbit"
	"github.com/janearc/nilgai/internal/fitstest"
	"go.uber.org/mock/gomock"
)

// The screen tests' own numbers, each defined.
const (
	fiveCentres  = 5   // defined: stars over one batch and into a second
	fullBatch    = 8   // defined: the tests' batch, four stars twice
	restOfThem   = 2   // defined: the fifth star twice
	nearEnough   = 1.0 // defined: pc, the centre star's pass under either sun
	oneLine      = 1   // defined: one star found
	refusedAt    = 1   // defined: the star the solver names as run away
	fourStars    = 4   // defined: the stars in the runaway batch
	pastTheBatch = 9   // defined: a place no batch of four has
)

// the test: the screen keeps the star at the centre and leaves the one
// far from it; it counts the rows, those kept, those the frame refused,
// and those found, and writes each found star's row, id and passes.
func TestTheScreenKeepsStarsThatPassNear(test *testing.T) {
	answers := &farFromCentre{}
	fast := centreRow(fastID)
	fast.Receding(pastLight, radialError)
	poor := farRow(poorID)
	poor.Distant(poorMas, poorError)
	product := productAt(test, centreRow(centreID), farRow(farID), poor,
		fast)
	into := intoA(test, "screen.tsv")
	code, _, errs := firangRan(running(test, answers), published{}, "screen",
		"--numbers", numbersPath, "--gaia", product, "--into", into)
	if code != exitRan {
		test.Fatalf("exit %d: %s", code, errs)
	}
	table := writtenAt(test, into)
	if len(table.notes) != oneLine || !strings.Contains(table.notes[firstLine],
		"4 rows, 3 kept at a parallax of 5 times its error, 1 refused by the "+
			"frame, 1 within 1000 pc") {
		test.Errorf("the screen's notes are %q", table.notes)
	}
	if !slices.Equal(table.columns, screenColumns) ||
		len(table.lines) != oneLine {
		test.Fatalf("the screen wrote %v %v", table.columns, table.lines)
	}
	line := table.lines[firstLine]
	if table.field(test, line, "row") != "0" ||
		table.field(test, line, "source_id") != "1" {
		test.Errorf("the screen found %v", line)
	}
	for _, column := range []string{"near_frame_pc", "near_model_pc"} {
		near, err := strconv.ParseFloat(table.field(test, line, column), 64)
		if err != nil || !(near < nearEnough) {
			test.Errorf("the centre star passes %s at %g pc, %v", column, near,
				err)
		}
	}
}

// the test: a screen over more stars than a batch holds runs them in
// batches as full as the setting's: four stars, each twice, then the
// fifth twice.
func TestTheScreenRunsInFullBatches(test *testing.T) {
	answers := &farFromCentre{}
	var rows []*fitstest.GaiaStar
	for each := range fiveCentres {
		id := int64(each + centreID)
		rows = append(rows, centreRow(id))
	}
	into := intoA(test, "screen.tsv")
	code, _, errs := firangRan(running(test, answers), published{}, "screen",
		"--numbers", numbersPath, "--gaia", productAt(test, rows...),
		"--into", into)
	if code != exitRan || !slices.Equal(answers.batches, []int{fullBatch,
		restOfThem}) {
		test.Fatalf("exit %d, batches %v: %s", code, answers.batches, errs)
	}
	if found := writtenAt(test, into); len(found.lines) != fiveCentres {
		test.Errorf("the screen found %d of %d", len(found.lines),
			fiveCentres)
	}
}

// refusingOne is a solver that names one star of its first batch as run
// away and answers the rest as farFromCentre does.
type refusingOne struct {
	answers farFromCentre
	refused bool
}

// back refuses the first batch once, then answers.
func (solver *refusingOne) back(stars []orbit.State, span measure.Duration,
	steps int) ([]orbit.Pass, []orbit.State, error) {
	if !solver.refused {
		solver.refused = true
		return nil, nil, gpuorbit.NewRunaway([]int{refusedAt})
	}
	return solver.answers.back(stars, span, steps)
}

// fourOut are four stars at rest on the x axis, one to four kiloparsecs
// from the centre.
func fourOut() []orbit.State {
	var stars []orbit.State
	for each := range fourStars {
		away := float64(each + oneLine)
		stars = append(stars, orbit.StateOf(away, nothing,
			nothing, nothing, nothing, nothing))
	}
	return stars
}

// the test and the anti-test: a star the solver names as run away is
// marked so, not a number among the passes, and the rest run again and
// keep their own passes; a star named past the batch is refused, and
// any other refusal of the solver stops the run with its words.
func TestARunawayIsMarkedAndTheRestRunAgain(test *testing.T) {
	controller := gomock.NewController(test)
	made := NewMocksolver(controller)
	refusing := &refusingOne{}
	made.EXPECT().Back(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(refusing.back).Times(restOfThem)
	at := setting{made: made, batch: fullBatch}
	near, err := at.back(fourOut())
	if err != nil {
		test.Fatal(err)
	}
	for which, pass := range near {
		want := float64(which+oneLine) * parsecsInKpc
		if which == refusedAt && !math.IsNaN(pass) ||
			which != refusedAt && !(math.Abs(pass-want) <= nearEnough) {
			test.Errorf("star %d passes at %g pc", which, pass)
		}
	}
	for _, each := range []struct {
		refusal error
		said    string
	}{{errors.New("the device is gone"), "the device is gone"},
		{gpuorbit.NewRunaway([]int{pastTheBatch}),
			"the solver named star 9 of a batch of 4"}} {
		refuser := NewMocksolver(controller)
		refuser.EXPECT().Back(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, nil, each.refusal)
		stopped := setting{made: refuser, batch: fullBatch}
		if near, err := stopped.back(fourOut()); err == nil || near != nil ||
			!strings.Contains(err.Error(), each.said) {
			test.Errorf("%v gave %v, %v", each.refusal, near, err)
		}
	}
}
