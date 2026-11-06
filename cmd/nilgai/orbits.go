package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"strconv"

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
	"github.com/janearc/puzzlemouth/numbers"
	"github.com/janearc/puzzlemouth/pzm"
)

// orbitsNumbers are the numbers the verb runs by when a call gives
// none, as orbits.yaml has them.
//
//go:embed orbits.yaml
var orbitsNumbers []byte

// orbitsFile is the name the numbers are read under.
const orbitsFile = "orbits.yaml"

// The verb's own numbers, each defined.
const (
	notKept        = -1   // defined: the speed of a row that is not kept
	wholeShare     = 1    // defined: the whole of the kept stars
	noShare        = 0    // defined: none of them
	masInArcsecond = 1000 // defined: a thousand milliarcseconds in one
	million        = 1e6  // defined: a million years in one
	oneStar        = 1    // defined: the least selected, and the least left
	noRows         = 0    // defined: a catalogue of no rows
	inPlane        = 0    // defined: no height, no width: on the x axis
	before         = -1   // defined: sorts first, as slices.SortFunc has it
	after          = 1    // defined: sorts after
	same           = 0    // defined: sorts with
)

// The two ways the seam between the frame's Sun and the model's may be
// closed, each named as --sun takes it.
const (
	sunAsPublished = "frame" // the frame's Sun as the ledger has it
	sunAsTheModel  = "model" // the frame's Sun at the model's speed
)

// acrossPerMasYearKpc is the speed in kilometres per second of a star at
// a kiloparsec whose proper motion is a milliarcsecond a year: an
// astronomical unit a Julian year.
var acrossPerMasYearKpc = length.AU / si.Kilo(length.Metre) /
	(length.LightYear / si.Speed)

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

// ledger is what the verb takes from celestial's research ledger: the
// galaxy the stars run back through and the frame they are placed in,
// each as published. Tests give their own.
type ledger interface {
	Galaxy() (*orbit.MilkyWay, error)
	Frame() (*galactocentric.Frame, error)
}

// product is the catalogue the stars are read from, as celestial-data's
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

// orbitsWith returns the orbits verb, opening its solver and reading its
// ledger as given, its numbers from the yaml given. The real tool opens
// the Metal kernel, reads celestial's ledger and embeds orbits.yaml.
func orbitsWith(open opener, read ledger, raw []byte) func(pzm.Call) error {
	return runner{open, read, raw}.orbits
}

// runner holds what an orbits call opens its solver with, reads its
// ledger from and takes its numbers from.
type runner struct {
	open opener
	read ledger
	raw  []byte
}

// orbits is the verb: the product's fastest stars run back through the
// Milky Way to their nearest pass by Sgr A*, one line a star.
func (runs runner) orbits(call pzm.Call) error {
	asked, err := orbitsAskedFor(call, runs.raw)
	if err != nil {
		return err
	}
	if call.DryRun() {
		fmt.Fprintln(call.Out(), asked.plan())
		return nil
	}
	galaxy, err := runs.read.Galaxy()
	if err != nil {
		return err
	}
	published, err := runs.read.Frame()
	if err != nil {
		return err
	}
	frame, err := frameFor(asked.sun, published, *galaxy)
	if err != nil {
		return err
	}
	file, err := os.Open(asked.product)
	if err != nil {
		return err
	}
	defer file.Close()
	catalogue, err := gaia.OpenCatalogue(file)
	if err != nil {
		return err
	}
	made, err := runs.open.Open(*galaxy, asked.batch)
	if err != nil {
		return err
	}
	defer made.Close()
	return asked.run(catalogue, frame, made, call.Out())
}

// orbitsAsked is what an orbits call asks for.
type orbitsAsked struct {
	product, sun          string
	share, quality, years float64
	steps, batch          int
}

// orbitsAskedFor reads the call's flags, each number from the call or
// from the numbers given, orbits.yaml's. It refuses a call without
// --gaia or --sun, a --sun that is neither frame nor model, a share of
// none or past the whole, and steps, a batch or a span of none.
func orbitsAskedFor(call pzm.Call, raw []byte) (asked orbitsAsked,
	err error) {
	if !call.Given("gaia") {
		return asked, errors.New("orbits needs --gaia, the product")
	}
	asked.product = call.Value("gaia")
	asked.sun = call.Value("sun")
	if asked.sun != sunAsPublished && asked.sun != sunAsTheModel {
		return asked, errors.New("orbits needs --sun frame or --sun " +
			"model: the frame's Sun and the model's differ by " +
			"20 km/s along the rotation, and which to give the " +
			"stars is a choice")
	}
	picked, err := numbers.Read(orbitsFile, raw)
	if err != nil {
		return asked, err
	}
	for name, into := range map[string]*float64{"fastest": &asked.share,
		"quality": &asked.quality, "span": &asked.years} {
		if *into, err = numberOr(call, name, picked); err != nil {
			return asked, err
		}
	}
	for name, into := range map[string]*int{"steps": &asked.steps,
		"batch": &asked.batch} {
		if *into, err = countOr(call, name, picked); err != nil {
			return asked, err
		}
	}
	return asked, asked.check()
}

