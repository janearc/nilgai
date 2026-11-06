package main

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/astrogo/fitsio"
	"github.com/janearc/celestial-data/gaia"
	"github.com/janearc/celestial/galactocentric"
	gpuorbit "github.com/janearc/celestial/gpu/orbit"
	"github.com/janearc/celestial/grav"
	"github.com/janearc/celestial/measure"
	"github.com/janearc/celestial/orbit"
	"github.com/janearc/celestial/research"
	"github.com/janearc/nilgai/internal/fitstest"
	"go.uber.org/mock/gomock"
)

// The made-up stars the orbits tests write, each value defined here.
// The parallax at the floor and its error are binary fractions, so
// that a float32 stores them exactly and five times the error is the
// parallax to the bit.
const (
	madeRA        = 10.0      // defined: degrees, a benign place on the sky
	madeDec       = 10.0      // defined: degrees
	nearParallax  = 1.0       // defined: mas, a star a kiloparsec away
	nearerParallx = 2.0       // defined: mas, half a kiloparsec away
	floorParallax = 0.625     // defined: mas, five times its error: just kept
	floorError    = 0.125     // defined: mas, a fifth of the floor parallax
	poorParallax  = 0.1       // defined: mas, as big as its error: not kept
	noParallax    = 0.0       // defined: mas, no distance: not kept
	noError       = 0.0       // defined: mas, no error beside no parallax
	smallError    = 0.1       // defined: mas, a tenth of the near parallax
	fastRadial    = 300.0     // defined: km/s, the fastest star's speed
	runawaySpeed  = 400.0     // defined: km/s along y; only the fastest passes
	acrossMotion  = 100.0     // defined: mas a year, a motion across the sky
	slowRadial    = 100.0     // defined: km/s, the cut at half of six
	slowerRadial  = 50.0      // defined: km/s
	slowestRadial = 10.0      // defined: km/s
	atRest        = 0.0       // defined: km/s, a speed of none, still kept
	poorRadial    = 500.0     // defined: km/s, fast but not kept
	pastLight     = 5e5       // defined: km/s, which the frame refuses
	noMotion      = 0.0       // defined: no motion
	exactRadial   = 0.0       // defined: km/s, a radial velocity without error
	halfOfThem    = 0.5       // defined: --fastest, half the kept stars
	allOfThem     = 1         // defined: --fastest, every kept star
	oneOfThem     = 0.01      // defined: --fastest, the single fastest
	sevenOfTen    = 0.7       // defined: --fastest, seven of ten
	anyParallax   = 0         // defined: --quality none, any positive parallax
	threeStars    = 3         // defined: the three selected at half of six
	fourStars     = 4         // defined: half of seven kept, the half rounding up
	twoStars      = 2         // defined: the two left when one runs away
	sixKept       = 6         // defined: the kept stars of ten
	sevenKept     = 7         // defined: the kept stars at --quality none
	sevenChosen   = 7         // defined: seven of ten at --fastest 0.7
	tenSpeeds     = 10        // defined: ten speeds, one to ten
	oneStarRan    = 1         // defined: the one that ran, the other away
	nearA         = 0.5       // defined: kpc, the first star's nearest pass
	nearStep      = 0.05      // defined: kpc, each next star's pass is nearer by
	agoUnits      = 0.001     // defined: units of time, the pass's age
	shareMoved    = 0.01      // defined: a pass's step share
	sharePulled   = 0.001     // defined: a pass's other step share
	fewSteps      = 100       // defined: --steps, few
	fewYears      = 1e6       // defined: --span, a million years
	noShareFlag   = 0         // defined: --fastest none, refused
	pastWhole     = 2         // defined: --fastest past the whole, refused
	noSteps       = 0         // defined: --steps none, refused
	noBatch       = 0         // defined: --batch none, refused
	noSpan        = 0         // defined: --span none, refused
	belowNone     = -1        // defined: --quality below none, refused
	speedRoom     = 1e-6      // defined: km/s, the model's speed to rounding
	lineRoom      = 1e-5      // defined: a star line's numbers, as a share
	twoCalls      = 2         // defined: a batch refused for a runaway, then the rest
	firstBatch    = 0         // defined: the first batch a solver was given
	noSuns        = -1        // defined: a hole of less than no mass, refused
	someSuns      = 4e6       // defined: a hole of a few million suns
	modelSun      = 8         // defined: kpc, the absurd model's sun
	twoRows       = 2         // defined: a mocked catalogue's rows
	oneAtATime    = 1         // defined: --batch one star at a time
	parsecsInKpc  = 1000      // defined: parsecs in a kiloparsec
	inAMillion    = 1e6       // defined: years in a million
	modelCircular = 220       // defined: MWPotential2014's speed at its sun, 8 kpc
	circularRoom  = 0.01      // defined: km/s, over celestial's 220.005
	speedB        = 237.02352 // defined: 4.740470464 * 100 mas/yr / 2 mas
	seamKmS       = 20.0      // defined: 251.55 - 219.3 - 12.24, by hand
	seamRoom      = 0.5       // defined: km/s, the arithmetic's rounding
	lineFields    = 12        // defined: a star line's numbers
	noSuchStar    = 3         // defined: the first place past a batch of three
	sunRoom       = 1e-5      // defined: a share; the sun sunward, to the tilt
)

