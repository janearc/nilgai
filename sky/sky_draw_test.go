package sky_test

import (
	"cmp"
	"image"
	"math"
	"slices"
	"testing"

	"github.com/janearc/libreadme/colour"
	"github.com/janearc/libtheme-css/spaces/radiation"
	"github.com/janearc/nilgai"
	"github.com/janearc/nilgai/band"
	"github.com/janearc/nilgai/sky"
)

// middle returns the pixel at the middle of the fixture's sky, on each
// axis.
func (read skyFixture) middle() int { return read.Pixels / halves }

// farOff returns a place that no sky of the fixture's size reaches: a
// degree east of the centre.
func (read skyFixture) farOff(test *testing.T) nilgai.Position {
	test.Helper()
	far, err := nilgai.NewPosition(float64(read.Centre.RA)+one,
		float64(read.Centre.Dec))
	if err != nil {
		test.Fatal(err)
	}
	return far
}

// sourceAt returns a source of the catalogue at the position. Its
// one-sigma ellipse has the semi-axes given in arcseconds and the angle
// given in degrees.
func sourceAt(test *testing.T, catalogue nilgai.Catalogue,
	at nilgai.Position, major, minor, angle float64) nilgai.Source {
	test.Helper()
	ellipse, err := nilgai.NewErrorEllipse(major, minor, angle,
		nilgai.OneSigma())
	if err != nil {
		test.Fatal(err)
	}
	first, last, err := band.FilterBands("W1")
	if err != nil {
		test.Fatal(err)
	}
	flux, err := nilgai.NewDetection("W1", first, last,
		nilgai.EnergyFlux(one), nilgai.EnergyFlux(one))
	if err != nil {
		test.Fatal(err)
	}
	var epoch nilgai.When
	source, err := nilgai.NewSource(catalogue, "a source", at, ellipse,
		epoch, []nilgai.Measurement{*flux})
	if err != nil {
		test.Fatal(err)
	}
	return *source
}

// bigSource returns a source of the catalogue at the fixture's centre,
// with the fixture's axes and the angle given.
func (read skyFixture) bigSource(test *testing.T,
	catalogue nilgai.Catalogue, angle float64) nilgai.Source {
	test.Helper()
	return sourceAt(test, catalogue, read.centre(test), read.Axes.Major,
		read.Axes.Minor, angle)
}

// rgbAt returns the pixel's colour as libreadme has it, 0 to 255.
func rgbAt(picture *image.RGBA, x, y int) colour.RGB {
	each := picture.RGBAAt(x, y)
	return colour.RGB{R: float64(each.R), G: float64(each.G),
		B: float64(each.B)}
}

// catalogues maps every catalogue that has an ink to the profile in
// band/data/profiles.json that it takes its hue from.
var catalogues = map[nilgai.Catalogue]string{nilgai.Fermi: "fermi",
	nilgai.Chandra: "chandra", nilgai.AllWISE: "neowise",
	nilgai.NEOWISE: "neowise", nilgai.MeerKAT: "meerkat"}

// distinct lists one catalogue for each instrument. AllWISE and NEOWISE
// share an instrument, so only AllWISE is listed.
var distinct = []nilgai.Catalogue{nilgai.Fermi, nilgai.Chandra,
	nilgai.AllWISE, nilgai.MeerKAT}

// inkOf returns the colour that a catalogue's sources are drawn in. It
// reads the dot at a source's position, with the mark out of the picture.
func inkOf(test *testing.T, read skyFixture,
	catalogue nilgai.Catalogue) colour.RGB {
	test.Helper()
	source := sourceAt(test, catalogue, read.centre(test), one, one, none)
	picture := read.sky(test).Draw([]nilgai.Source{source},
		read.farOff(test))
	return rgbAt(picture, read.middle(), read.middle())
}

// markInk returns the colour of the mark. It reads the middle of a sky
// whose mark is there.
func markInk(test *testing.T, read skyFixture) colour.RGB {
	test.Helper()
	picture := read.sky(test).Draw(nil, read.centre(test))
	return rgbAt(picture, read.middle(), read.middle())
}