// orbitNames are the numbers' names in orbits.yaml, by flag.
var orbitNames = map[string]string{"fastest": "fastest-share",
	"quality": "quality-floor", "span": "span-years", "steps": "steps",
	"batch": "batch"}

// numberOr is a flag's number when the call gives it, and the picked
// number when it does not.
func numberOr(call pzm.Call, name string,
	picked numbers.Numbers) (float64, error) {
	if call.Given(name) {
		value, err := strconv.ParseFloat(call.Value(name), 64)
		if err != nil {
			return value, fmt.Errorf("--%s: %w", name, err)
		}
		return value, nil
	}
	number, err := picked.Get(orbitNames[name])
	if err != nil {
		return notKept, err
	}
	return number.Value, nil
}

// countOr is a flag's whole number when the call gives it, and the
// picked number when it does not.
func countOr(call pzm.Call, name string, picked numbers.Numbers) (int,
	error) {
	if call.Given(name) {
		value, err := strconv.Atoi(call.Value(name))
		if err != nil {
			return value, fmt.Errorf("--%s: %w", name, err)
		}
		return value, nil
	}
	number, err := picked.Get(orbitNames[name])
	if err != nil {
		return notKept, err
	}
	return int(number.Value), nil
}

// check refuses what cannot run: a share of none or past the whole, a
// quality floor below none, and steps, a batch or a span of none.
func (asked orbitsAsked) check() error {
	if !(asked.share > noShare && asked.share <= wholeShare) {
		return fmt.Errorf("--fastest %g: a share is over none and at "+
			"most one", asked.share)
	}
	if !(asked.quality >= noShare) {
		return fmt.Errorf("--quality %g: a floor is none or more",
			asked.quality)
	}
	if !(asked.years > noShare) || asked.steps <= noRows ||
		asked.batch <= noRows {
		return fmt.Errorf("--span %g, --steps %d, --batch %d: each is "+
			"over none", asked.years, asked.steps, asked.batch)
	}
	return nil
}

// plan says what the call would do, for a dry run.
func (asked orbitsAsked) plan() string {
	return fmt.Sprintf("would open %s, keep the stars whose parallax "+
		"is at least %g times its error and that have a radial "+
		"velocity, select the fastest %g of them by their speed "+
		"against the Sun, and run those back %g years in %d steps, "+
		"%d at a time, through the Milky Way with the Sun as the %s "+
		"has it", asked.product, asked.quality, asked.share,
		asked.years, asked.steps, asked.batch, asked.sun)
}

// milkyWay is the galaxy the stars run back through: MWPotential2014
// with a hole of so many suns at its centre. It refuses a mass the
// physics cannot hold.
func milkyWay(suns float64) (*orbit.MilkyWay, error) {
	hole, err := grav.Point(measure.SolarMasses(suns))
	if err != nil {
		return nil, err
	}
	return orbit.NewMilkyWay(orbit.MWPotential2014(), *hole)
}

// frameFor is the frame with its Sun as published or given the model's
// speed along the rotation, as --sun asks.
func frameFor(sun string, frame *galactocentric.Frame,
	galaxy orbit.MilkyWay) (*galactocentric.Frame, error) {
	if sun == sunAsPublished {
		return frame, nil
	}
	return atTheModelsSpeed(frame, galaxy)
}