// notNumbers is what the tests give the verb in place of orbits.yaml
// when it should refuse the numbers: not yaml.
const notNumbers = "numbers: ["

// without is orbits.yaml with one number renamed, so that the verb
// cannot find it.
func withoutNumber(name string) []byte {
	text := string(orbitsNumbers)
	renamed := strings.Replace(text, "name: "+name, "name: gone-"+name,
		oneStarRan)
	return []byte(renamed)
}

// star is a made-up row with an id, a parallax and its error, a motion
// across the sky and a radial velocity, the rest nothing.
func star(id int64, parallax, parError, pmRA,
	radial float64) *fitstest.GaiaStar {
	made := fitstest.NewGaiaStar(id)
	made.At(madeRA, madeDec)
	made.Distant(parallax, parError)
	made.Moving(pmRA, noMotion)
	made.Receding(radial, exactRadial)
	return made
}

// The made-up stars' ids, each defined: a letter's place in the
// alphabet.
const (
	starA = 1  // defined: the fastest, 300 km/s away
	starB = 2  // defined: the second, across the sky at half a kiloparsec
	starC = 3  // defined: slow, 100 km/s: the cut at half of six
	starD = 4  // defined: slower
	starE = 5  // defined: fast but its parallax is noise: not kept
	starF = 6  // defined: no radial velocity: not kept
	starG = 7  // defined: faster than light, which the frame refuses
	starH = 8  // defined: no parallax: not kept, though it moves (an infinite speed otherwise)
	starI = 9  // defined: a parallax just at the floor: kept, slowest
	starJ = 10 // defined: at rest: kept, a speed of none
	starK = 11 // defined: no proper motion, a speed not a number: not kept
)

// tenStars are the made-up stars: six kept, of which the fastest
// three are selected at --fastest 0.5, and four not kept.
func tenStars() []*fitstest.GaiaStar {
	return []*fitstest.GaiaStar{
		star(starA, nearParallax, smallError, noMotion, fastRadial),
		star(starB, nearerParallx, smallError, acrossMotion, noMotion),
		star(starC, nearParallax, smallError, noMotion, slowRadial),
		star(starD, nearParallax, smallError, noMotion, slowerRadial),
		star(starE, poorParallax, poorParallax, noMotion, poorRadial),
		star(starF, nearParallax, smallError, noMotion, math.NaN()),
		star(starH, noParallax, noError, acrossMotion, slowRadial),
		star(starI, floorParallax, floorError, noMotion, slowestRadial),
		star(starJ, nearParallax, smallError, noMotion, atRest),
		star(starK, nearParallax, smallError, math.NaN(), slowRadial)}
}

// productOf writes a product of the rows given into a folder and gives
// its path, or the test stops.
func productOf(test *testing.T, rows ...*fitstest.GaiaStar) string {
	test.Helper()
	path := filepath.Join(test.TempDir(), "gaia.fits")
	file, err := os.Create(path)
	if err != nil {
		test.Fatal(err)
	}
	defer file.Close()
	if err := fitstest.WriteGaia(file, rows...); err != nil {
		test.Fatal(err)
	}
	return path
}

// tenProduct is a product of the ten stars, written, or the test
// stops.
func tenProduct(test *testing.T) string {
	test.Helper()
	return productOf(test, tenStars()...)
}

// answering is a solver's answer to a batch: a pass for each star, the
// first star's the farthest, each next nearer by the same step, and
// the batches it was given.
type answering struct {
	batches [][]orbit.State
}

// back answers a batch with passes, nearer star by star.
func (answers *answering) back(stars []orbit.State, _ measure.Duration,
	_ int) ([]orbit.Pass, []orbit.State, error) {
	answers.batches = append(answers.batches, stars)
	var passes []orbit.Pass
	for which := range stars {
		near := nearA - nearStep*float64(which)
		passes = append(passes, orbit.NewPass(near, agoUnits, shareMoved,
			sharePulled))
	}
	return passes, stars, nil
}

