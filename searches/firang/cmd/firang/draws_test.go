package main

import (
	"fmt"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/janearc/celestial-data/gaia"
	gpuorbit "github.com/janearc/celestial/gpu/orbit"
	"github.com/janearc/celestial/measure"
	"github.com/janearc/celestial/orbit"
	"go.uber.org/mock/gomock"
)

// The draws tests' own numbers, each defined.
const (
	tenDraws    = 10   // defined: the tests' draws a star
	tied        = 0.9  // defined: a correlation near one
	wideError   = 5.0  // defined: mas, ten times the parallax it is on
	halfMas     = 0.5  // defined: mas, two kiloparsecs off
	kpcRoom     = 1e-9 // defined: a share, the far star's distance to rounding
	secondLine  = 1    // defined: a table's second star
	twoLines    = 2    // defined: two stars drawn
	firstOfDraw = 0    // defined: the first draw a batch has
)

// screenedAt writes a screen's table of the stars given, by row, id and
// passes, and gives its path.
func screenedAt(test *testing.T, stars ...screened) string {
	test.Helper()
	lines := []string{"# a screen's table, for a test",
		strings.Join(screenColumns, "\t")}
	for _, star := range stars {
		lines = append(lines, fmt.Sprintf("%d\t%d\t%g\t%g", star.row, star.id,
			star.nearFrame, star.nearModel))
	}
	path := intoA(test, "screen.tsv")
	text := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(text), fileBit); err != nil {
		test.Fatal(err)
	}
	return path
}

// number is a line's field under a column as a number, or the test
// stops.
func (table written) number(test *testing.T, line []string,
	column string) float64 {
	test.Helper()
	value, err := strconv.ParseFloat(table.field(test, line, column), 64)
	if err != nil {
		test.Fatal(err)
	}
	return value
}

// drawsRan runs the draws over the product and screened stars given,
// with the solver's answers given, and gives what it wrote.
func drawsRan(test *testing.T, open opener, product string,
	stars ...screened) written {
	test.Helper()
	into := intoA(test, "draws.tsv")
	code, _, errs := firangRan(open, published{}, "draws", "--numbers",
		numbersPath, "--gaia", product, "--screened", screenedAt(test,
			stars...), "--into", into)
	if code != exitRan {
		test.Fatalf("exit %d: %s", code, errs)
	}
	return writtenAt(test, into)
}

// the test: the star at the centre passes near in every draw and comes
// first; the star far from it never does; each has every draw placed and
// run, and the far star is as far from the centre as the frame places
// it.
func TestTheDrawsCountHowOftenAStarPassesNear(test *testing.T) {
	product := productAt(test, centreRow(centreID), farRow(farID))
	table := drawsRan(test, running(test, &farFromCentre{}), product,
		screened{row: secondRow, id: farID}, screened{row: firstRow, id: centreID})
	if !slices.Equal(table.columns, drawsColumns) ||
		len(table.lines) != twoLines {
		test.Fatalf("the draws wrote %v %v", table.columns, table.lines)
	}
	centre, far := table.lines[firstLine], table.lines[secondLine]
	wants := map[string][]float64{
		"source_id":    {centreID, farID},
		"share_near":   {whole, nothing},
		"share_screen": {whole, nothing},
		"passed":       {tenDraws, tenDraws},
		"unplaced":     {nothing, nothing},
		"ran_away":     {nothing, nothing},
	}
	for column, want := range wants {
		got := []float64{table.number(test, centre, column),
			table.number(test, far, column)}
		if !slices.Equal(got, want) {
			test.Errorf("%s is %v, want %v", column, got, want)
		}
	}
	frame, err := published{}.Frame()
	if err != nil {
		test.Fatal(err)
	}
	farProduct := productAt(test, farRow(farID))
	source := firstSource(test, farProduct)
	state, err := source.State(*frame)
	if err != nil {
		test.Fatal(err)
	}
	x, y, z := state.Numbers().At()
	away := math.Sqrt(x*x + y*y + z*z)
	if got := table.number(test, far, "r_gc_kpc"); !(math.Abs(got-away) <=
		kpcRoom*away) {
		test.Errorf("the far star is %g kpc from the centre, want %g", got,
			away)
	}
}

// the anti-test: a star whose errors no draw can come from, three ties
// one the wrong way, is left out and named in the notes; a star whose
// parallax is a tenth of its error draws to no distance often, and every
// draw is either run or counted unplaced.
func TestAStarThatCannotBeDrawnIsLeftOutAndOneUnplacedIsCounted(
	test *testing.T) {
	knotted := centreRow(centreID)
	knotted.Tied(tied, tied, -tied)
	loose := farRow(farID)
	loose.Distant(halfMas, wideError)
	product := productAt(test, knotted, loose)
	table := drawsRan(test, running(test, &farFromCentre{}), product,
		screened{row: firstRow, id: centreID}, screened{row: secondRow, id: farID})
	notes := strings.Join(table.notes, "\n")
	if len(table.lines) != oneLine || !strings.Contains(notes,
		"star 1 is left out: ") || !strings.Contains(notes,
		"not positive definite") {
		test.Fatalf("the draws wrote %q, %v", notes, table.lines)
	}
	line := table.lines[firstLine]
	passed := table.number(test, line, "passed")
	unplaced := table.number(test, line, "unplaced")
	if !(unplaced > nothing) || passed+unplaced != tenDraws {
		test.Errorf("%g passed and %g unplaced of %d", passed, unplaced,
			tenDraws)
	}
}

