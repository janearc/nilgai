package sky_test

import (
	"math"
	"testing"

	"github.com/janearc/libreadme/colour"
	"github.com/janearc/libtheme-css/spaces/radiation"
	"github.com/janearc/libtheme-css/spaces/srgb"
	"github.com/janearc/nilgai"
)

// hueAt returns the hue, in degrees, of the colour that a profile shows at
// a wavelength in nm, as libtheme gives it.
func hueAt(each radiation.Profile, nm float64) float64 {
	lamps, _ := srgb.FromSwatch(each.Swatch(nm))
	red, green, blue := lamps.Bytes()
	hue, _, _ := colour.HSL(colour.RGB{R: float64(red), G: float64(green),
		B: float64(blue)})
	return hue
}

// hueOf returns the hue of a drawn ink.
func hueOf(ink colour.RGB) float64 {
	hue, _, _ := colour.HSL(ink)
	return hue
}

// hueApart returns how far apart two hues are, measured the short way
// round the colour wheel.
func hueApart(first, second float64) float64 {
	return turnsApart(first, second)
}

// profileOf returns the nilgai profile called name.
func profileOf(test *testing.T, name string) radiation.Profile {
	test.Helper()
	for _, each := range nilgaiProfiles(test) {
		if each.Name == name {
			return each
		}
	}
	test.Fatalf("no profile %s", name)
	return radiation.Profile{}
}

// TestInksKeepTheHueOfTheMiddleOfTheBand checks that each ink keeps the
// hue its profile shows at the middle of its band. The middle is the
// geometric mean of the band's ends, which is even in the logarithm, as
// libtheme spreads its ranges. Only the ink's lightness moves.
func TestInksKeepTheHueOfTheMiddleOfTheBand(test *testing.T) {
	read := loadSky(test)
	for _, catalogue := range distinct {
		each := profileOf(test, catalogues[catalogue])
		middle := math.Sqrt(each.Source.Lo * each.Source.Hi)
		ink := hueOf(inkOf(test, read, catalogue))
		got := hueApart(ink, hueAt(each, middle))
		test.Logf("%s: %.2f degrees from the middle's hue", catalogue, got)
		if got > float64(read.HueApart) {
			test.Errorf("%s ink is %v degrees of hue from the middle of "+
				"its band; at most %v", catalogue, got, read.HueApart)
		}
	}
}

// TestInksAreNotTheHueOfTheArithmeticMean is an anti-test. The middle of a
// band is not its arithmetic mean. For a band as wide as Fermi's or
// MeerKAT's the two show different hues, and the ink is nearer the
// geometric one. The test checks Fermi.
func TestInksAreNotTheHueOfTheArithmeticMean(test *testing.T) {
	read := loadSky(test)
	each := profileOf(test, catalogues[nilgai.Fermi])
	middle := (each.Source.Lo + each.Source.Hi) / halves
	ink := hueOf(inkOf(test, read, nilgai.Fermi))
	if hueApart(ink, hueAt(each, middle)) <= float64(read.HueApart) {
		test.Errorf("the arithmetic mean's hue %v is not apart from the "+
			"ink's %v by more than %v", hueAt(each, middle), ink,
			read.HueApart)
	}
}