// running is a mocked opener and solver: the opener gives the solver
// once, the solver closes once.
func running(test *testing.T) (*Mockopener, *Mocksolver) {
	test.Helper()
	controller := gomock.NewController(test)
	open, made := NewMockopener(controller), NewMocksolver(controller)
	open.EXPECT().Open(gomock.Any(), gomock.Any()).Return(made, nil)
	made.EXPECT().Close()
	return open, made
}

// answered is a mocked opener whose solver answers any batch, and the
// answers, or the test stops.
func answered(test *testing.T) (*Mockopener, *answering) {
	test.Helper()
	open, made := running(test)
	answers := &answering{}
	made.EXPECT().Back(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		answers.back).AnyTimes()
	return open, answers
}

// orbitsRan runs the verb with the arguments after it, on the opener,
// the ledger and the numbers given, and gives the exit code and what it
// wrote and said.
func orbitsRan(open opener, read ledger, raw []byte,
	args ...string) (int, string, string) {
	var out, errs bytes.Buffer
	words := append([]string{"orbits"}, args...)
	code := toolWith("", realSearches(), nil, open, read, raw).RunWith(
		words, nothingToRead(), &out, &errs)
	return code, out.String(), errs.String()
}

// starLines are the lines of what orbits wrote that are stars: those
// after the header's lines.
func starLines(out string) []string {
	var stars []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if !strings.HasPrefix(line, "#") && !strings.HasPrefix(line,
			"source_id") {
			stars = append(stars, line)
		}
	}
	return stars
}

// lineRead is a star's line as written, read back.
type lineRead struct {
	id                                     int64
	near, ago, moved, pulled, speed        float64
	ra, dec, parallax, pmRA, pmDec, radial float64
}

// readLine reads a star's line, or the test stops.
func readLine(test *testing.T, line string) lineRead {
	test.Helper()
	var got lineRead
	count, err := fmt.Sscanf(line, "%d %g %g %g %g %g %g %g %g %g %g %g",
		&got.id, &got.near, &got.ago, &got.moved, &got.pulled, &got.speed,
		&got.ra, &got.dec, &got.parallax, &got.pmRA, &got.pmDec, &got.radial)
	if err != nil || count != lineFields {
		test.Fatalf("the line %q reads %d numbers, %v", line, count, err)
	}
	return got
}

// pair is a number read back beside the number the hand works.
type pair struct{ got, want float64 }

// near says whether a number is within a share of another.
func near(got, want, share float64) bool {
	return math.Abs(got-want) <= math.Abs(want)*share
}

// Example_orbits says what orbits would do, and reads nothing.
func Example_orbits() {
	var out, errs bytes.Buffer
	code := toolWith("", realSearches(), nil, metal{}, published{},
		orbitsNumbers).RunWith([]string{"orbits", "--gaia", "gaia.fits",
		"--sun", "frame", "--dry-run"}, nothingToRead(), &out, &errs)
	fmt.Print(code, " ", out.String())
	// Output:
	// 0 would open gaia.fits, keep the stars whose parallax is at least 5 times its error and that have a radial velocity, select the fastest 0.1 of them by their speed against the Sun, and run those back 1e+07 years in 10000 steps, 65536 at a time, through the Milky Way with the Sun as the frame has it
}

// the fastest half of the six kept stars, the three, go through the
// solver as one batch and come out one line a star, the nearest pass
// first: the third star, then the second, then the first; the header
// says what was kept and selected, the cut at the third's speed, and
// no star is left out.
func TestTheFastestShareIsRunBackAndWrittenNearestFirst(test *testing.T) {
	open, made := running(test)
	answers := &answering{}
	made.EXPECT().Back(gomock.Len(threeStars), measure.Years(fewYears),
		fewSteps).DoAndReturn(answers.back)
	code, out, errs := orbitsRan(open, published{}, orbitsNumbers, "--gaia",
		tenProduct(test), "--sun", "frame", "--fastest",
		fmt.Sprint(halfOfThem), "--span", fmt.Sprint(fewYears), "--steps",
		fmt.Sprint(fewSteps))
	if code != exitRan {
		test.Fatalf("exit %d: %s", code, errs)
	}
	stars := starLines(out)
	if len(stars) != threeStars {
		test.Fatalf("wrote %d stars: %s", len(stars), out)
	}
	var kept, rows, selected int
	var cut float64
	header := strings.Split(out, "\n")[oneStarRan]
	if _, err := fmt.Sscanf(header, "# %d of %d rows kept, %d at %g km/s",
		&kept, &rows, &selected, &cut); err != nil || kept != sixKept ||
		rows != len(tenStars()) || selected != threeStars || cut != slowRadial {
		test.Errorf("the header says %q, %v", header, err)
	}
	if strings.Contains(out, "left out") {
		test.Errorf("a star is left out with none refused: %s", out)
	}
	for which, want := range []int64{starC, starB, starA} {
		if !strings.HasPrefix(stars[which], fmt.Sprint(want)+"\t") {
			test.Errorf("star %d is not %d: %s", which, want, stars[which])
		}
	}
}

