package sky_test

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/janearc/libreadme/colour"
	"github.com/janearc/libreadme/profile"
	"github.com/janearc/libtheme-css/spaces/radiation"
	"github.com/janearc/nilgai"
	"github.com/janearc/nilgai/band"
	"github.com/janearc/nilgai/sky"
)

// shippedSky holds data/sky.json as nilgai ships it.
//
//go:embed data/sky.json
var shippedSky []byte

// offset is a distance in pixels from a place in a picture, to the right
// and down. A negative distance is to the left and up.
type offset struct{ Right, Down float64 }

// panel holds a sky's width in arcseconds and its size in pixels.
type panel struct {
	Width  Num
	Pixels int
}

// skyFixture holds testdata/sky.json.
type skyFixture struct {
	Centre struct{ RA, Dec Num }
	Width  Num
	Pixels int
	Arm    int
	Slack  Num

	HueApart Num
	Axes     struct{ Major, Minor float64 }
	Reader   struct {
		Unreadable struct{ Low, High, Centre float64 }
		Floor      float64
	}
	Mark struct {
		East, North float64
		Left, Up    int
	}
	Strip    struct{ First, Second panel }
	BadSizes []struct {
		Name string
		panel
	}
	GoodSizes []struct {
		Name string
		panel
	}
	BadStyles []struct {
		Name string
		Raw  string
		Set  map[string]any
		Want []string
	}
	AtTheFloor map[string]any
	Ellipses   []struct {
		Name         string
		Angle        Num
		Major, Minor offset
	}
	Ground []struct {
		Name        string
		Right, Down int
		FromCorner  bool
	}
}

// loadSky reads the sky fixture.
func loadSky(test *testing.T) skyFixture {
	test.Helper()
	var read skyFixture
	load(test, "sky", &read)
	return read
}

// shipped returns data/sky.json as a table of its keys.
func shipped(test *testing.T) map[string]any {
	test.Helper()
	var style map[string]any
	if err := json.Unmarshal(shippedSky, &style); err != nil {
		test.Fatal(err)
	}
	return style
}

// styleWith returns the shipped style with the keys in set replaced.
func styleWith(test *testing.T, set map[string]any) []byte {
	test.Helper()
	style := shipped(test)
	for key, value := range set {
		style[key] = value
	}
	data, err := json.Marshal(style)
	if err != nil {
		test.Fatal(err)
	}
	return data
}

// styleColour returns the colour that the shipped style gives under key.
func styleColour(test *testing.T, key string) colour.RGB {
	test.Helper()
	text, isText := shipped(test)[key].(string)
	if !isText {
		test.Fatalf("data/sky.json has no colour %s", key)
	}
	parsed, err := colour.Parse(text)
	if err != nil {
		test.Fatal(err)
	}
	return parsed.RGB
}

// styleGap returns the shipped style's gap between panels, in pixels.
func styleGap(test *testing.T) int {
	test.Helper()
	gap, isNumber := shipped(test)["gap"].(float64)
	if !isNumber {
		test.Fatal("data/sky.json has no gap")
	}
	return int(gap)
}

// skyAims returns the contrast that data/sky.json sets for each profile,
// and how near an ink must land to it, as a share of the contrast.
func skyAims(test *testing.T) (aims map[string]float64, within float64) {
	test.Helper()
	var style struct {
		Contrast map[string]float64
		Within   float64
	}
	if err := json.Unmarshal(shippedSky, &style); err != nil {
		test.Fatal(err)
	}
	return style.Contrast, style.Within
}

// readerProfile returns libreadme's current profile, which nilgai draws
// for.
func readerProfile(test *testing.T) profile.Profile {
	test.Helper()
	reader, err := profile.Current()
	if err != nil {
		test.Fatal(err)
	}
	return reader
}

// nilgaiProfiles returns nilgai's own profiles. It fails the test if they
// do not read.
func nilgaiProfiles(test *testing.T) []radiation.Profile {
	test.Helper()
	profiles, err := band.Profiles()
	if err != nil {
		test.Fatal(err)
	}
	return profiles
}

// centre returns the fixture's centre as a position.
func (read skyFixture) centre(test *testing.T) nilgai.Position {
	test.Helper()
	centre, err := nilgai.NewPosition(float64(read.Centre.RA),
		float64(read.Centre.Dec))
	if err != nil {
		test.Fatal(err)
	}
	return centre
}

// sky returns the fixture's sky, built on nilgai's profiles and the
// current reader. sky fails the test if the sky is refused.
func (read skyFixture) sky(test *testing.T) *sky.Sky {
	test.Helper()
	return read.skyOf(test, read.Width, read.Pixels)
}

// skyOf returns a sky of the width and size given, at the fixture's
// centre.
func (read skyFixture) skyOf(test *testing.T, width Num,
	pixels int) *sky.Sky {
	test.Helper()
	sky, err := sky.NewSky(read.centre(test), float64(width), pixels,
		nilgaiProfiles(test), readerProfile(test))
	if err != nil {
		test.Fatal(err)
	}
	return sky
}