// TestInksHoldTheFloorAndTheirAim checks that each instrument's drawn ink
// holds the reader's floor against the ground, and lands near the contrast
// that data/sky.json aims for. It measures the pixels drawn, not the
// colour the solver meant.
func TestInksHoldTheFloorAndTheirAim(test *testing.T) {
	read := loadSky(test)
	aims, within := skyAims(test)
	floor := readerProfile(test).Contrast.ChipMin
	ground := styleColour(test, "ground")
	for catalogue, name := range catalogues {
		got := colour.Contrast(inkOf(test, read, catalogue), ground)
		aim := aims[name]
		near := aim * (within + float64(read.Slack))
		if got < floor || math.Abs(got-aim) > near {
			test.Errorf("%s ink is %.3f:1 on the ground; floor %v, "+
				"aim %v within %v", catalogue, got, floor, aim, near)
		}
	}
}

// TestInksAreNotTheGroundOrTheMark is an anti-test. No ink may be the
// ground colour or the mark colour, and none may be in the reader's prose
// band, which belongs to the mark alone.
func TestInksAreNotTheGroundOrTheMark(test *testing.T) {
	read := loadSky(test)
	reader := readerProfile(test)
	ground := styleColour(test, "ground")
	mark := markInk(test, read)
	for catalogue := range catalogues {
		ink := inkOf(test, read, catalogue)
		contrast := colour.Contrast(ink, ground)
		if ink == ground || ink == mark || contrast >= reader.Contrast.Low {
			test.Errorf("%s ink %v is the ground, the mark or in the "+
				"prose band at %.3f:1", catalogue, ink, contrast)
		}
	}
}

// reading pairs a catalogue with a contrast, either measured or aimed at.
type reading struct {
	catalogue nilgai.Catalogue
	contrast  float64
}

// highestFirst orders readings from the most contrast to the least, as a
// comparison function for slices.SortFunc.
func highestFirst(first, second reading) int {
	return cmp.Compare(second.contrast, first.contrast)
}

// catalogueOf returns the catalogue that a reading is of.
func catalogueOf(each reading) nilgai.Catalogue { return each.catalogue }

// TestInksKeepTheOrderSkyJSONGives checks that the inks fall in the order
// of the contrasts in data/sky.json: darkest at the shortest wavelengths.
func TestInksKeepTheOrderSkyJSONGives(test *testing.T) {
	read := loadSky(test)
	aims, _ := skyAims(test)
	ground := styleColour(test, "ground")
	var measured, aimed []reading
	for _, catalogue := range distinct {
		ink := inkOf(test, read, catalogue)
		measured = append(measured, reading{catalogue,
			colour.Contrast(ink, ground)})
		aimed = append(aimed, reading{catalogue, aims[catalogues[catalogue]]})
	}
	slices.SortFunc(measured, highestFirst)
	slices.SortFunc(aimed, highestFirst)
	got, want := make([]nilgai.Catalogue, none), make([]nilgai.Catalogue, none)
	for place := range measured {
		got = append(got, catalogueOf(measured[place]))
		want = append(want, catalogueOf(aimed[place]))
	}
	if !slices.Equal(got, want) {
		test.Errorf("inks run %v by contrast; sky.json runs %v", got, want)
	}
}

// TestOneInstrumentIsOneInk is an anti-test. One instrument has the same
// ink in either of its catalogues, and different instruments have
// different inks.
func TestOneInstrumentIsOneInk(test *testing.T) {
	read := loadSky(test)
	if inkOf(test, read, nilgai.AllWISE) !=
		inkOf(test, read, nilgai.NEOWISE) {
		test.Error("AllWISE and Neowise are one instrument and two inks")
	}
	seen := map[colour.RGB]nilgai.Catalogue{}
	for _, catalogue := range distinct {
		ink := inkOf(test, read, catalogue)
		if other, shared := seen[ink]; shared {
			test.Errorf("%s and %s share an ink", catalogue, other)
		}
		seen[ink] = catalogue
	}
}

// TestMarkIsInTheProseBand checks that the mark's contrast against the
// ground is in the reader's prose band, so the mark reads first, and
// that it is higher than any instrument's.
func TestMarkIsInTheProseBand(test *testing.T) {
	read := loadSky(test)
	reader := readerProfile(test)
	ground := styleColour(test, "ground")
	mark := colour.Contrast(markInk(test, read), ground)
	if mark < reader.Contrast.Low || mark > reader.Contrast.High {
		test.Errorf("the mark is %.3f:1; the prose band is %v to %v",
			mark, reader.Contrast.Low, reader.Contrast.High)
	}
	for catalogue := range catalogues {
		ink := colour.Contrast(inkOf(test, read, catalogue), ground)
		if ink >= mark {
			test.Errorf("%s ink at %.3f:1 outdoes the mark's %.3f:1",
				catalogue, ink, mark)
		}
	}
}