// a star's line carries its pass in parsecs and millions of years, its
// step shares, its speed and the star as the catalogue had it, each as
// the hand works it: the first star's pass is half a kiloparsec, its
// speed its radial velocity; the second's speed is its proper motion
// at its distance.
func TestAStarsLineIsWhatTheHandWorks(test *testing.T) {
	open, _ := answered(test)
	code, out, errs := orbitsRan(open, published{}, orbitsNumbers, "--gaia",
		tenProduct(test), "--sun", "frame", "--fastest",
		fmt.Sprint(halfOfThem))
	if code != exitRan {
		test.Fatalf("exit %d: %s", code, errs)
	}
	lines := map[int64]lineRead{}
	for _, line := range starLines(out) {
		got := readLine(test, line)
		lines[got.id] = got
	}
	first, second := lines[starA], lines[starB]
	ago := orbit.Units(agoUnits).InYears() / inAMillion
	wants := map[string]pair{
		"near":     {first.near, nearA * parsecsInKpc},
		"ago":      {first.ago, ago},
		"moved":    {first.moved, shareMoved},
		"pulled":   {first.pulled, sharePulled},
		"speed":    {first.speed, fastRadial},
		"ra":       {first.ra, madeRA},
		"dec":      {first.dec, madeDec},
		"parallax": {first.parallax, nearParallax},
		"radial":   {first.radial, fastRadial},
		"across":   {second.speed, speedB},
		"pmra":     {second.pmRA, acrossMotion},
		"nearer":   {second.near, (nearA - nearStep) * parsecsInKpc},
	}
	for name, each := range wants {
		if !near(each.got, each.want, lineRoom) {
			test.Errorf("%s is %g, want %g", name, each.got, each.want)
		}
	}
}

// the shares at their edges: all of the kept stars run at --fastest
// 1, the single fastest at a share under one star's, and --quality
// none keeps every star with a positive parallax.
func TestTheSharesAtTheirEdges(test *testing.T) {
	for _, each := range []struct {
		flags []string
		stars int
	}{{[]string{"--fastest", fmt.Sprint(allOfThem)}, sixKept},
		{[]string{"--fastest", fmt.Sprint(oneOfThem)}, oneStarRan},
		{[]string{"--fastest", fmt.Sprint(halfOfThem), "--quality",
			fmt.Sprint(anyParallax)}, fourStars}} {
		open, answers := answered(test)
		args := append([]string{"--gaia", tenProduct(test), "--sun",
			"frame"}, each.flags...)
		code, out, errs := orbitsRan(open, published{}, orbitsNumbers,
			args...)
		stars := starLines(out)
		if code != exitRan || len(stars) != each.stars {
			test.Errorf("%v: exit %d, %d stars: %s%s", each.flags, code,
				len(stars), out, errs)
		}
		if len(answers.batches) == firstBatch {
			test.Errorf("%v ran nothing", each.flags)
		}
	}
	open, _ := answered(test)
	_, out, _ := orbitsRan(open, published{}, orbitsNumbers, "--gaia",
		tenProduct(test), "--sun", "frame", "--quality",
		fmt.Sprint(anyParallax))
	rows := len(tenStars())
	kept := fmt.Sprintf("# %d of %d rows kept", sevenKept, rows)
	if !strings.Contains(out, kept) {
		test.Errorf("--quality none kept: %s", out)
	}
}

// the share is counted by nearest integer, at least one: of ten
// speeds, --fastest 0.7 selects seven, where a ceiling of the doubles'
// 3.0000000000000004 would have left four; a share under a tenth
// selects one; the whole selects ten.
func TestTheShareIsCountedByNearestInteger(test *testing.T) {
	var speeds []float32
	for each := range tenSpeeds {
		speeds = append(speeds, float32(each+oneStarRan))
	}
	for share, want := range map[float64]int{sevenOfTen: sevenChosen,
		oneOfThem: oneStarRan, allOfThem: tenSpeeds} {
		cut, kept := cutAt(speeds, share)
		if got := selected(speeds, cut); got != want || kept != tenSpeeds {
			test.Errorf("--fastest %g selects %d of %d at %g, want %d", share,
				got, kept, cut, want)
		}
	}
}

