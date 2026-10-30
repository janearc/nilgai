package nilgai_test

import (
	"slices"
	"testing"

	"github.com/janearc/libtheme-css/primitives/bands"
	"github.com/janearc/nilgai"
)

// Each fixture filter covers the libtheme bands the fixture says. W1,
// 2800 nm to 3800 nm, starts in band 274 and ends in 285. A filter whose
// high edge lands exactly on a band's low edge stops in the band below,
// and one whose low edge lands exactly on a band's low edge starts in
// that band.
func TestBandsBetween(test *testing.T) {
	for _, each := range loadWorld(test).Filters {
		first, last, err := nilgai.BandsBetween(float64(each.Lo),
			float64(each.Hi))
		if err != nil || first != bands.Band(each.First) ||
			last != bands.Band(each.Last) {
			test.Errorf("%s: %d to %d, %v; want %d to %d", each.Name,
				int(first), int(last), err, each.First, each.Last)
		}
	}
}

// The edge cases are what they claim to be: their edges sit exactly on
// libtheme's, so the test above is testing the edge and not near it.
func TestBandsBetweenEdgesAreExact(test *testing.T) {
	for _, each := range loadWorld(test).Filters {
		if each.OnEdge == "" {
			continue
		}
		edge := map[string]Num{"lo": each.Lo, "hi": each.Hi}[each.OnEdge]
		start, _ := bands.Of(float64(edge)).Edges()
		if Num(start) != edge {
			test.Errorf("%s: %v nm is not a libtheme edge", each.Name, edge)
		}
	}
}

func TestBandsBetweenRefuses(test *testing.T) {
	for _, each := range loadWorld(test).BadBands {
		first, last, err := nilgai.BandsBetween(float64(each.Lo),
			float64(each.Hi))
		if err == nil || first != bands.Band(none) ||
			last != bands.Band(none) {
			test.Errorf("%q was accepted: %d to %d", each.Name,
				int(first), int(last))
		}
	}
}

func TestQualityInWords(test *testing.T) {
	words := map[nilgai.Quality]string{nilgai.Good: "good",
		nilgai.Doubtful: "doubtful", nilgai.Bad: "bad"}
	for quality, want := range words {
		if quality.String() != want {
			test.Errorf("%d: %q", int(quality), quality.String())
		}
	}
}

func TestQualityOutOfRangeIsInvalid(test *testing.T) {
	for _, each := range loadWorld(test).InvalidNumbers {
		if got := nilgai.Quality(each).String(); got != "invalid" {
			test.Errorf("%d: %q", each, got)
		}
	}
}

func TestKindInWords(test *testing.T) {
	words := map[nilgai.Kind]string{nilgai.Detection: "detection",
		nilgai.UpperLimit: "upper limit"}
	for kind, want := range words {
		if kind.String() != want {
			test.Errorf("%d: %q", int(kind), kind.String())
		}
	}
}

func TestKindOutOfRangeIsInvalid(test *testing.T) {
	for _, each := range loadWorld(test).InvalidNumbers {
		if got := nilgai.Kind(each).String(); got != "invalid" {
			test.Errorf("%d: %q", each, got)
		}
	}
}

func TestFlagKeepsWhatItWasGiven(test *testing.T) {
	for _, value := range []string{" D0 ", ""} {
		flag, err := nilgai.NewFlag("cc_flags", value)
		if err != nil || flag.Name() != "cc_flags" || flag.Value() != value {
			test.Errorf("%q: %+v, %v", value, flag, err)
		}
	}
}

func TestFlagWithNoColumnIsRefused(test *testing.T) {
	if flag, err := nilgai.NewFlag("", "D0"); err == nil ||
		flag != (nilgai.Flag{}) {
		test.Errorf("%+v, %v", flag, err)
	}
}

func TestWhen(test *testing.T) {
	var unknown nilgai.When
	if _, known := unknown.MJD(); known {
		test.Error("the zero When is known")
	}
	for _, each := range loadWorld(test).WhenOk {
		when, err := nilgai.NewWhen(float64(each))
		mjd, known := when.MJD()
		if err != nil || !known || Num(mjd) != each {
			test.Errorf("%v: %v, %v", each, when, err)
		}
	}
}

func TestWhenRefuses(test *testing.T) {
	for _, each := range loadWorld(test).WhenRefused {
		when, err := nilgai.NewWhen(float64(each))
		if _, known := when.MJD(); err == nil || known {
			test.Errorf("%v was accepted", each)
		}
	}
}