// inkedNear says whether any pixel within one pixel of the point has the
// colour. A line is two pixels wide and a point is rounded to a pixel, so
// the test allows that much slack.
func inkedNear(picture *image.RGBA, x, y float64, ink colour.RGB) bool {
	left, top := int(math.Round(x)), int(math.Round(y))
	for across := -one; across <= one; across++ {
		for down := -one; down <= one; down++ {
			if rgbAt(picture, left+across, top+down) == ink {
				return true
			}
		}
	}
	return false
}

// end is one end of an ellipse's axes, and the place where the fixture
// says that end falls.
type end struct {
	axis string
	at   offset
}

// TestDrawPutsInkOnTheOneSigmaEllipse checks that Draw inks a source's
// one-sigma ellipse at the ends of both its axes, whichever way it points.
// TestDrawPutsInkOnTheOneSigmaEllipse lays the major axis along the
// position angle, measured east of north, with east to the left.
func TestDrawPutsInkOnTheOneSigmaEllipse(test *testing.T) {
	read := loadSky(test)
	sky := read.sky(test)
	middle := float64(read.middle())
	ink := inkOf(test, read, nilgai.Chandra)
	for _, each := range read.Ellipses {
		source := read.bigSource(test, nilgai.Chandra, float64(each.Angle))
		picture := sky.Draw([]nilgai.Source{source}, read.farOff(test))
		for _, tip := range []end{{"major", each.Major},
			{"minor", each.Minor}} {
			x, y := middle+tip.at.Right, middle+tip.at.Down
			if !inkedNear(picture, x, y, ink) {
				test.Errorf("%s: no ink at the %s axis's end, %v from "+
					"the middle", each.Name, tip.axis, tip.at)
			}
		}
	}
}

// TestDrawLeavesGroundAwayFromTheEllipse is an anti-test. The ellipse is
// an outline and no larger than one sigma, so these places are ground:
//   - the middle of the major axis
//   - a point inside the minor axis
//   - the place where a 95% radius would end
//   - a far corner
func TestDrawLeavesGroundAwayFromTheEllipse(test *testing.T) {
	read := loadSky(test)
	middle := read.middle()
	ground := styleColour(test, "ground")
	source := read.bigSource(test, nilgai.Chandra, none)
	picture := read.sky(test).Draw([]nilgai.Source{source},
		read.farOff(test))
	for _, each := range read.Ground {
		x, y := middle+each.Right, middle+each.Down
		if each.FromCorner {
			x, y = each.Right, each.Down
		}
		if rgbAt(picture, x, y) != ground {
			test.Errorf("%s (%d, %d) is not the ground", each.Name, x, y)
		}
	}
}

// TestDrawPlacesTheMarkEastLeftAndNorthUp checks that a mark east and
// north of the centre is drawn left and up. Its cross is centred on the
// mark, and each arm is the share of the picture's width that
// data/sky.json sets.
func TestDrawPlacesTheMarkEastLeftAndNorthUp(test *testing.T) {
	read := loadSky(test)
	sky := read.sky(test)
	middle := read.middle()
	toward := float64(read.Centre.Dec) * math.Pi / degreesInHalfACircle
	east := read.Mark.East / arcsecondsPerDegree / math.Cos(toward)
	north := read.Mark.North / arcsecondsPerDegree
	at, err := nilgai.NewPosition(float64(read.Centre.RA)+east,
		float64(read.Centre.Dec)+north)
	if err != nil {
		test.Fatal(err)
	}
	mark := markInk(test, read)
	ground := styleColour(test, "ground")
	picture := sky.Draw(nil, at)
	left, up := middle-read.Mark.Left, middle-read.Mark.Up
	arm := read.Arm
	for _, tip := range []image.Point{{X: left, Y: up},
		{X: left + arm, Y: up}, {X: left - arm, Y: up},
		{X: left, Y: up + arm}, {X: left, Y: up - arm}} {
		if rgbAt(picture, tip.X, tip.Y) != mark {
			test.Errorf("no mark at %v; the cross is centred on (%d, %d) "+
				"with arms of %d", tip, left, up, arm)
		}
	}
	for _, off := range []image.Point{{X: middle, Y: middle},
		{X: left + arm + arm, Y: up}, {X: left + arm, Y: up + arm}} {
		if rgbAt(picture, off.X, off.Y) != ground {
			test.Errorf("%v is not the ground", off)
		}
	}
}