// a batch of one runs the selected stars one at a time, each batch
// full as it is made.
func TestABatchOfOneRunsTheStarsOneAtATime(test *testing.T) {
	open, made := running(test)
	answers := &answering{}
	made.EXPECT().Back(gomock.Len(oneAtATime), gomock.Any(),
		gomock.Any()).DoAndReturn(answers.back).Times(threeStars)
	code, out, errs := orbitsRan(open, published{}, orbitsNumbers, "--gaia",
		tenProduct(test), "--sun", "frame", "--fastest",
		fmt.Sprint(halfOfThem), "--batch", fmt.Sprint(oneAtATime))
	if code != exitRan || len(starLines(out)) != threeStars {
		test.Fatalf("exit %d: %s%s", code, out, errs)
	}
}

// a product with no star kept selects nothing and runs nothing: the
// header says so and no star is written.
func TestAProductWithNothingKeptRunsNothing(test *testing.T) {
	open, _ := running(test)
	unkept := []*fitstest.GaiaStar{star(starE, poorParallax, poorParallax, noMotion,
		poorRadial), star(starF, nearParallax, smallError, noMotion,
		math.NaN())}
	code, out, errs := orbitsRan(open, published{}, orbitsNumbers, "--gaia",
		productOf(test, unkept...), "--sun", "frame")
	if code != exitRan || len(starLines(out)) != noRows ||
		!strings.Contains(out, "# 0 of 2 rows kept, 0 at") {
		test.Fatalf("exit %d: %s%s", code, out, errs)
	}
}

// refusing is a solver that names as run away every star in a batch
// going faster along the rotation than the fastest radial velocity
// with the sun's motion added, and answers a batch with none.
type refusing struct {
	answers answering
}

// back refuses a batch with such a star, naming each, and answers the
// rest.
func (refuses *refusing) back(stars []orbit.State, span measure.Duration,
	steps int) ([]orbit.Pass, []orbit.State, error) {
	var away []int
	for which, star := range stars {
		if _, vy, _ := star.Numbers().Moving(); vy > runawaySpeed {
			away = append(away, which)
		}
	}
	if len(away) > noRows {
		return nil, nil, gpuorbit.NewRunaway(away)
	}
	return refuses.answers.back(stars, span, steps)
}

// the anti-test: a batch the solver refuses for a star that ran away
// is run again without that star, which is left out and named in the
// header, and the rest are written: the first star, the fastest, goes
// past the solver's bound with the sun's motion added. In batches of
// one, that star's batch runs away whole and is not run again.
func TestAStarThatRunsAwayIsLeftOutAndNamed(test *testing.T) {
	for _, each := range []struct {
		flags []string
		calls int
	}{{nil, twoCalls},
		{[]string{"--batch", fmt.Sprint(oneAtATime)}, threeStars}} {
		open, made := running(test)
		refuses := &refusing{}
		made.EXPECT().Back(gomock.Any(), gomock.Any(),
			gomock.Any()).DoAndReturn(refuses.back).Times(each.calls)
		args := append([]string{"--gaia", tenProduct(test), "--sun", "frame",
			"--fastest", fmt.Sprint(halfOfThem)}, each.flags...)
		code, out, errs := orbitsRan(open, published{}, orbitsNumbers,
			args...)
		if code != exitRan {
			test.Fatalf("%v: exit %d: %s", each.flags, code, errs)
		}
		stars := starLines(out)
		if len(stars) != twoStars || !strings.Contains(out,
			fmt.Sprintf("# star %d is left out: its orbit ran away", starA)) {
			test.Fatalf("%v: wrote %d stars: %s", each.flags, len(stars), out)
		}
	}
}

// the anti-test: a solver that refuses a batch for any other reason
// stops the run with its words and nothing written, whether the batch
// was the last or one filled on the way; and one that names a star the
// batch does not have stops it too.
func TestASolverThatRefusesOtherwiseStopsTheRun(test *testing.T) {
	gone := errors.New("the device is gone")
	for _, each := range []struct {
		refusal error
		flags   []string
		said    string
	}{{gone, nil, "the device is gone"},
		{gone, []string{"--batch", fmt.Sprint(oneAtATime)},
			"the device is gone"},
		{gpuorbit.NewRunaway([]int{noSuchStar}), nil,
			fmt.Sprintf("named star %d of a batch of %d", noSuchStar,
				threeStars)}} {
		open, made := running(test)
		made.EXPECT().Back(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			nil, nil, each.refusal)
		args := append([]string{"--gaia", tenProduct(test), "--sun", "frame",
			"--fastest", fmt.Sprint(halfOfThem)}, each.flags...)
		code, out, errs := orbitsRan(open, published{}, orbitsNumbers,
			args...)
		if code == exitRan || out != "" || !strings.Contains(errs, each.said) {
			test.Errorf("%v %v: exit %d, wrote %q, said %q", each.refusal,
				each.flags, code, out, errs)
		}
	}
}