func TestFluxUnitsDiffer(test *testing.T) {
	read := loadWorld(test)
	density := nilgai.FluxDensity(read.Density.Value)
	energy := nilgai.EnergyFlux(read.Density.Value)
	if density.Unit() != nilgai.Jansky ||
		energy.Unit() != nilgai.ErgPerSecPerCm2 ||
		density.Unit() == energy.Unit() {
		test.Errorf("%v and %v", density.Unit(), energy.Unit())
	}
	if density.Value() != energy.Value() {
		test.Error("the same number reads differently")
	}
}

func TestNewDetection(test *testing.T) {
	read := loadWorld(test)
	w1 := read.W1.build(test)
	measurement := read.density(test, w1)
	first, last := measurement.Bands()
	if measurement.Kind() != nilgai.Detection || measurement.IsUpperLimit() ||
		measurement.Quality() != nilgai.Good ||
		Num(measurement.Value().Value()) != read.Density.Value ||
		Num(measurement.Error().Value()) != read.Density.Error ||
		measurement.Filter() != read.W1.Name ||
		first != w1.first || last != w1.last {
		test.Errorf("%+v", measurement)
	}
}

func TestNewDetectionAcceptsZeroFlux(test *testing.T) {
	w1 := loadWorld(test).W1.build(test)
	if _, err := nilgai.NewDetection(w1.name, w1.first, w1.last,
		nilgai.EnergyFlux(none), nilgai.EnergyFlux(none)); err != nil {
		test.Error(err)
	}
}

// mixedPair is a value and an error in different units.
type mixedPair struct{ value, sigma nilgai.Flux }

func TestNewDetectionRefusesMixedUnits(test *testing.T) {
	read := loadWorld(test)
	w1 := read.W1.build(test)
	density := nilgai.FluxDensity(read.Density.Value)
	energy := nilgai.EnergyFlux(read.Density.Value)
	for _, each := range []mixedPair{{density, energy}, {energy, density}} {
		got, err := nilgai.NewDetection(w1.name, w1.first, w1.last,
			each.value, each.sigma)
		if err == nil || got.Value() != nil {
			test.Errorf("%v with %v was accepted", each.value.Unit(),
				each.sigma.Unit())
		}
	}
}

func TestNewDetectionRefusesBadFluxes(test *testing.T) {
	read := loadWorld(test)
	w1 := read.W1.build(test)
	good := nilgai.FluxDensity(read.Density.Value)
	for _, each := range read.BadFluxes {
		bad := each.flux()
		if _, err := nilgai.NewDetection(w1.name, w1.first, w1.last,
			bad, good); err == nil {
			test.Errorf("%s was accepted as a value", each.Name)
		}
		if _, err := nilgai.NewDetection(w1.name, w1.first, w1.last,
			good, bad); err == nil {
			test.Errorf("%s was accepted as an error", each.Name)
		}
	}
}

// badFilters are a filter with no name, and one whose bands run
// backwards, made from W1 by changing one thing each.
func badFilters(w1 filter) []filter {
	return []filter{{name: "", first: w1.first, last: w1.last},
		{name: w1.name, first: w1.last, last: w1.first}}
}

func TestNewDetectionRefusesBadFilters(test *testing.T) {
	read := loadWorld(test)
	density := nilgai.FluxDensity(read.Density.Value)
	for _, each := range badFilters(read.W1.build(test)) {
		if _, err := nilgai.NewDetection(each.name, each.first, each.last,
			density, density); err == nil {
			test.Errorf("%+v was accepted", each)
		}
	}
}

func TestNewUpperLimit(test *testing.T) {
	read := loadWorld(test)
	w1 := read.W1.build(test)
	limit, err := nilgai.NewUpperLimit(w1.name, w1.first, w1.last,
		nilgai.EnergyFlux(read.Energy.Value))
	if err != nil || !limit.IsUpperLimit() || limit.Error() != nil ||
		limit.Kind() != nilgai.UpperLimit {
		test.Errorf("%+v, %v", limit, err)
	}
}

func TestNewUpperLimitRefuses(test *testing.T) {
	read := loadWorld(test)
	w1 := read.W1.build(test)
	for _, each := range read.BadFluxes {
		if _, err := nilgai.NewUpperLimit(w1.name, w1.first, w1.last,
			each.flux()); err == nil {
			test.Errorf("%s was accepted", each.Name)
		}
	}
	density := nilgai.FluxDensity(read.Density.Value)
	for _, each := range badFilters(w1) {
		if _, err := nilgai.NewUpperLimit(each.name, each.first, each.last,
			density); err == nil {
			test.Errorf("a limit through %+v was accepted", each)
		}
	}
}