// TestDrawLeavesOutWhatIsOutsideOrHasNoInk checks that Draw leaves out a
// source outside the square and a source of a catalogue with no ink. The
// picture is all ground.
func TestDrawLeavesOutWhatIsOutsideOrHasNoInk(test *testing.T) {
	read := loadSky(test)
	ground := styleColour(test, "ground")
	outside := sourceAt(test, nilgai.Chandra, read.farOff(test),
		read.Axes.Major, read.Axes.Minor, none)
	picture := read.sky(test).Draw([]nilgai.Source{outside},
		read.farOff(test))
	if marked := notGround(picture, ground); marked != none {
		test.Errorf("%d pixels inked for a source outside the sky", marked)
	}
	without := withoutProfile(nilgaiProfiles(test), "chandra")
	sky, err := sky.NewSky(read.centre(test), float64(read.Width),
		read.Pixels, without, readerProfile(test))
	if err != nil {
		test.Fatal(err)
	}
	here := read.bigSource(test, nilgai.Chandra, none)
	picture = sky.Draw([]nilgai.Source{here}, read.farOff(test))
	if marked := notGround(picture, ground); marked != none {
		test.Errorf("%d pixels inked for a catalogue without an ink",
			marked)
	}
}

// TestDrawClipsASourceOnTheEdge checks that a source on the edge of the
// square is drawn in part: the pixels inside are inked and the rest are
// clipped. The fixture's width is 12 arcsec, so a source 6 arcsec east,
// half a width, is on the left edge.
func TestDrawClipsASourceOnTheEdge(test *testing.T) {
	read := loadSky(test)
	toward := float64(read.Centre.Dec) * math.Pi / degreesInHalfACircle
	halfWidth := float64(read.Width) / halves
	east := halfWidth / arcsecondsPerDegree / math.Cos(toward)
	edge, err := nilgai.NewPosition(float64(read.Centre.RA)+east,
		float64(read.Centre.Dec))
	if err != nil {
		test.Fatal(err)
	}
	source := sourceAt(test, nilgai.Chandra, edge, one, one, none)
	picture := read.sky(test).Draw([]nilgai.Source{source},
		read.farOff(test))
	ground := styleColour(test, "ground")
	if rgbAt(picture, none, read.middle()) == ground ||
		rgbAt(picture, read.Pixels-one, read.middle()) != ground {
		test.Error("a source on the left edge is not drawn there, or is " +
			"drawn on the right")
	}
}

// TestDrawDrawsWhatIsInsideAndHasAnInk is an anti-test. The same source
// that TestDrawLeavesOutWhatIsOutsideOrHasNoInk leaves out is drawn in a
// sky that has its profile.
func TestDrawDrawsWhatIsInsideAndHasAnInk(test *testing.T) {
	read := loadSky(test)
	ground := styleColour(test, "ground")
	here := read.bigSource(test, nilgai.Chandra, none)
	picture := read.sky(test).Draw([]nilgai.Source{here},
		read.farOff(test))
	if marked := notGround(picture, ground); marked == none {
		test.Error("a source in the middle of its sky is not drawn")
	}
}

// notGround counts the pixels that are not the ground.
func notGround(picture *image.RGBA, ground colour.RGB) int {
	marked := none
	bounds := picture.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if rgbAt(picture, x, y) != ground {
				marked++
			}
		}
	}
	return marked
}

// named is a profile name, with a method to test a profile against it.
type named string

// is says whether the profile has this name.
func (name named) is(each radiation.Profile) bool {
	return each.Name == string(name)
}

// withoutProfile returns the profiles minus the one called name.
func withoutProfile(profiles []radiation.Profile,
	name string) []radiation.Profile {
	return slices.DeleteFunc(slices.Clone(profiles), named(name).is)
}