// the anti-test: a star the frame refuses, one past light, is left out
// with the frame's words, and the rest run: seven kept of eleven, four
// selected at a half, one refused, three written.
func TestAStarTheFrameRefusesIsLeftOut(test *testing.T) {
	open, answers := answered(test)
	eleven := append(tenStars(), star(starG, nearParallax, smallError,
		noMotion, pastLight))
	code, out, errs := orbitsRan(open, published{}, orbitsNumbers, "--gaia",
		productOf(test, eleven...), "--sun", "frame", "--fastest",
		fmt.Sprint(halfOfThem))
	if code != exitRan {
		test.Fatalf("exit %d: %s", code, errs)
	}
	stars := starLines(out)
	if len(stars) != threeStars || len(answers.batches) != oneStarRan ||
		!strings.Contains(out, fmt.Sprintf("# star %d is left out:", starG)) {
		test.Fatalf("wrote %d stars: %s", len(stars), out)
	}
}

// refused is a call that cannot run: its arguments, and the words its
// refusal says.
type refused struct {
	args []string
	said string
}

// the anti-test: a call that cannot run is refused before anything is
// opened, saying why: without --gaia or --sun, with a --sun that is
// neither frame nor model, with a share of none or past the whole,
// with no steps, no batch or no span, with a quality below none, with
// a span or steps that are not numbers, with a product that is not
// there, or with numbers the verb cannot read or that lack one.
func TestAnOrbitsCallThatCannotRunIsRefused(test *testing.T) {
	open := NewMockopener(gomock.NewController(test))
	product := tenProduct(test)
	calls := map[string]refused{
		"no gaia":      {[]string{"--sun", "frame"}, "needs --gaia"},
		"no sun":       {[]string{"--gaia", product}, "needs --sun"},
		"another sun":  {[]string{"--gaia", product, "--sun", "mine"}, "needs --sun"},
		"no share":     {[]string{"--gaia", product, "--sun", "frame", "--fastest", fmt.Sprint(noShareFlag)}, "--fastest 0:"},
		"past whole":   {[]string{"--gaia", product, "--sun", "frame", "--fastest", fmt.Sprint(pastWhole)}, "--fastest 2:"},
		"no steps":     {[]string{"--gaia", product, "--sun", "frame", "--steps", fmt.Sprint(noSteps)}, "--steps 0,"},
		"no batch":     {[]string{"--gaia", product, "--sun", "frame", "--batch", fmt.Sprint(noBatch)}, "--batch 0:"},
		"no span":      {[]string{"--gaia", product, "--sun", "frame", "--span", fmt.Sprint(noSpan)}, "--span 0,"},
		"below none":   {[]string{"--gaia", product, "--sun", "frame", "--quality", fmt.Sprint(belowNone)}, "--quality -1:"},
		"not a number": {[]string{"--gaia", product, "--sun", "frame", "--span", "soon"}, "--span:"},
		"not a count":  {[]string{"--gaia", product, "--sun", "frame", "--steps", "many"}, "--steps:"},
		"not there":    {[]string{"--gaia", filepath.Join(test.TempDir(), "none"), "--sun", "model"}, "no such file"},
	}
	for why, each := range calls {
		if code, _, errs := orbitsRan(open, published{}, orbitsNumbers,
			each.args...); code == exitRan || !strings.Contains(errs, each.said) {
			test.Errorf("%s: exit %d, %s", why, code, errs)
		}
	}
	for why, each := range map[string]struct {
		raw  []byte
		said string
	}{"not yaml": {[]byte(notNumbers), orbitsFile},
		"no steps number": {withoutNumber("steps"), "steps"},
		"no share number": {withoutNumber("fastest-share"), "fastest-share"}} {
		if code, _, errs := orbitsRan(open, published{}, each.raw, "--gaia",
			product, "--sun", "frame"); code == exitRan ||
			!strings.Contains(errs, each.said) {
			test.Errorf("%s: exit %d, %s", why, code, errs)
		}
	}
}

