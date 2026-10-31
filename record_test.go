package nilgai_test

import (
	"slices"
	"testing"

	"github.com/janearc/nilgai"
)

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

// TestFluxesAreHeldInSI: Jansky and ErgPerSecPerCm2 turn one of their
// units into what the IAU and the cgs system define it as.
func TestFluxesAreHeldInSI(test *testing.T) {
	want := loadWorld(test).InSI
	if got := Num(nilgai.Jansky(unit)); !near(got, want.Jansky, want.Rel) {
		test.Errorf("one jansky is %v W/m^2/Hz, want %v", got, want.Jansky)
	}
	got := Num(nilgai.ErgPerSecPerCm2(unit))
	if !near(got, want.ErgPerSecPerCm2, want.Rel) {
		test.Errorf("one erg/s/cm^2 is %v W/m^2, want %v", got,
			want.ErgPerSecPerCm2)
	}
}

func TestNewDetection(test *testing.T) {
	read := loadWorld(test)
	w1 := read.W1.build()
	measurement := read.density(test, w1)
	first, last := measurement.Bands()
	value, valueIsDensity := measurement.Value().(nilgai.FluxDensity)
	sigma, sigmaIsDensity := measurement.Error().(nilgai.FluxDensity)
	if measurement.Kind() != nilgai.Detection || measurement.IsUpperLimit() ||
		measurement.Quality() != nilgai.Good ||
		!valueIsDensity || !sigmaIsDensity ||
		value != nilgai.Jansky(float64(read.Density.Value)) ||
		sigma != nilgai.Jansky(float64(read.Density.Error)) ||
		measurement.Filter() != read.W1.Name ||
		first != w1.first || last != w1.last {
		test.Errorf("%+v", measurement)
	}
}

func TestNewDetectionAcceptsZeroFlux(test *testing.T) {
	w1 := loadWorld(test).W1.build()
	if _, err := nilgai.NewDetection(w1.name, w1.first, w1.last,
		nilgai.EnergyFlux(none), nilgai.EnergyFlux(none)); err != nil {
		test.Error(err)
	}
}

// TestBadFluxesAreRefused is the anti-test: each bad flux is refused, with
// nil, as a detection's value, as its error and as an upper limit.
func TestBadFluxesAreRefused(test *testing.T) {
	read := loadWorld(test)
	w1 := read.W1.build()
	everywhere := refused{value: true, err: true, limit: true}
	for _, each := range read.BadFluxes {
		if got := each.refusals(test, read, w1); got != everywhere {
			test.Errorf("%s: refused %+v", each.Name, got)
		}
	}
}

// badFilters returns two bad filters, each made from W1 by changing one
// thing:
//   - a filter with no name
//   - a filter whose bands run backwards
func badFilters(w1 filter) []filter {
	return []filter{{name: "", first: w1.first, last: w1.last},
		{name: w1.name, first: w1.last, last: w1.first}}
}

func TestNewDetectionRefusesBadFilters(test *testing.T) {
	read := loadWorld(test)
	density := nilgai.Jansky(float64(read.Density.Value))
	for _, each := range badFilters(read.W1.build()) {
		got, err := nilgai.NewDetection(each.name, each.first, each.last,
			density, density)
		if err == nil || got != nil {
			test.Errorf("%+v was accepted", each)
		}
	}
}

func TestNewUpperLimit(test *testing.T) {
	read := loadWorld(test)
	w1 := read.W1.build()
	energy := float64(read.Energy.Value)
	limit, err := nilgai.NewUpperLimit(w1.name, w1.first, w1.last,
		nilgai.ErgPerSecPerCm2(energy))
	if err != nil || !limit.IsUpperLimit() || limit.Error() != nil ||
		limit.Kind() != nilgai.UpperLimit {
		test.Errorf("%+v, %v", limit, err)
	}
}

func TestNewUpperLimitRefusesBadFilters(test *testing.T) {
	read := loadWorld(test)
	density := nilgai.Jansky(float64(read.Density.Value))
	for _, each := range badFilters(read.W1.build()) {
		got, err := nilgai.NewUpperLimit(each.name, each.first, each.last,
			density)
		if err == nil || got != nil {
			test.Errorf("a limit through %+v was accepted", each)
		}
	}
}

func TestWithMethodsKeepWhatTheyAreGiven(test *testing.T) {
	read := loadWorld(test)
	plain := read.density(test, read.W1.build())
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
	plain := read.density(test, read.W1.build())
	for _, each := range read.InvalidNumbers {
		if _, err := plain.WithReading(nilgai.Quality(each), nil); err == nil {
			test.Errorf("quality %d was accepted", each)
		}
	}
}

// source returns the fixture's source.
func source(test *testing.T, read world,
	measurements []nilgai.Measurement) (*nilgai.Source, error) {
	test.Helper()
	return nilgai.NewSource(nilgai.Fermi, "4FGL J1745.6-2859",
		read.position(test), read.ellipse(test), nilgai.When{}, measurements)
}

func TestNewSource(test *testing.T) {
	read := loadWorld(test)
	one := read.density(test, read.W1.build())
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
	one := []nilgai.Measurement{read.density(test, read.W1.build())}
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
		if err == nil || got != nil {
			test.Errorf("%s was accepted", each.name)
		}
	}
}

// sourceArgs holds the arguments of one call to NewSource that must fail.
type sourceArgs struct {
	name         string
	catalogue    nilgai.Catalogue
	id           string
	ellipse      nilgai.ErrorEllipse
	measurements []nilgai.Measurement
}

func TestSourceCopiesItsMeasurements(test *testing.T) {
	read := loadWorld(test)
	one := []nilgai.Measurement{read.density(test, read.W1.build())}
	got, _ := source(test, read, one)
	one[none] = nilgai.Measurement{}
	got.Measurements()[none] = nilgai.Measurement{}
	if got.Measurements()[none].Kind() != nilgai.Detection {
		test.Error("a caller's slice changed the source")
	}
}

func TestEveryCatalogueIsAccepted(test *testing.T) {
	read := loadWorld(test)
	one := []nilgai.Measurement{read.density(test, read.W1.build())}
	for _, each := range []nilgai.Catalogue{nilgai.MeerKAT, nilgai.Fermi,
		nilgai.Chandra, nilgai.NEOWISE} {
		if _, err := nilgai.NewSource(each, "id", read.position(test),
			read.ellipse(test), nilgai.When{}, one); err != nil {
			test.Errorf("%s: %v", each, err)
		}
	}
}