// TestStripSizesAndGaps checks that a strip is as wide as its skies and
// the gaps between them, and as tall as the tallest sky. The gaps, and the
// space below a shorter sky, are the between colour. Each sky is drawn
// from the same sources.
func TestStripSizesAndGaps(test *testing.T) {
	read := loadSky(test)
	first, second := read.Strip.First, read.Strip.Second
	short := read.skyOf(test, first.Width, first.Pixels)
	tall := read.skyOf(test, second.Width, second.Pixels)
	between := styleColour(test, "between")
	ground := styleColour(test, "ground")
	gap, height := styleGap(test), second.Pixels
	here := sourceAt(test, nilgai.Chandra, read.centre(test), one, one, none)
	strip := sky.Strip([]*sky.Sky{short, tall},
		[]nilgai.Source{here}, read.farOff(test))
	wide := first.Pixels + gap + second.Pixels
	want := image.Rect(none, none, wide, height)
	if strip.Bounds() != want {
		test.Fatalf("strip is %v, want %v", strip.Bounds(), want)
	}
	below := first.Pixels + (height-first.Pixels)/halves
	gaps := []image.Point{{X: first.Pixels, Y: none},
		{X: first.Pixels + gap - one, Y: height - one},
		{X: first.Pixels / halves, Y: below}}
	for _, at := range gaps {
		if rgbAt(strip, at.X, at.Y) != between {
			test.Errorf("%v is %v, not the between colour", at,
				rgbAt(strip, at.X, at.Y))
		}
	}
	grounds := []image.Point{{X: none, Y: none},
		{X: first.Pixels + gap, Y: none},
		{X: wide - one, Y: height - one}}
	for _, at := range grounds {
		if rgbAt(strip, at.X, at.Y) != ground {
			test.Errorf("%v is not the ground", at)
		}
	}
	firstMiddle := first.Pixels / halves
	secondMiddle := first.Pixels + gap + second.Pixels/halves
	inked := []image.Point{{X: firstMiddle, Y: firstMiddle},
		{X: secondMiddle, Y: second.Pixels / halves}}
	for _, at := range inked {
		if rgbAt(strip, at.X, at.Y) == ground {
			test.Errorf("no source drawn at %v", at)
		}
	}
}

// TestStripIsAsTallAsTheTallestInAnyPlace checks that a strip is as tall
// as its tallest sky wherever that sky sits: first, in the middle or last.
func TestStripIsAsTallAsTheTallestInAnyPlace(test *testing.T) {
	read := loadSky(test)
	first, second := read.Strip.First, read.Strip.Second
	short := read.skyOf(test, first.Width, first.Pixels)
	tall := read.skyOf(test, second.Width, second.Pixels)
	orders := [][]*sky.Sky{{tall, short}, {short, tall, short},
		{short, tall}}
	for _, skies := range orders {
		strip := sky.Strip(skies, nil, read.farOff(test))
		if strip.Bounds().Dy() != second.Pixels {
			test.Errorf("%d skies are %d high, want %d", len(skies),
				strip.Bounds().Dy(), second.Pixels)
		}
	}
}

// TestStripOfTheShortSkyIsShort is an anti-test. A strip of short skies
// is only as high as they are, not as high as a tall sky it does not hold.
func TestStripOfTheShortSkyIsShort(test *testing.T) {
	read := loadSky(test)
	first := read.Strip.First
	short := read.skyOf(test, first.Width, first.Pixels)
	strip := sky.Strip([]*sky.Sky{short, short}, nil, read.farOff(test))
	if strip.Bounds().Dy() != first.Pixels {
		test.Errorf("two short skies are %d high, want %d",
			strip.Bounds().Dy(), first.Pixels)
	}
}

// TestStripOfOneSkyHasNoGap is an anti-test. A strip of one sky has no
// gap and is the size of that sky.
func TestStripOfOneSkyHasNoGap(test *testing.T) {
	read := loadSky(test)
	strip := sky.Strip([]*sky.Sky{read.sky(test)}, nil,
		read.farOff(test))
	want := image.Rect(none, none, read.Pixels, read.Pixels)
	if strip.Bounds() != want {
		test.Errorf("a strip of one sky is %v, want %v", strip.Bounds(),
			want)
	}
}

// TestStripOfNoSkiesIsEmpty checks that a strip of no skies is empty.
func TestStripOfNoSkiesIsEmpty(test *testing.T) {
	read := loadSky(test)
	strip := sky.Strip(nil, nil, read.farOff(test))
	if !strip.Bounds().Empty() {
		test.Errorf("a strip of no skies is %v", strip.Bounds())
	}
}

// TestStripOfSkiesIsNotEmpty is an anti-test. A strip of two skies is not
// empty.
func TestStripOfSkiesIsNotEmpty(test *testing.T) {
	read := loadSky(test)
	skies := []*sky.Sky{read.sky(test), read.sky(test)}
	if sky.Strip(skies, nil, read.farOff(test)).Bounds().Empty() {
		test.Error("a strip of two skies is empty")
	}
}