// the anti-test: a product that is not the query's, a ledger that
// cannot give the galaxy or the frame, and a solver that cannot be
// opened, each stop the call with their error.
func TestWhatCannotBeOpenedStopsTheCall(test *testing.T) {
	controller := gomock.NewController(test)
	open, read := NewMockopener(controller), NewMockledger(controller)
	notTheQuerys := filepath.Join(test.TempDir(), "sizes.fits")
	file, err := os.Create(notTheQuerys)
	if err != nil {
		test.Fatal(err)
	}
	defer file.Close()
	sizes := []fitsio.Column{{Name: "size", Format: "E"}}
	if err := fitstest.Write(file, "sizes", sizes, &struct {
		Size float32 `fits:"size"`
	}{}); err != nil {
		test.Fatal(err)
	}
	product := tenProduct(test)
	noGalaxy, noFrame := errors.New("no galaxy"), errors.New("no frame")
	noGPU := errors.New("no gpu here")
	gomock.InOrder(
		read.EXPECT().Galaxy().Return(nil, noGalaxy),
		read.EXPECT().Galaxy().DoAndReturn(published{}.Galaxy),
		read.EXPECT().Frame().Return(nil, noFrame),
		read.EXPECT().Galaxy().DoAndReturn(published{}.Galaxy),
		read.EXPECT().Frame().DoAndReturn(published{}.Frame),
		read.EXPECT().Galaxy().DoAndReturn(published{}.Galaxy),
		read.EXPECT().Frame().DoAndReturn(published{}.Frame),
		open.EXPECT().Open(gomock.Any(), gomock.Any()).Return(nil, noGPU))
	for _, each := range []struct{ said, product string }{
		{"no galaxy", product}, {"no frame", product},
		{"no source_id", notTheQuerys}, {"no gpu here", product}} {
		if code, _, errs := orbitsRan(open, read, orbitsNumbers, "--gaia",
			each.product, "--sun", "frame"); code == exitRan ||
			!strings.Contains(errs, each.said) {
			test.Errorf("%s: exit %d, %s", each.said, code, errs)
		}
	}
}

// the model's circular speed at its own sun, 8 kpc, is the published
// 220 km/s.
func TestTheModelsCircularSpeedIsThePublished(test *testing.T) {
	galaxy, err := published{}.Galaxy()
	if err != nil {
		test.Fatal(err)
	}
	if got := circular(*galaxy, modelSun); !(math.Abs(got-modelCircular) <=
		circularRoom) {
		test.Errorf("the model circles at %g km/s at %g kpc, want %g", got,
			float64(modelSun), float64(modelCircular))
	}
}

// the sun is as far from the centre along the x axis, toward itself,
// as the ledger says, to the tilt's cosine.
func TestTheSunIsAsFarFromTheCentreAsTheLedgerSays(test *testing.T) {
	frame, err := published{}.Frame()
	if err != nil {
		test.Fatal(err)
	}
	place, _ := frame.Sun()
	want := research.Value("sgr-a-distance") / parsecsInKpc
	if got := sunward(place); !near(got, want, sunRoom) {
		test.Errorf("the sun is %g kpc sunward of the centre, want %g", got,
			want)
	}
}

// the model's sun: with --sun model every star's motion along the
// rotation is slower than with --sun frame by the seam, about 20 km/s,
// the published speed less the model's circular speed at the sun's
// distance and the sun's own motion, worked by hand; its motion toward
// the centre and toward the pole is the same either way.
func TestTheModelsSunSlowsEveryStarByTheSeam(test *testing.T) {
	var batches [][]orbit.State
	for _, sun := range []string{sunAsPublished, sunAsTheModel} {
		open, answers := answered(test)
		if code, _, errs := orbitsRan(open, published{}, orbitsNumbers,
			"--gaia", tenProduct(test), "--sun", sun, "--fastest",
			fmt.Sprint(halfOfThem)); code != exitRan {
			test.Fatalf("%s: exit %d: %s", sun, code, errs)
		}
		batches = append(batches, answers.batches[firstBatch])
	}
	asPublished, asModel := batches[firstBatch], batches[oneStarRan]
	for which := range asPublished {
		ux, published, uz := asPublished[which].Numbers().Moving()
		mx, model, mz := asModel[which].Numbers().Moving()
		if !(math.Abs(published-model-seamKmS) <= seamRoom) ||
			!(math.Abs(ux-mx) <= speedRoom) || !(math.Abs(uz-mz) <= speedRoom) {
			test.Errorf("star %d goes (%g %g %g) with the frame's sun and (%g "+
				"%g %g) with the model's, want %g apart along the rotation "+
				"and the same otherwise", which, ux, published, uz, mx, model,
				mz, seamKmS)
		}
	}
}

