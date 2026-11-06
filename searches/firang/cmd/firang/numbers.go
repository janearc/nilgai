package main

import (
	"fmt"
	"os"

	"github.com/janearc/puzzlemouth/numbers"
)

// picked is the numbers firang runs by, as numbers.yaml has them.
type picked struct {
	quality, years, screenPc, nearPc    float64
	speed, share, rvError               float64
	least, most, fromPc                 float64
	flightLeast, flightMost, flightStep float64
	steps, batch, draws, thrown         int
	seed, validationSeed                uint64
}

// pickedFrom reads the numbers in the file at the path given. It
// refuses a file it cannot read, numbers it cannot parse, a number it
// lacks, and numbers that cannot run.
func pickedFrom(path string) (*picked, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	read, err := numbers.Read(path, raw)
	if err != nil {
		return nil, err
	}
	var made picked
	floats := map[string]*float64{"quality-floor": &made.quality,
		"span-years": &made.years, "screen-pc": &made.screenPc,
		"near-pc": &made.nearPc, "candidate-speed": &made.speed,
		"candidate-share": &made.share, "candidate-rv-error": &made.rvError,
		"thrown-least": &made.least, "thrown-most": &made.most,
		"thrown-from-pc": &made.fromPc, "flight-least-myr": &made.flightLeast,
		"flight-most-myr": &made.flightMost,
		"flight-step-myr": &made.flightStep}
	for name, into := range floats {
		if *into, err = valueOf(read, name); err != nil {
			return nil, err
		}
	}
	counts := map[string]*int{"steps": &made.steps, "batch": &made.batch,
		"draws": &made.draws, "thrown": &made.thrown}
	for name, into := range counts {
		value, err := valueOf(read, name)
		if err != nil {
			return nil, err
		}
		*into = int(value)
	}
	seeds := map[string]*uint64{"seed": &made.seed,
		"validation-seed": &made.validationSeed}
	for name, into := range seeds {
		value, err := valueOf(read, name)
		if err != nil {
			return nil, err
		}
		*into = uint64(value)
	}
	return &made, made.check()
}

// valueOf is a number's value by name, or the reading's refusal.
func valueOf(read numbers.Numbers, name string) (float64, error) {
	number, err := read.Get(name)
	if err != nil {
		return nothing, err
	}
	return number.Value, nil
}

// check refuses numbers that cannot run: counts and lengths of none or
// less, a quality floor below none, a share past the whole, thrown
// speeds or flights the wrong way round, and a flight past the span.
func (by picked) check() error {
	for name, count := range map[string]int{"steps": by.steps,
		"batch": by.batch, "draws": by.draws, "thrown": by.thrown} {
		if count <= nothing {
			return fmt.Errorf("%s is %d; it is over none", name, count)
		}
	}
	for name, length := range map[string]float64{"span-years": by.years,
		"screen-pc": by.screenPc, "near-pc": by.nearPc,
		"thrown-from-pc": by.fromPc, "flight-step-myr": by.flightStep} {
		if !(length > nothing) {
			return fmt.Errorf("%s is %g; it is over none", name, length)
		}
	}
	if !(by.quality >= nothing) || !(by.share >= nothing && by.share <= whole) {
		return fmt.Errorf("quality-floor %g and candidate-share %g: a floor "+
			"is none or more, a share none to the whole", by.quality,
			by.share)
	}
	if !(by.least <= by.most) || !(by.flightLeast <= by.flightMost) ||
		!(by.flightMost*inAMillion <= by.years) {
		return fmt.Errorf("thrown %g to %g km/s, flights %g to %g Myr in "+
			"%g years: each the right way round, inside the span", by.least,
			by.most, by.flightLeast, by.flightMost, by.years)
	}
	return nil
}