// runningAway is a solver that names each batch's first draw as run
// away, once a batch, and answers the rest as farFromCentre does.
type runningAway struct {
	answers farFromCentre
	named   bool
}

// back names the first draw, then answers the rest.
func (solver *runningAway) back(stars []orbit.State, span measure.Duration,
	steps int) ([]orbit.Pass, []orbit.State, error) {
	if !solver.named {
		solver.named = true
		return nil, nil, gpuorbit.NewRunaway([]int{firstOfDraw})
	}
	solver.named = false
	return solver.answers.back(stars, span, steps)
}

// the anti-test: a draw that ran away is counted so, and not as a pass:
// ten draws are two of the solver's batches, and each names one.
func TestADrawThatRanAwayIsCounted(test *testing.T) {
	controller := gomock.NewController(test)
	open, made := NewMockopener(controller), NewMocksolver(controller)
	open.EXPECT().Open(gomock.Any(), gomock.Any()).Return(made, nil)
	made.EXPECT().Close()
	solver := &runningAway{}
	made.EXPECT().Back(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(solver.back).AnyTimes()
	centre := productAt(test, centreRow(centreID))
	table := drawsRan(test, open, centre,
		screened{row: firstRow, id: centreID})
	line := table.lines[firstLine]
	if table.number(test, line, "ran_away") != restOfThem ||
		table.number(test, line, "passed") != tenDraws-restOfThem {
		test.Errorf("the draws counted %v", line)
	}
}

// the test: drawn stars order by their share near, the most first, then
// by their middle pass, the nearest first; as near as often and as near
// in the middle, they keep their order.
func TestDrawnStarsOrderNearestFirst(test *testing.T) {
	by, err := pickedFrom(numbersPath)
	if err != nil {
		test.Fatal(err)
	}
	often := drawnStar{passes: []float64{nearEnough, nearEnough}}
	seldom := drawnStar{passes: []float64{nearEnough, parsecsInKpc,
		parsecsInKpc}}
	nearer := drawnStar{passes: []float64{nearEnough, parsecsInKpc / halves,
		parsecsInKpc}}
	orders := []int{by.nearerFirst(often, seldom),
		by.nearerFirst(seldom, often), by.nearerFirst(seldom, nearer),
		by.nearerFirst(nearer, seldom), by.nearerFirst(often, often)}
	if !slices.Equal(orders, []int{before, after, after, before, same}) {
		test.Errorf("the orders are %v", orders)
	}
	if none := (drawnStar{}).quantile(midShare); !math.IsNaN(none) {
		test.Errorf("a star with no passes has a middle of %g", none)
	}
}

// firstSource is the first source of the product at the path given, or
// the test stops.
func firstSource(test *testing.T, path string) *gaia.Source {
	test.Helper()
	catalogue, file, err := opened(path)
	if err != nil {
		test.Fatal(err)
	}
	defer file.Close()
	source, err := catalogue.Source(firstLine)
	if err != nil {
		test.Fatal(err)
	}
	return source
}

// the anti-test: a screen's table the draws cannot read stops them with
// its words: not there, without a column, with a pass or a row that is
// not a number.
func TestAScreenTableTheDrawsCannotReadStopsThem(test *testing.T) {
	product := productAt(test, centreRow(centreID))
	tables := map[string]string{
		"no such file": intoA(test, "none.tsv"),
		"has no column near_model_pc": tableOf(test, "row\tsource_id\t"+
			"near_frame_pc\n0\t1\t2\n"),
		"invalid syntax": tableOf(test, strings.Join(screenColumns, "\t")+
			"\n0\t1\tfar\t2\n"),
		"parsing \"first\"": tableOf(test, strings.Join(screenColumns, "\t")+
			"\nfirst\t1\t2\t2\n"),
	}
	for said, path := range tables {
		open := NewMockopener(gomock.NewController(test))
		code, _, errs := firangRan(open, published{}, "draws", "--numbers",
			numbersPath, "--gaia", product, "--screened", path, "--into",
			intoA(test, "draws.tsv"))
		if code == exitRan || !strings.Contains(errs, said) {
			test.Errorf("%s: exit %d, %s", said, code, errs)
		}
	}
}

// tableOf writes a table's text to a file and gives its path.
func tableOf(test *testing.T, text string) string {
	test.Helper()
	path := intoA(test, "table.tsv")
	if err := os.WriteFile(path, []byte(text), fileBit); err != nil {
		test.Fatal(err)
	}
	return path
}