// absurd is the published model's shapes and shares with its sun
// circling past light, as a galaxy, or the test stops.
func absurd(test *testing.T) *orbit.MilkyWay {
	test.Helper()
	shaped := orbit.MWPotential2014()
	bulge, disc, halo := shaped.Shares()
	exponent, cutoff := shaped.BulgeShape()
	length, height := shaped.DiscShape()
	model := orbit.NewModel(modelSun, pastLight).WithBulge(bulge, exponent,
		cutoff).WithDisc(disc, length, height).WithHalo(halo,
		shaped.HaloShape())
	hole, err := grav.Point(measure.SolarMasses(someSuns))
	if err != nil {
		test.Fatal(err)
	}
	galaxy, err := orbit.NewMilkyWay(model, *hole)
	if err != nil {
		test.Fatal(err)
	}
	return galaxy
}

// the anti-test: a model whose sun would go past light is refused at
// --sun model, and a hole of less than no mass makes no galaxy.
func TestAModelThePhysicsCannotHoldIsRefused(test *testing.T) {
	controller := gomock.NewController(test)
	open, read := NewMockopener(controller), NewMockledger(controller)
	galaxy := absurd(test)
	read.EXPECT().Galaxy().Return(galaxy, nil)
	read.EXPECT().Frame().DoAndReturn(published{}.Frame)
	if code, _, errs := orbitsRan(open, read, orbitsNumbers, "--gaia",
		tenProduct(test), "--sun", "model"); code == exitRan ||
		!strings.Contains(errs, "the sun's motion") {
		test.Errorf("a model past light: exit %d, %s", code, errs)
	}
	if galaxy, err := milkyWay(noSuns); err == nil || galaxy != nil {
		test.Errorf("a hole of less than no mass gave %v, %v", galaxy, err)
	}
}

// the anti-test: a row the catalogue cannot give stops the run, in
// the pass over every row and in the pass over the selected.
func TestARowTheCatalogueCannotGiveStopsTheRun(test *testing.T) {
	catalogue := NewMockproduct(gomock.NewController(test))
	made := NewMocksolver(gomock.NewController(test))
	frame, err := galactocentric.Published()
	if err != nil {
		test.Fatal(err)
	}
	path := tenProduct(test)
	real, err := os.Open(path)
	if err != nil {
		test.Fatal(err)
	}
	defer real.Close()
	stars, err := gaia.OpenCatalogue(real)
	if err != nil {
		test.Fatal(err)
	}
	asked := orbitsAsked{share: halfOfThem, years: fewYears,
		steps: fewSteps, batch: twoRows}
	catalogue.EXPECT().Rows().Return(int64(twoRows)).AnyTimes()
	first, again := errors.New("the first row"), errors.New("the row again")
	gomock.InOrder(
		catalogue.EXPECT().Source(gomock.Any()).Return(nil, first),
		catalogue.EXPECT().Source(gomock.Any()).DoAndReturn(stars.Source),
		catalogue.EXPECT().Source(gomock.Any()).DoAndReturn(stars.Source),
		catalogue.EXPECT().Source(gomock.Any()).Return(nil, again))
	for _, said := range []string{"the first row", "the row again"} {
		var out bytes.Buffer
		err := asked.run(catalogue, frame, made, &out)
		if err == nil || !strings.Contains(err.Error(), said) {
			test.Errorf("%s: %v", said, err)
		}
	}
}

// failing is a writer that takes nothing.
type failing struct{}

// Write refuses every write.
func (failing) Write([]byte) (int, error) {
	return noRows, errors.New("nothing written")
}

// the anti-test: a writer that takes nothing stops the writing with
// its error.
func TestAWriterThatTakesNothingStopsTheWriting(test *testing.T) {
	if err := (orbitsAsked{}).write(failing{}, nil, nil); err == nil {
		test.Error("nothing written was taken as written")
	}
}

// results order by their nearest pass, the nearest first, and two at
// one distance keep their order.
func TestResultsOrderByTheirNearestPass(test *testing.T) {
	nearer, farther := result{near: nearA - nearStep}, result{near: nearA}
	if byNearest(nearer, farther) >= same || byNearest(farther, nearer) <=
		same || byNearest(nearer, nearer) != same {
		test.Errorf("the order is %d, %d, %d", byNearest(nearer, farther),
			byNearest(farther, nearer), byNearest(nearer, nearer))
	}
}
