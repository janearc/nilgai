package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"

	"github.com/janearc/celestial/measure"
	"github.com/janearc/celestial/orbit"
	"github.com/janearc/puzzlemouth/pzm"
)

// The validation's own numbers, each defined.
const (
	turn         = 2 * math.Pi // defined: a circle is two pi radians round
	halves       = 2           // defined: from one pole to the other is two
	parsecsInKpc = 1000        // defined: parsecs in a kiloparsec
	lowestPole   = -1          // defined: the south pole's height on a sphere
	fewestSteps  = 1           // defined: a flight is one step at the least
)

// thrown is a star thrown from Sgr A*: how fast, in kilometres per
// second a parsec out, how long ago in millions of years, and the state
// it is carried from: a parsec out, its motion reversed.
type thrown struct {
	speed, flight float64
	start         orbit.State
}

// validated is a thrown star as Gaia would see it, drawn: the thrown
// star, whose errors it borrowed, its true pass in parsecs, where it is
// and how fast now, how far from the Sun and how well its parallax is
// measured, and its draws.
type validated struct {
	thrown
	lender                   int64
	truePass, away, speedNow float64
	distance, quality        float64
	drawn                    drawnStar
}

// validate is the verb: stars thrown from Sgr A*, carried forward, seen
// with the errors of stars the screen found, and drawn as those were,
// a line a star kept by the quality floor to --into.
func (runs runner) validate(call pzm.Call) error {
	by, err := runs.asked(call, "gaia", "screened", "into")
	if err != nil {
		return err
	}
	if call.DryRun() {
		fmt.Fprintf(call.Out(), "would throw %d stars from sgr a* at %g to "+
			"%g km/s, %g to %g Myr ago, see them with the errors of stars "+
			"in %s from %s, draw each %d times, and write them to %s\n",
			by.thrown, by.least, by.most, by.flightLeast, by.flightMost,
			call.Value("screened"), call.Value("gaia"), by.draws,
			call.Value("into"))
		return nil
	}
	lenders, err := screenedFrom(call.Value("screened"))
	if err != nil {
		return err
	}
	if len(lenders) == nothing {
		return fmt.Errorf("%s has no stars to lend their errors",
			call.Value("screened"))
	}
	catalogue, file, err := opened(call.Value("gaia"))
	if err != nil {
		return err
	}
	defer file.Close()
	at, err := settingFor(runs.read, runs.open, *by)
	if err != nil {
		return err
	}
	defer at.made.Close()
	random := rand.New(rand.NewPCG(by.validationSeed, by.validationSeed))
	found, lost, err := at.validateAll(catalogue, lenders, *by, random)
	if err != nil {
		return err
	}
	return writeTo(call.Value("into"), validatedTable(found, lost, *by))
}

// validateAll throws the stars, carries them, sees them and draws those
// the floor keeps; it gives them and how many were lost on the way.
func (at setting) validateAll(catalogue product, lenders []screened,
	by picked, random drawing) ([]validated, int, error) {
	stars := throw(by, random)
	nows, gone, err := at.carry(stars, by)
	if err != nil {
		return nil, nothing, err
	}
	var found []validated
	var truths []orbit.State
	lost := nothing
	for which, star := range stars {
		if gone[which] {
			lost++
			continue
		}
		seen, truth, err := at.see(catalogue, lenders, star, nows[which],
			by, random)
		if err != nil {
			return nil, nothing, err
		}
		if seen == nil {
			lost++
			continue
		}
		found, truths = append(found, *seen), append(truths, *truth)
	}
	passes, err := at.back(truths)
	if err != nil {
		return nil, nothing, err
	}
	for which := range found {
		found[which].truePass = passes[which]
	}
	if err := at.drawValidated(found, by, random); err != nil {
		return nil, nothing, err
	}
	return found, lost, nil
}

// throw is the stars thrown: each in a direction drawn evenly over the
// sphere, at a speed and a flight drawn evenly in their ranges, the
// flight rounded to the step so stars of one flight are carried
// together.
func throw(by picked, random drawing) []thrown {
	stars := make([]thrown, by.thrown)
	from := by.fromPc / parsecsInKpc
	for which := range stars {
		up := lowestPole + halves*random.Float64()
		round := turn * random.Float64()
		across := math.Sqrt(whole - up*up)
		x, y, z := across*math.Cos(round), across*math.Sin(round), up
		speed := by.least + (by.most-by.least)*random.Float64()
		raw := by.flightLeast + (by.flightMost-by.flightLeast)*
			random.Float64()
		flight := math.Round(raw/by.flightStep) * by.flightStep
		start := orbit.StateOf(from*x, from*y, from*z, -speed*x, -speed*y,
			-speed*z)
		stars[which] = thrown{speed: speed, flight: flight, start: start}
	}
	return stars
}

