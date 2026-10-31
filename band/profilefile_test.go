package band_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/spaces/radiation"
	"github.com/janearc/nilgai/band"
)

// goodEntry is an entry that is accepted: X-rays of 1 to 2 keV, shown in
// blue.
func goodEntry() band.ProfileEntry {
	return band.ProfileEntry{
		Name:  "x",
		From:  band.Quantity{Value: 1, Unit: "keV"},
		To:    band.Quantity{Value: 2, Unit: "keV"},
		Shown: radiation.Range{Lo: 470, Hi: 490},
	}
}

// TestAnEntryGivesItsProfile: goodEntry gives a profile whose source is
// its band, short end first, and whose render is its shown range.
func TestAnEntryGivesItsProfile(test *testing.T) {
	entry := goodEntry()
	entry.From, entry.To = entry.To, entry.From
	got, err := entry.Profile()
	if err != nil {
		test.Fatal(err)
	}
	one, _ := band.Quantity{Value: 1, Unit: "keV"}.Nanometres()
	two, _ := band.Quantity{Value: 2, Unit: "keV"}.Nanometres()
	if got.Name != "x" || got.Source.Lo != two || got.Source.Hi != one ||
		got.Render != entry.Shown {
		test.Errorf("profile %+v", got)
	}
}

// TestAnEntryMayBeShownInAllOfVisibleLight: an entry shown out to either
// end of visible light is not refused.
func TestAnEntryMayBeShownInAllOfVisibleLight(test *testing.T) {
	visible := radiation.Visible
	for _, shown := range []radiation.Range{visible,
		{Lo: visible.Lo, Hi: 490}, {Lo: 470, Hi: visible.Hi}} {
		entry := goodEntry()
		entry.Shown = shown
		if _, err := entry.Profile(); err != nil {
			test.Error(err)
		}
	}
}

// TestAnEntryIsRefusedForEachReason holds the anti-tests: each entry is
// refused with nil and with the words that say why.
func TestAnEntryIsRefusedForEachReason(test *testing.T) {
	visible := radiation.Visible
	tests := []struct {
		name   string
		change func(*band.ProfileEntry)
		says   []string
	}{
		{"no name", func(each *band.ProfileEntry) { each.Name = "" },
			[]string{"no name"}},
		{"a from end it cannot read", func(each *band.ProfileEntry) {
			each.From.Unit = "furlong"
		}, []string{"from end", `"furlong" is not a unit`}},
		{"a to end it cannot read", func(each *band.ProfileEntry) {
			each.To.Value = 0
		}, []string{"to end", "must be finite and above zero"}},
		{"a band with no width", func(each *band.ProfileEntry) {
			each.To = band.Quantity{Value: 1000, Unit: "eV"}
		}, []string{"no width"}},
		{"shown backwards", func(each *band.ProfileEntry) {
			each.Shown = radiation.Range{Lo: 490, Hi: 470}
		}, []string{"shown from 490 to 470 nm runs backwards"}},
		{"shown with no width", func(each *band.ProfileEntry) {
			each.Shown = radiation.Range{Lo: 480, Hi: 480}
		}, []string{"shown from 480 to 480 nm runs backwards"}},
		{"shown below visible light", func(each *band.ProfileEntry) {
			each.Shown = radiation.Range{Lo: visible.Lo - 1, Hi: 490}
		}, []string{"leaves visible light"}},
		{"shown above visible light", func(each *band.ProfileEntry) {
			each.Shown = radiation.Range{Lo: 470, Hi: visible.Hi + 1}
		}, []string{"leaves visible light"}},
	}
	for _, each := range tests {
		entry := goodEntry()
		each.change(&entry)
		got, err := entry.Profile()
		if err == nil || got != nil {
			test.Errorf("%s: %v and %v", each.name, got, err)
			continue
		}
		for _, says := range each.says {
			if !strings.Contains(err.Error(), says) {
				test.Errorf("%s: %q does not say %q", each.name, err,
					says)
			}
		}
	}
}

// TestLeavingVisibleLightSaysWhereItIs: refusing a shown range outside
// visible light says where visible light is.
func TestLeavingVisibleLightSaysWhereItIs(test *testing.T) {
	entry := goodEntry()
	entry.Shown = radiation.Range{Lo: 100, Hi: 490}
	_, err := entry.Profile()
	visible := radiation.Visible
	want := fmt.Sprintf("shown from 100 to 490 nm leaves visible "+
		"light, %v to %v nm", visible.Lo, visible.Hi)
	if err == nil || err.Error() != want {
		test.Errorf("%v, want %q", err, want)
	}
}

