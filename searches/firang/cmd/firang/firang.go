package main

import (
	"fmt"
	"math"
	"os"

	"github.com/janearc/celestial-data/gaia"
	"github.com/janearc/celestial/galactocentric"
	"github.com/janearc/celestial/gpu"
	gpuorbit "github.com/janearc/celestial/gpu/orbit"
	"github.com/janearc/celestial/grav"
	"github.com/janearc/celestial/length"
	"github.com/janearc/celestial/measure"
	"github.com/janearc/celestial/orbit"
	"github.com/janearc/celestial/research"
	"github.com/janearc/celestial/si"
)

// The command's own numbers, each defined.
const (
	masInArcsecond = 1000 // defined: a thousand milliarcseconds in one
	inAMillion     = 1e6  // defined: years in a million
	nothing        = 0    // defined: no length, no speed, no rows
	inPlane        = 0    // defined: no height, no width: on the x axis
	firstStar      = 0    // defined: where a batch of stars starts
)

// solver runs stars back through a field to their nearest pass by the
// centre, as gpu/orbit's does.
type solver interface {
	Back(stars []orbit.State, span measure.Duration,
		steps int) ([]orbit.Pass, []orbit.State, error)
	Close()
}

// opener opens a solver over the galaxy with room for so many stars.
type opener interface {
	Open(galaxy orbit.MilkyWay, most int) (solver, error)
}

// ledger is what firang takes from celestial's research ledger: the
// galaxy stars run through and the frame they are placed in. Tests give
// their own.
type ledger interface {
	Galaxy() (*orbit.MilkyWay, error)
	Frame() (*galactocentric.Frame, error)
}

// product is the catalogue stars are read from, as celestial-data's
// gaia.Catalogue is.
type product interface {
	Rows() int64
	Source(index int64) (*gaia.Source, error)
}

// metal opens the kernel on the Metal device.
type metal struct{}

// Open compiles gpu/orbit's kernel over the galaxy on the Metal device.
func (metal) Open(galaxy orbit.MilkyWay, most int) (solver, error) {
	made, err := gpuorbit.Open(gpu.Metal(), gpuorbit.MilkyWay(galaxy), most)
	if err != nil {
		return nil, err
	}
	return made, nil
}

// published is the ledger as celestial publishes it.
type published struct{}

// Galaxy is MWPotential2014 with Sgr A* at its centre, its mass from
// the ledger.
func (published) Galaxy() (*orbit.MilkyWay, error) {
	return milkyWay(research.Value("sgr-a-mass"))
}

// Frame is the galactocentric frame as the ledger has it.
func (published) Frame() (*galactocentric.Frame, error) {
	return galactocentric.Published()
}

// milkyWay is MWPotential2014 with a hole of so many suns at its
// centre. It refuses a mass the physics cannot hold.
func milkyWay(suns float64) (*orbit.MilkyWay, error) {
	hole, err := grav.Point(measure.SolarMasses(suns))
	if err != nil {
		return nil, err
	}
	return orbit.NewMilkyWay(orbit.MWPotential2014(), *hole)
}

// setting is what a verb runs stars in: the frame with the Sun as
// published, the Sun's motion toward the centre and toward the pole, the
// two estimates of its speed along the rotation, and the solver with
// its span, steps and batch.
type setting struct {
	frame            *galactocentric.Frame
	toCentre, toPole float64
	low, high        float64
	made             solver
	span             measure.Duration
	steps, batch     int
}

// settingFor is the setting from the ledger, with a solver opened over
// its galaxy, by the numbers given. The caller closes the solver.
//
// The two estimates of the Sun's speed along the rotation are the
// published frame's, Sgr A*'s apparent motion reflected, and the model's
// circular speed at the Sun's distance plus the Sun's own motion.
func settingFor(read ledger, open opener, by picked) (*setting, error) {
	galaxy, err := read.Galaxy()
	if err != nil {
		return nil, err
	}
	frame, err := read.Frame()
	if err != nil {
		return nil, err
	}
	place, moving := frame.Sun()
	modelled := circular(*galaxy, sunward(place)) +
		research.Value("sun-motion-v")
	measured := moving.AlongY().InKilometresPerSecond()
	made, err := open.Open(*galaxy, by.batch)
	if err != nil {
		return nil, err
	}
	return &setting{frame: frame,
		toCentre: moving.AlongX().InKilometresPerSecond(),
		toPole:   moving.AlongZ().InKilometresPerSecond(),
		low:      math.Min(modelled, measured),
		high:     math.Max(modelled, measured), made: made,
		span: measure.Years(by.years), steps: by.steps, batch: by.batch}, nil
}

// sunward is how far a place is from the centre along the x axis
// toward the Sun, in kiloparsecs: the Sun's own distance, for the Sun.
func sunward(place measure.Coord) float64 {
	by := place.Minus(measure.Origin())
	return -by.AlongX().InMetres() / si.Kilo(length.Parsec)
}