func TestWithMethodsKeepWhatTheyAreGiven(test *testing.T) {
	read := loadWorld(test)
	plain := read.density(test, read.W1.build(test))
	flag, _ := nilgai.NewFlag("flag", " D0 ")
	reading, err := plain.WithReading(nilgai.Doubtful, []nilgai.Flag{flag})
	if err != nil {
		test.Fatal(err)
	}
	raw := reading.WithRaw("12.345")
	latest := float64(slices.Max(read.WhenOk))
	when, _ := nilgai.NewWhen(latest)
	timed := raw.WithTime(when)
	flags := timed.Flags()
	flags[none] = nilgai.Flag{}
	if timed.Quality() != nilgai.Doubtful || timed.Raw() != "12.345" ||
		timed.Flags()[none].Value() != " D0 " || len(timed.Flags()) != len(flags) {
		test.Errorf("%+v", timed)
	}
	if _, known := timed.Time().MJD(); !known {
		test.Error("the time was lost")
	}
	if plain.Quality() != nilgai.Good || plain.Raw() != "" {
		test.Error("a With method changed the original")
	}
}

func TestWithReadingRefusesAnInvalidQuality(test *testing.T) {
	read := loadWorld(test)
	plain := read.density(test, read.W1.build(test))
	for _, each := range read.InvalidNumbers {
		if _, err := plain.WithReading(nilgai.Quality(each), nil); err == nil {
			test.Errorf("quality %d was accepted", each)
		}
	}
}

// source builds the fixture's source.
func source(test *testing.T, read world,
	measurements []nilgai.Measurement) (nilgai.Source, error) {
	test.Helper()
	return nilgai.NewSource(nilgai.Fermi, "4FGL J1745.6-2859",
		read.position(test), read.ellipse(test), nilgai.When{}, measurements)
}

func TestNewSource(test *testing.T) {
	read := loadWorld(test)
	one := read.density(test, read.W1.build(test))
	measurements := []nilgai.Measurement{one}
	got, err := source(test, read, measurements)
	if err != nil || got.ID() != "4FGL J1745.6-2859" ||
		got.Catalogue() != nilgai.Fermi || len(got.Measurements()) != len(measurements) ||
		got.Position() != read.position(test) ||
		got.Error() != read.ellipse(test) {
		test.Fatalf("%+v, %v", got, err)
	}
	if _, known := got.Epoch().MJD(); known {
		test.Error("a source with no epoch has one")
	}
}

func TestNewSourceRefuses(test *testing.T) {
	read := loadWorld(test)
	one := []nilgai.Measurement{read.density(test, read.W1.build(test))}
	position, ellipse := read.position(test), read.ellipse(test)
	for _, each := range []sourceArgs{
		{"unknown catalogue", "other", "id", ellipse, one},
		{"no id", nilgai.Fermi, "", ellipse, one},
		{"no ellipse", nilgai.Fermi, "id", nilgai.ErrorEllipse{}, one},
		{"no measurements", nilgai.Fermi, "id", ellipse, nil},
		{"unmade measurement", nilgai.Fermi, "id", ellipse,
			[]nilgai.Measurement{{}}},
	} {
		got, err := nilgai.NewSource(each.catalogue, each.id, position,
			each.ellipse, nilgai.When{}, each.measurements)
		if err == nil || got.ID() != "" {
			test.Errorf("%s was accepted", each.name)
		}
	}
}

// sourceArgs is one try at NewSource that must fail.
type sourceArgs struct {
	name         string
	catalogue    nilgai.Catalogue
	id           string
	ellipse      nilgai.ErrorEllipse
	measurements []nilgai.Measurement
}

func TestSourceCopiesItsMeasurements(test *testing.T) {
	read := loadWorld(test)
	one := []nilgai.Measurement{read.density(test, read.W1.build(test))}
	got, _ := source(test, read, one)
	one[none] = nilgai.Measurement{}
	got.Measurements()[none] = nilgai.Measurement{}
	if got.Measurements()[none].Kind() != nilgai.Detection {
		test.Error("a caller's slice changed the source")
	}
}

func TestEveryCatalogueIsAccepted(test *testing.T) {
	read := loadWorld(test)
	one := []nilgai.Measurement{read.density(test, read.W1.build(test))}
	for _, each := range []nilgai.Catalogue{nilgai.MeerKAT, nilgai.Fermi,
		nilgai.Chandra, nilgai.NEOWISE} {
		if _, err := nilgai.NewSource(each, "id", read.position(test),
			read.ellipse(test), nilgai.When{}, one); err != nil {
			test.Errorf("%s: %v", each, err)
		}
	}
}