// goodFile is a file of the good entry under two names.
func goodFile() *band.ProfileFile {
	other := goodEntry()
	other.Name = "y"
	return &band.ProfileFile{Entries: []band.ProfileEntry{goodEntry(),
		other}}
}

// TestAFileGivesItsProfilesInOrder: ProfileFile.Profiles keeps the file's
// order.
func TestAFileGivesItsProfilesInOrder(test *testing.T) {
	got, err := goodFile().Profiles()
	if err != nil || len(got) != 2 || got[none].Name != "x" ||
		got[one].Name != "y" {
		test.Errorf("%v and %v", got, err)
	}
}

// TestAFileRefusesANameUsedTwice is the anti-test: a name used twice is
// refused, and gives no profiles.
func TestAFileRefusesANameUsedTwice(test *testing.T) {
	file := goodFile()
	file.Entries[one].Name = "x"
	got, err := file.Profiles()
	if err == nil || got != nil ||
		!strings.Contains(err.Error(), `profile "x" is named twice`) {
		test.Errorf("%v and %v", got, err)
	}
}

// TestAFileRefusesEverythingTogether is the anti-test: every refusal in a
// file comes back together, each naming its profile, and no profiles do.
func TestAFileRefusesEverythingTogether(test *testing.T) {
	file := goodFile()
	file.Entries[none].From.Unit = "furlong"
	file.Entries[one].Shown = radiation.Range{Lo: 490, Hi: 470}
	got, err := file.Profiles()
	if err == nil || got != nil {
		test.Fatalf("%v and %v", got, err)
	}
	for _, says := range []string{`profile "x": its from end`,
		`profile "y": shown from 490 to 470 nm runs backwards`} {
		if !strings.Contains(err.Error(), says) {
			test.Errorf("%q does not say %q", err, says)
		}
	}
}

// TestReadingRefusesAKeyTheFileDoesNotHave is the anti-test: a key the
// file does not have is refused at that key, with no file, whether it is
// misspelled in an entry or in the file.
func TestReadingRefusesAKeyTheFileDoesNotHave(test *testing.T) {
	for key, doc := range map[string]string{
		"frm":  `{"profiles": [{"name": "x", "frm": {}}]}`,
		"unit": `{"profiles": [{"name": "x", "from": {"unit": "m", "unit2": 1}}]}`,
		"prof": `{"prof": []}`,
	} {
		got, err := band.NewProfileFile([]byte(doc))
		if err == nil || got != nil || !strings.Contains(err.Error(), key) {
			test.Errorf("%s: %v and %v", key, got, err)
		}
	}
}

// TestReadingRefusesWhatIsNotJSON is the anti-test: input that is not
// JSON gives no file.
func TestReadingRefusesWhatIsNotJSON(test *testing.T) {
	got, err := band.NewProfileFile([]byte("profiles"))
	if err == nil || got != nil {
		test.Errorf("%v and %v", got, err)
	}
}

// TestReadingKeepsWhatTheFileSays: NewProfileFile keeps the entries, and
// the note and citation the file carries for its reader.
func TestReadingKeepsWhatTheFileSays(test *testing.T) {
	doc := `{"source": "a note", "profiles": [{"name": "x", "band": "cited",
		"from": {"value": 1, "unit": "keV"}, "to": {"value": 2, "unit": "keV"},
		"shown": {"lo": 470, "hi": 490}}]}`
	got, err := band.NewProfileFile([]byte(doc))
	if err != nil || got.Source != "a note" || len(got.Entries) != one ||
		got.Entries[none].Band != "cited" ||
		got.Entries[none].Shown.Hi != 490 {
		test.Errorf("%+v and %v", got, err)
	}
}

// TestEmbeddedGivesTheEmbeddedProfiles: Embedded gives the profiles that
// band.Profiles gives.
func TestEmbeddedGivesTheEmbeddedProfiles(test *testing.T) {
	want, err := band.Profiles()
	if err != nil {
		test.Fatal(err)
	}
	var source band.ProfileSource = band.NewEmbedded()
	got, err := source.Profiles()
	if err != nil || len(got) != len(want) {
		test.Errorf("%v and %v", got, err)
	}
}