// circular is the galaxy's circular speed in the plane so far from the
// centre, in kilometres per second: the root of the distance times the
// pull there, which points to the centre, so its x part is negative.
func circular(galaxy orbit.MilkyWay, far float64) float64 {
	x, _, _ := galaxy.Pull(orbit.PlaceOf(far, inPlane, inPlane)).Along()
	return math.Sqrt(-far * x)
}

// movingAt is the frame with the Sun going so fast along the rotation,
// its motion toward the centre and toward the pole as published.
func (at setting) movingAt(alongRotation float64) (*galactocentric.Frame,
	error) {
	return at.frame.MovingSun(at.toCentre, alongRotation, at.toPole)
}

// estimates are the two frames: the Sun at the published speed along
// the rotation, and at the model's.
func (at setting) estimates() (*galactocentric.Frame,
	*galactocentric.Frame, error) {
	_, moving := at.frame.Sun()
	measured := moving.AlongY().InKilometresPerSecond()
	modelled := at.low + at.high - measured
	model, err := at.movingAt(modelled)
	if err != nil {
		return nil, nil, err
	}
	return at.frame, model, nil
}

// outcome is a star run through the solver: its nearest pass in
// parsecs, where it ended, and whether it ran away instead.
type outcome struct {
	near    float64
	end     orbit.State
	ranAway bool
}

// run is each star run back over the span given in the steps given, in
// batches. A star that ran away is marked so; any other refusal of the
// solver is the error.
func (at setting) run(stars []orbit.State, span measure.Duration,
	steps int) ([]outcome, error) {
	outcomes := make([]outcome, len(stars))
	for first := firstStar; first < len(stars); first += at.batch {
		last := min(first+at.batch, len(stars))
		err := at.runBatch(stars[first:last], outcomes[first:last], span,
			steps)
		if err != nil {
			return nil, err
		}
	}
	return outcomes, nil
}

// back is each star's nearest pass in parsecs over the setting's span,
// not a number for a star that ran away.
func (at setting) back(stars []orbit.State) ([]float64, error) {
	outcomes, err := at.run(stars, at.span, at.steps)
	if err != nil {
		return nil, err
	}
	near := make([]float64, len(outcomes))
	for which, each := range outcomes {
		near[which] = each.near
		if each.ranAway {
			near[which] = math.NaN()
		}
	}
	return near, nil
}

// runBatch runs one batch, writing each star's outcome in its place. The
// stars the solver names as run away are marked and the rest run again.
func (at setting) runBatch(stars []orbit.State, outcomes []outcome,
	span measure.Duration, steps int) error {
	places := make([]int, len(stars))
	for which := range places {
		places[which] = which
	}
	for len(stars) > nothing {
		passes, ends, err := at.made.Back(stars, span, steps)
		if err == nil {
			for which, pass := range passes {
				near := pass.Near().InMetres() / length.Parsec
				outcomes[places[which]] = outcome{near: near, end: ends[which]}
			}
			return nil
		}
		away, ran := gpuorbit.RanAway(err)
		if !ran {
			return err
		}
		stars, places, err = without(stars, places, away, outcomes)
		if err != nil {
			return err
		}
	}
	return nil
}

// without is the stars and their places with those the solver named as
// run away taken out, each marked in its outcome. It refuses a star the
// batch does not have.
func without(stars []orbit.State, places, away []int,
	outcomes []outcome) ([]orbit.State, []int, error) {
	gone := map[int]bool{}
	for _, which := range away {
		if which < nothing || which >= len(stars) {
			return nil, nil, fmt.Errorf("the solver named star %d of a "+
				"batch of %d", which, len(stars))
		}
		gone[which] = true
		outcomes[places[which]].ranAway = true
	}
	var still []orbit.State
	var stillPlaces []int
	for which := range stars {
		if !gone[which] {
			still = append(still, stars[which])
			stillPlaces = append(stillPlaces, places[which])
		}
	}
	return still, stillPlaces, nil
}

// kept says whether a star is worth running: a parallax positive and
// at least the floor times its error, a radial velocity, and a proper
// motion, each a number.
func kept(source *gaia.Source, quality float64) bool {
	mas := masOf(source.Parallax())
	floor := masOf(source.ParallaxError())
	motion := source.Motion()
	numbers := []float64{source.RadialVelocity(), motion.Ascension(),
		motion.Declination()}
	for _, each := range numbers {
		if math.IsNaN(each) {
			return false
		}
	}
	return mas > nothing && mas >= quality*floor
}

// masOf is an angle in milliarcseconds.
func masOf(angle measure.Angle) float64 {
	return angle.InArcseconds() * masInArcsecond
}

// opened is the catalogue in the product at the path given, with the
// file to close once the verb is done.
func opened(path string) (*gaia.Catalogue, *os.File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	catalogue, err := gaia.OpenCatalogue(file)
	if err != nil {
		file.Close()
		return nil, nil, err
	}
	return catalogue, file, nil
}
