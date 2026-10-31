package nilgai_test

import (
	"cmp"
	"errors"
	"slices"
	"testing"

	"github.com/janearc/nilgai"
)

// looks returns the fixture's looks: the first and last in W1, and the
// second in W2, as the fixture's times give them.
func looks(test *testing.T, read world) []nilgai.Measurement {
	test.Helper()
	filters := map[string]filter{read.W1.Name: read.W1.build(),
		read.W2.Name: read.W2.build()}
	var made []nilgai.Measurement
	for _, each := range read.Looks {
		made = append(made, read.look(test, filters[each.Band], each.MJD))
	}
	return made
}

func curve(test *testing.T, read world,
	points []nilgai.Measurement) (*nilgai.LightCurve, error) {
	test.Helper()
	return nilgai.NewLightCurve("J1745-2859", read.position(test),
		read.ellipse(test), points)
}

func TestLightCurve(test *testing.T) {
	read := loadWorld(test)
	in := looks(test, read)
	got, err := curve(test, read, in)
	if err != nil || got.Len() != len(in) || got.ID() != "J1745-2859" ||
		got.Position() != read.position(test) ||
		got.Error() != read.ellipse(test) {
		test.Fatalf("%+v, %v", got, err)
	}
	first, last, ok := got.Span()
	if !ok || Num(first) != read.Span.First || Num(last) != read.Span.Last {
		test.Errorf("span %v %v %v", first, last, ok)
	}
	if !slices.IsSortedFunc(got.Points(), byTime) {
		test.Error("points are out of order")
	}
}

// byTime orders two looks by MJD.
func byTime(first, second nilgai.Measurement) int {
	firstMJD, _ := first.Time().MJD()
	secondMJD, _ := second.Time().MJD()
	return cmp.Compare(firstMJD, secondMJD)
}

func TestLightCurveDoesNotReorderTheCallersLooks(test *testing.T) {
	read := loadWorld(test)
	in := looks(test, read)
	if _, err := curve(test, read, in); err != nil {
		test.Fatal(err)
	}
	if mjd, _ := in[none].Time().MJD(); Num(mjd) != read.Looks[none].MJD {
		test.Error("NewLightCurve sorted the caller's slice")
	}
}

func TestLightCurveInFilter(test *testing.T) {
	read := loadWorld(test)
	got, _ := curve(test, read, looks(test, read))
	if within := got.InFilter(read.W2.Name); len(within) != read.Span.W2Looks {
		test.Errorf("W2 looks: %d", len(within))
	}
	if within := got.InFilter("no such filter"); len(within) != none {
		test.Errorf("looks through a filter that is not there: %d",
			len(within))
	}
}

func TestLightCurveRefuses(test *testing.T) {
	read := loadWorld(test)
	good := looks(test, read)
	untimed := []nilgai.Measurement{read.density(test, read.W1.build())}
	position, ellipse := read.position(test), read.ellipse(test)
	timed, err := nilgai.NewWhen(float64(read.Looks[none].MJD))
	if err != nil {
		test.Fatal(err)
	}
	for _, each := range []curveArgs{
		{"no id", "", ellipse, good},
		{"no ellipse", "x", nilgai.ErrorEllipse{}, good},
		{"no points", "x", ellipse, nil},
		{"no time", "x", ellipse, untimed},
		{"unmade point", "x", ellipse, []nilgai.Measurement{{}}},
		{"unmade point with a time", "x", ellipse,
			[]nilgai.Measurement{nilgai.Measurement{}.WithTime(timed)}},
	} {
		got, err := nilgai.NewLightCurve(each.id, position, each.ellipse,
			each.points)
		if err == nil || got != nil {
			test.Errorf("%s was accepted", each.name)
		}
	}
}

// curveArgs holds the arguments of one call to NewLightCurve that must
// fail.
type curveArgs struct {
	name    string
	id      string
	ellipse nilgai.ErrorEllipse
	points  []nilgai.Measurement
}

func TestAnEmptyLightCurveHasNoSpan(test *testing.T) {
	var empty nilgai.LightCurve
	if _, _, ok := empty.Span(); ok || empty.Len() != none {
		test.Error("an empty curve has a span or a length")
	}
}

func TestRowError(test *testing.T) {
	spec := loadWorld(test).RowError
	cause := errors.New("not a number")
	got := nilgai.NewRowError(spec.File, spec.Row, spec.Column, cause)
	if got.Error() != spec.Text || got.File() != spec.File ||
		got.Row() != spec.Row || got.Column() != spec.Column {
		test.Errorf("%q", got.Error())
	}
	if !errors.Is(got, cause) {
		test.Error("RowError hides its cause")
	}
	var found *nilgai.RowError
	if !errors.As(error(got), &found) || found != got {
		test.Error("errors.As does not find the RowError")
	}
}

func TestRowErrorDoesNotMatchAStranger(test *testing.T) {
	spec := loadWorld(test).RowError
	got := nilgai.NewRowError(spec.File, spec.Row, spec.Column,
		errors.New("bad"))
	if errors.Is(got, errors.New("other")) {
		test.Error("RowError matches an unrelated error")
	}
}
