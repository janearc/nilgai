package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"

	"github.com/janearc/celestial-data/gaia"
	"github.com/janearc/celestial/orbit"
	"github.com/janearc/puzzlemouth/pzm"
)

// drawing is a source of random draws, as math/rand's Rand is: normal
// ones for a star's errors, uniform ones for the Sun's speed and for
// picking a star.
type drawing interface {
	NormFloat64() float64
	Float64() float64
	IntN(count int) int
}

// The draws' own numbers, each defined.
const (
	lowShare  = 0.16 // defined: one standard deviation below the middle
	midShare  = 0.5  // defined: the middle
	highShare = 0.84 // defined: one standard deviation above
	whole     = 1    // defined: the whole of a share
	oneStar   = 1    // defined: the fewest stars in a batch
)

// drawnStar is a star run back across its errors: the screen's star,
// the source, its passes in parsecs sorted, the draws that could not be
// placed and that ran away, and where it is and how it moves about the
// centre now, at its central values with the Sun as published.
type drawnStar struct {
	star                 screened
	source               *gaia.Source
	passes               []float64
	unplaced, ranAway    int
	away, speed, outward float64
}

// draws is the verb: each star the screen found run back draws times
// across its errors and the Sun's speed, a line a star to --into, the
// most often near first.
func (runs runner) draws(call pzm.Call) error {
	by, err := runs.asked(call, "gaia", "screened", "into")
	if err != nil {
		return err
	}
	if call.DryRun() {
		fmt.Fprintf(call.Out(), "would run each star in %s back %d times "+
			"across its errors and the sun's speed, from %s, and write how "+
			"often each passes within %g pc to %s\n",
			call.Value("screened"), by.draws, call.Value("gaia"), by.nearPc,
			call.Value("into"))
		return nil
	}
	stars, err := screenedFrom(call.Value("screened"))
	if err != nil {
		return err
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
	random := rand.New(rand.NewPCG(by.seed, by.seed))
	drawn, left, err := at.drawAll(catalogue, stars, *by, random)
	if err != nil {
		return err
	}
	return writeTo(call.Value("into"), drawnTable(drawn, left, *by))
}

// drawAll runs each star back across its errors, as many stars a batch
// as the batch holds their draws. A star whose errors cannot be drawn
// from is left out with a note.
func (at setting) drawAll(catalogue product, stars []screened,
	by picked, random drawing) ([]drawnStar, []string, error) {
	var drawn []drawnStar
	var left []string
	perBatch := max(oneStar, at.batch/by.draws)
	for first := firstStar; first < len(stars); first += perBatch {
		last := min(first+perBatch, len(stars))
		batch, notes, err := at.drawBatch(catalogue, stars[first:last], by,
			random)
		if err != nil {
			return nil, nil, err
		}
		drawn, left = append(drawn, batch...), append(left, notes...)
	}
	return drawn, left, nil
}

// drawBatch reads a batch of stars from the catalogue and draws them.
func (at setting) drawBatch(catalogue product, stars []screened,
	by picked, random drawing) ([]drawnStar, []string, error) {
	var read []drawnStar
	for _, star := range stars {
		source, err := catalogue.Source(star.row)
		if err != nil {
			return nil, nil, err
		}
		read = append(read, drawnStar{star: star, source: source})
	}
	return at.drawRun(read, by, random)
}

// drawRun draws each star across its errors and runs the draws back
// together. A star whose errors cannot be drawn from is left out with a
// note.
func (at setting) drawRun(stars []drawnStar, by picked,
	random drawing) ([]drawnStar, []string, error) {
	var drawn []drawnStar
	var left []string
	var states []orbit.State
	var owners []int
	for _, star := range stars {
		spread, err := star.source.Spread()
		if err != nil {
			left = append(left, fmt.Sprintf("star %d is left out: %v",
				star.star.id, err))
			continue
		}
		made, unplaced, err := at.drawn(spread, by.draws, random)
		if err != nil {
			return nil, nil, err
		}
		for range made {
			owners = append(owners, len(drawn))
		}
		states = append(states, made...)
		star.unplaced = unplaced
		drawn = append(drawn, at.central(star))
	}
	near, err := at.back(states)
	if err != nil {
		return nil, nil, err
	}
	for which, pass := range near {
		owner := &drawn[owners[which]]
		if math.IsNaN(pass) {
			owner.ranAway++
			continue
		}
		owner.passes = append(owner.passes, pass)
	}
	for which := range drawn {
		slices.Sort(drawn[which].passes)
	}
	return drawn, left, nil
}

// drawn is a spread's draws placed in the frame, each with the Sun's
// speed along the rotation drawn between the two estimates, and how
// many could not be placed: drawn past a pole, or to no distance, or
// refused by the frame.
func (at setting) drawn(spread *gaia.Spread, count int,
	random drawing) ([]orbit.State, int, error) {
	var states []orbit.State
	unplaced := nothing
	for range count {
		one, err := spread.Drawn(random)
		alongRotation := at.low + (at.high-at.low)*random.Float64()
		if err != nil {
			unplaced++
			continue
		}
		frame, err := at.movingAt(alongRotation)
		if err != nil {
			return nil, nothing, err
		}
		state, err := one.State(*frame)
		if err != nil {
			unplaced++
			continue
		}
		states = append(states, *state)
	}
	return states, unplaced, nil
}

// central is the star with where it is and how it moves about the
// centre now, at its central values with the Sun as published: its
// distance from the centre, its speed, and the part of it outward.
func (at setting) central(star drawnStar) drawnStar {
	state, err := star.source.State(*at.frame)
	if err != nil {
		star.away, star.speed, star.outward = math.NaN(), math.NaN(),
			math.NaN()
		return star
	}
	x, y, z := state.Numbers().At()
	vx, vy, vz := state.Numbers().Moving()
	star.away = math.Sqrt(x*x + y*y + z*z)
	star.speed = math.Sqrt(vx*vx + vy*vy + vz*vz)
	star.outward = (x*vx + y*vy + z*vz) / star.away
	return star
}

// within is the share of the draws asked for whose pass is under the
// radius.
func (star drawnStar) within(radius float64, draws int) float64 {
	count := nothing
	for _, pass := range star.passes {
		if pass < radius {
			count++
		}
	}
	return float64(count) / float64(draws)
}

// quantile is the pass a share of the way up the sorted passes, or not
// a number when none passed.
func (star drawnStar) quantile(share float64) float64 {
	if len(star.passes) == nothing {
		return math.NaN()
	}
	last := float64(len(star.passes) - oneStar)
	return star.passes[int(math.Round(share*last))]
}

// drawsColumns are the draws table's columns.
var drawsColumns = []string{"source_id", "near_frame_pc", "near_model_pc",
	"share_near", "share_screen", "p16_pc", "p50_pc", "p84_pc", "passed",
	"unplaced", "ran_away", "r_gc_kpc", "v_gc_kms", "v_out_kms",
	"parallax_over_error", "distance_kpc", "rv_kms", "rv_error_kms",
	"g_mag", "ruwe", "ra_deg", "dec_deg"}

// drawnTable is the draws as a table, the stars most often near first
// and, among those as often, the nearest middle pass first.
func drawnTable(drawn []drawnStar, left []string, by picked) table {
	slices.SortStableFunc(drawn, by.nearerFirst)
	notes := []string{fmt.Sprintf("firang draws: %d stars, %d draws each; "+
		"shares are of the draws asked for, near is within %g pc, screen "+
		"within %g pc", len(drawn), by.draws, by.nearPc, by.screenPc)}
	notes = append(notes, left...)
	var lines [][]string
	for _, star := range drawn {
		lines = append(lines, star.line(by))
	}
	return table{notes: notes, columns: drawsColumns, lines: lines}
}

// nearerFirst orders drawn stars by their share near, the most first,
// then by their middle pass, the nearest first.
func (by picked) nearerFirst(one, other drawnStar) int {
	oneShare := one.within(by.nearPc, by.draws)
	otherShare := other.within(by.nearPc, by.draws)
	if oneShare != otherShare {
		return compared(otherShare, oneShare)
	}
	return compared(one.quantile(midShare), other.quantile(midShare))
}

// compared is before, same or after, as slices' sorting has them, for
// one number against another.
func compared(one, other float64) int {
	switch {
	case one < other:
		return before
	case one > other:
		return after
	}
	return same
}

// The orders slices' sorting reads, each defined.
const (
	before = -1 // defined: sorts first
	after  = 1  // defined: sorts after
	same   = 0  // defined: sorts with
)

// line is a drawn star's line, in the draws table's columns.
func (star drawnStar) line(by picked) []string {
	source := star.source
	mas := masOf(source.Parallax())
	numbers := []float64{star.star.nearFrame, star.star.nearModel,
		star.within(by.nearPc, by.draws), star.within(by.screenPc, by.draws),
		star.quantile(lowShare), star.quantile(midShare),
		star.quantile(highShare)}
	counts := []int{len(star.passes), star.unplaced, star.ranAway}
	placed := []float64{star.away, star.speed, star.outward,
		mas / masOf(source.ParallaxError()), whole / mas,
		source.RadialVelocity(), source.RadialVelocityError(),
		source.Magnitude(), source.RUWE(),
		source.Place().Ascension().InDegrees(),
		source.Place().Declination().InDegrees()}
	fields := []string{fmt.Sprint(source.ID())}
	for _, each := range numbers {
		fields = append(fields, fmt.Sprint(each))
	}
	for _, each := range counts {
		fields = append(fields, fmt.Sprint(each))
	}
	for _, each := range placed {
		fields = append(fields, fmt.Sprint(each))
	}
	return fields
}