// carry is each thrown star carried forward for its flight: a back-run
// of its reversed motion, whose end, its motion reversed again, is the
// star now. Stars of one flight go together; a star that ran away is
// marked gone.
func (at setting) carry(stars []thrown, by picked) ([]orbit.State,
	[]bool, error) {
	flights := map[float64][]int{}
	for which, star := range stars {
		flights[star.flight] = append(flights[star.flight], which)
	}
	nows := make([]orbit.State, len(stars))
	gone := make([]bool, len(stars))
	for flight, members := range flights {
		years := flight * inAMillion
		share := float64(by.steps) * years / by.years
		rounded := int(math.Round(share))
		steps := max(fewestSteps, rounded)
		var starts []orbit.State
		for _, which := range members {
			starts = append(starts, stars[which].start)
		}
		outcomes, err := at.run(starts, measure.Years(years), steps)
		if err != nil {
			return nil, nil, err
		}
		for place, which := range members {
			gone[which] = outcomes[place].ranAway
			nows[which] = reversed(outcomes[place].end)
		}
	}
	return nows, gone, nil
}

// reversed is a state with its motion turned round.
func reversed(state orbit.State) orbit.State {
	x, y, z := state.Numbers().At()
	vx, vy, vz := state.Numbers().Moving()
	return orbit.StateOf(x, y, z, -vx, -vy, -vz)
}

// see is a thrown star as Gaia would see it, with the errors of a star
// drawn from the lenders, and its true state; nil when its errors cannot
// be drawn from, or its seen parallax is under the floor.
func (at setting) see(catalogue product, lenders []screened, star thrown,
	now orbit.State, by picked, random drawing) (*validated, *orbit.State,
	error) {
	moving, err := now.Motion()
	if err != nil {
		return nil, nil, nil
	}
	sighting, err := at.frame.Seen(now.Place(), *moving)
	if err != nil {
		return nil, nil, nil
	}
	lender := lenders[random.IntN(len(lenders))]
	source, err := catalogue.Source(lender.row)
	if err != nil {
		return nil, nil, err
	}
	truth := source.As(*sighting)
	spread, err := truth.Spread()
	if err != nil {
		return nil, nil, nil
	}
	observed, err := spread.Drawn(random)
	if err != nil || !kept(observed, by.quality) {
		return nil, nil, nil
	}
	x, y, z := now.Numbers().At()
	vx, vy, vz := now.Numbers().Moving()
	mas := masOf(observed.Parallax())
	return &validated{thrown: star, lender: lender.id,
		away:     math.Sqrt(x*x + y*y + z*z),
		speedNow: math.Sqrt(vx*vx + vy*vy + vz*vz), distance: whole / mas,
		quality: mas / masOf(observed.ParallaxError()),
		drawn:   drawnStar{source: observed}}, &now, nil
}

// drawValidated draws each seen star as a real one is drawn, a batch
// at a time.
func (at setting) drawValidated(found []validated, by picked,
	random drawing) error {
	perBatch := max(oneStar, at.batch/by.draws)
	for first := firstStar; first < len(found); first += perBatch {
		last := min(first+perBatch, len(found))
		var stars []drawnStar
		for _, each := range found[first:last] {
			stars = append(stars, each.drawn)
		}
		drawn, _, err := at.drawRun(stars, by, random)
		if err != nil {
			return err
		}
		for which, star := range drawn {
			found[first+which].drawn = star
		}
	}
	return nil
}

// validateColumns are the validation table's columns.
var validateColumns = []string{"thrown_kms", "flight_myr",
	"true_pass_pc", "r_gc_kpc", "v_gc_kms", "distance_kpc",
	"parallax_over_error", "share_near", "share_screen", "p16_pc",
	"p50_pc", "p84_pc", "passed", "unplaced", "ran_away", "lender_id"}

// validatedTable is the validation as a table: a header that says what
// was thrown, kept and lost, the true passes' middle, and how often the
// stars' draws came near, then one line a star.
func validatedTable(found []validated, lost int, by picked) table {
	var trues, shares []float64
	for _, each := range found {
		trues = append(trues, each.truePass)
		shares = append(shares, each.drawn.within(by.nearPc, by.draws))
	}
	slices.Sort(trues)
	slices.Sort(shares)
	middle := drawnStar{passes: trues}
	near := drawnStar{passes: shares}
	notes := []string{fmt.Sprintf("firang validate: %d thrown, %d kept by "+
		"the floor, %d lost to the carry, the errors or the floor",
		by.thrown, len(found), lost), fmt.Sprintf("true passes: median %g "+
		"pc; shares within %g pc: %g, %g, %g at the 16th, 50th and 84th "+
		"percentiles", middle.quantile(midShare), by.nearPc,
		near.quantile(lowShare), near.quantile(midShare),
		near.quantile(highShare))}
	var lines [][]string
	for _, each := range found {
		lines = append(lines, each.line(by))
	}
	return table{notes: notes, columns: validateColumns, lines: lines}
}

// line is a validated star's line, in the validation table's columns.
func (each validated) line(by picked) []string {
	drawn := each.drawn
	numbers := []float64{each.speed, each.flight, each.truePass,
		each.away, each.speedNow, each.distance, each.quality,
		drawn.within(by.nearPc, by.draws), drawn.within(by.screenPc, by.draws),
		drawn.quantile(lowShare), drawn.quantile(midShare),
		drawn.quantile(highShare)}
	var fields []string
	for _, number := range numbers {
		fields = append(fields, fmt.Sprint(number))
	}
	counts := []int{len(drawn.passes), drawn.unplaced, drawn.ranAway}
	for _, count := range counts {
		fields = append(fields, fmt.Sprint(count))
	}
	return append(fields, fmt.Sprint(each.lender))
}