// atTheModelsSpeed is the frame with its Sun going along the rotation
// at the model's circular speed at the Sun's distance plus the Sun's
// own motion with respect to the local standard of rest. It refuses a
// model whose Sun would go past light.
//
// The Sun's motion toward the centre and toward the pole stays the
// frame's: U as measured, W as Sgr A*'s apparent motion reflects it.
// So under --sun model, V is the model's and W is still the frame's.
func atTheModelsSpeed(frame *galactocentric.Frame,
	galaxy orbit.MilkyWay) (*galactocentric.Frame, error) {
	place, moving := frame.Sun()
	alongRotation := circular(galaxy, sunward(place)) +
		research.Value("sun-motion-v")
	return frame.MovingSun(moving.AlongX().InKilometresPerSecond(),
		alongRotation, moving.AlongZ().InKilometresPerSecond())
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

// run is the two passes: every row's speed, then the fastest share back
// through the solver, written one line a star, sorted by nearest pass,
// under a header that says what was kept, selected and left out.
func (asked orbitsAsked) run(catalogue product,
	frame *galactocentric.Frame, made solver, out io.Writer) error {
	speeds, err := speedsOf(catalogue, asked.quality)
	if err != nil {
		return err
	}
	cut, kept := cutAt(speeds, asked.share)
	results, left, err := asked.back(catalogue, speeds, cut, frame, made)
	if err != nil {
		return err
	}
	slices.SortFunc(results, byNearest)
	counted := fmt.Sprintf("%d of %d rows kept, %d at %g km/s or more "+
		"selected", kept, len(speeds), selected(speeds, cut), cut)
	notes := append([]string{counted}, left...)
	return asked.write(out, notes, results)
}

// speedsOf is pass one: each row's speed against the Sun, the
// transverse and the radial in quadrature, or less than none where
// the row is not kept.
//
// A row is not kept for a parallax that is not positive or under the
// quality floor of its error, for no radial velocity, or for a speed
// that is not a number, as a proper motion the archive left empty
// gives.
func speedsOf(catalogue product, quality float64) ([]float32, error) {
	speeds := make([]float32, catalogue.Rows())
	for row := range catalogue.Rows() {
		source, err := catalogue.Source(row)
		if err != nil {
			return nil, err
		}
		speeds[row] = notKept
		mas := source.Parallax().InArcseconds() * masInArcsecond
		floor := source.ParallaxError().InArcseconds() * masInArcsecond
		radial := source.RadialVelocity()
		if !(mas > noShare) || !(mas >= quality*floor) ||
			math.IsNaN(radial) {
			continue
		}
		motion := source.Motion()
		across := acrossPerMasYearKpc * math.Hypot(motion.Ascension(),
			motion.Declination()) / mas
		if speed := math.Hypot(across, radial); !math.IsNaN(speed) {
			speeds[row] = float32(speed)
		}
	}
	return speeds, nil
}

// cutAt is the speed the fastest share of the kept rows are at or
// over, and how many rows are kept. With none kept the cut is above
// any speed, so nothing is selected.
//
// The share is counted by nearest integer, a half rounding up, and is
// at least one star: of ten, 0.7 is seven, where the doubles'
// 3.0000000000000004 left would have made a ceiling four.
func cutAt(speeds []float32, share float64) (cut float32, kept int) {
	var sorted []float32
	for _, speed := range speeds {
		if speed >= noShare {
			sorted = append(sorted, speed)
		}
	}
	if len(sorted) == noRows {
		return float32(math.Inf(wholeShare)), noRows
	}
	slices.Sort(sorted)
	count := float64(len(sorted))
	rounded := math.Round(count * share)
	chosen := max(oneStar, int(rounded))
	return sorted[len(sorted)-chosen], len(sorted)
}

// selected is how many rows are at the cut or over it.
func selected(speeds []float32, cut float32) int {
	count := noRows
	for _, speed := range speeds {
		if speed >= cut {
			count++
		}
	}
	return count
}

// result is one star's nearest pass, with what it was run from.
type result struct {
	id                       int64
	near, ago, moved, pulled float64
	speed                    float32
	source                   *gaia.Source
}

// byNearest orders results by their nearest pass, the nearest first.
func byNearest(one, other result) int {
	switch {
	case one.near < other.near:
		return before
	case one.near > other.near:
		return after
	}
	return same
}

// back is pass two: the selected rows through the frame and the solver
// in batches, the last batch whatever is left, which may be nothing.
// It gives each star's result, and a note for each star left out: one
// the frame refused, or one whose orbit ran away.
func (asked orbitsAsked) back(catalogue product, speeds []float32,
	cut float32, frame *galactocentric.Frame, made solver) ([]result,
	[]string, error) {
	var results []result
	var left []string
	var batch []result
	var stars []orbit.State
	for row, speed := range speeds {
		if speed < cut {
			continue
		}
		source, err := catalogue.Source(int64(row))
		if err != nil {
			return nil, nil, err
		}
		state, err := source.State(*frame)
		if err != nil {
			refused := fmt.Sprintf("star %d is left out: %v",
				source.ID(), err)
			left = append(left, refused)
			continue
		}
		batch = append(batch, result{id: source.ID(), speed: speed,
			source: source})
		stars = append(stars, *state)
		if len(batch) < asked.batch {
			continue
		}
		if results, left, err = asked.ran(made, batch, stars, results,
			left); err != nil {
			return nil, nil, err
		}
		batch, stars = nil, nil
	}
	return asked.ran(made, batch, stars, results, left)
}

// ran runs a batch and appends its results, and a note for each of its
// stars that ran away, to those so far. An empty batch runs nothing.
func (asked orbitsAsked) ran(made solver, batch []result,
	stars []orbit.State, results []result, left []string) ([]result,
	[]string, error) {
	done, away, err := asked.backing(made, batch, stars)
	if err != nil {
		return nil, nil, err
	}
	return append(results, done...), append(left, away...), nil
}

// backing runs a batch of stars back. The stars whose orbits ran away,
// which the solver names, are left out with a note each and the rest
// run again; any other refusal of the solver is the run's error.
func (asked orbitsAsked) backing(made solver, batch []result,
	stars []orbit.State) ([]result, []string, error) {
	var left []string
	for len(batch) > noRows {
		passes, _, err := made.Back(stars, measure.Years(asked.years),
			asked.steps)
		if err == nil {
			return withPasses(batch, passes), left, nil
		}
		away, ran := gpuorbit.RanAway(err)
		if !ran {
			return nil, nil, err
		}
		if batch, stars, left, err = without(batch, stars, away,
			left); err != nil {
			return nil, nil, err
		}
	}
	return nil, left, nil
}

// without is the batch and its stars without those the solver named
// as run away, each named in a note. It refuses a star the batch does
// not have.
func without(batch []result, stars []orbit.State, away []int,
	left []string) ([]result, []orbit.State, []string, error) {
	gone := map[int]bool{}
	for _, which := range away {
		if which < noRows || which >= len(batch) {
			why := fmt.Errorf("the solver named star %d of a "+
				"batch of %d", which, len(batch))
			return nil, nil, nil, why
		}
		gone[which] = true
		left = append(left, fmt.Sprintf("star %d is left out: its "+
			"orbit ran away", batch[which].id))
	}
	var kept []result
	var still []orbit.State
	for which := range batch {
		if !gone[which] {
			kept = append(kept, batch[which])
			still = append(still, stars[which])
		}
	}
	return kept, still, left, nil
}

// withPasses is the batch with each star's pass written into its
// result, in parsecs and millions of years.
func withPasses(batch []result, passes []orbit.Pass) []result {
	for which, pass := range passes {
		batch[which].near = pass.Near().InMetres() / length.Parsec
		batch[which].ago = pass.Ago().InYears() / million
		batch[which].moved, batch[which].pulled = pass.Steps()
	}
	return batch
}

// write writes the header, each note on a line of its own, and one
// line a star: the id, the nearest pass in parsecs and how long ago in
// millions of years, the pass's step shares, the speed, and the star as
// the catalogue had it.
func (asked orbitsAsked) write(out io.Writer, notes []string,
	results []result) error {
	lines := []string{fmt.Sprintf("# nilgai orbits: %s; the fastest %g, "+
		"%d stars; sun as the %s has it", asked.product, asked.share,
		len(results), asked.sun)}
	for _, note := range notes {
		lines = append(lines, "# "+note)
	}
	lines = append(lines, "source_id\tnear_pc\tago_myr\tmoved\tpulled\t"+
		"speed_kms\tra_deg\tdec_deg\tparallax_mas\tpmra_masyr\t"+
		"pmdec_masyr\trv_kms")
	for _, each := range results {
		lines = append(lines, each.line())
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(out, line); err != nil {
			return err
		}
	}
	return nil
}

// line is a result's line, tab-separated.
func (each result) line() string {
	star := each.source
	return fmt.Sprintf("%d\t%g\t%g\t%g\t%g\t%g\t%g\t%g\t%g\t%g\t%g\t%g",
		each.id, each.near, each.ago, each.moved, each.pulled,
		each.speed, star.Place().Ascension().InDegrees(),
		star.Place().Declination().InDegrees(),
		star.Parallax().InArcseconds()*masInArcsecond,
		star.Motion().Ascension(), star.Motion().Declination(),
		star.RadialVelocity())
}
