package sky_test

import (
	"strings"
	"testing"

	"github.com/janearc/nilgai/sky"
)

// TestNewSkyRefusesWhatCannotBeDrawn checks that NewSky refuses a width or
// a pixel count that is not above zero.
func TestNewSkyRefusesWhatCannotBeDrawn(test *testing.T) {
	read := loadSky(test)
	for _, each := range read.BadSizes {
		made, err := sky.NewSky(read.centre(test), float64(each.Width),
			each.Pixels, nilgaiProfiles(test), readerProfile(test))
		if err == nil {
			test.Errorf("%s: a sky was made", each.Name)
		}
		if made != nil {
			test.Errorf("%s: a sky came back beside a refusal", each.Name)
		}
	}
}

// TestNewSkyDrawsSmallSizes is an anti-test. It checks that NewSky accepts
// sizes that are small but above zero.
func TestNewSkyDrawsSmallSizes(test *testing.T) {
	read := loadSky(test)
	for _, each := range read.GoodSizes {
		_, err := sky.NewSky(read.centre(test), float64(each.Width),
			each.Pixels, nilgaiProfiles(test), readerProfile(test))
		if err != nil {
			test.Errorf("%s: %v", each.Name, err)
		}
	}
}

// TestReadSkyRefusesBadStyles checks that ReadSky refuses, and says which
// fault it found:
//   - a style it cannot read
//   - a colour it cannot parse
//   - a profile with no contrast
//   - a contrast under the reader's floor
//   - a contrast that no ink can reach
func TestReadSkyRefusesBadStyles(test *testing.T) {
	read := loadSky(test)
	for _, each := range read.BadStyles {
		data := []byte(each.Raw)
		if each.Raw == "" {
			data = styleWith(test, each.Set)
		}
		made, err := sky.ReadSky(data, read.centre(test),
			float64(read.Width), read.Pixels, nilgaiProfiles(test),
			readerProfile(test))
		if err == nil {
			test.Errorf("%s: a sky was made", each.Name)
			continue
		}
		if made != nil {
			test.Errorf("%s: a sky came back beside a refusal", each.Name)
		}
		for _, word := range each.Want {
			if !strings.Contains(err.Error(), word) {
				test.Errorf("%s: %q does not say %q", each.Name, err,
					word)
			}
		}
	}
}

// TestReadSkyTakesTheStylesThatHoldTheFloor is an anti-test. It checks
// that ReadSky accepts the style nilgai ships, and a style whose contrast
// sits exactly at the reader's floor.
func TestReadSkyTakesTheStylesThatHoldTheFloor(test *testing.T) {
	read := loadSky(test)
	styles := map[string][]byte{
		"shipped":  shippedSky,
		"at floor": styleWith(test, read.AtTheFloor),
	}
	for name, data := range styles {
		_, err := sky.ReadSky(data, read.centre(test),
			float64(read.Width), read.Pixels, nilgaiProfiles(test),
			readerProfile(test))
		if err != nil {
			test.Errorf("%s: %v", name, err)
		}
	}
}

// TestReadSkyRefusesAMarkThatCannotReadFirst checks that NewSky refuses a
// mark that no lightness can bring into the reader's prose band. The
// fixture's band is one that no ground can give.
func TestReadSkyRefusesAMarkThatCannotReadFirst(test *testing.T) {
	read := loadSky(test)
	reader := readerProfile(test)
	unreadable := read.Reader.Unreadable
	reader.Contrast.Low = unreadable.Low
	reader.Contrast.High = unreadable.High
	reader.Contrast.Centre = unreadable.Centre
	_, err := sky.NewSky(read.centre(test), float64(read.Width),
		read.Pixels, nilgaiProfiles(test), reader)
	if err == nil || !strings.Contains(err.Error(), "mark") {
		test.Errorf("a mark that cannot read first: %v", err)
	}
}

// TestReadSkyHoldsTheReadersFloor is an anti-test. The floor comes from
// the reader, not from a number of nilgai's own, so a reader whose floor
// is above the style's lowest contrast makes NewSky refuse the style.
func TestReadSkyHoldsTheReadersFloor(test *testing.T) {
	read := loadSky(test)
	reader := readerProfile(test)
	reader.Contrast.ChipMin = read.Reader.Floor
	_, err := sky.NewSky(read.centre(test), float64(read.Width),
		read.Pixels, nilgaiProfiles(test), reader)
	if err == nil || !strings.Contains(err.Error(), "floor") {
		test.Errorf("a floor of %v against the shipped style: %v",
			read.Reader.Floor, err)
	}
}
