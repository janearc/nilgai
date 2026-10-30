package nilgai_test

import (
	"testing"

	"github.com/janearc/nilgai"
)

type levelCase struct {
	Name         string
	Level        Num
	Percent      Num
	Stated       Num
	Want, Rel    Num
	Major, Minor Num
	Angle        Num
	OneMajor     Num
	OneMinor     Num
}

type ellipseFixture struct {
	Radius, LevelRefused          []levelCase
	OneSigma                      struct{ Want, Rel, Radius, RadiusRel Num }
	Percent, PercentRefused       []levelCase
	ToOneSigma, ToOneSigmaRefused []levelCase
	Ellipse, EllipseRefused       []levelCase
}

func loadEllipses(test *testing.T) ellipseFixture {
	test.Helper()
	var fixture ellipseFixture
	load(test, "ellipse", &fixture)
	return fixture
}

func TestRadiusFactor(test *testing.T) {
	for _, each := range loadEllipses(test).Radius {
		got, err := nilgai.RadiusFactor(nilgai.Confidence(each.Level))
		if err != nil || !near(Num(got), each.Want, each.Rel) {
			test.Errorf("%s: %v, %v, want %v", each.Name, got, err, each.Want)
		}
	}
}

func TestRadiusFactorRefuses(test *testing.T) {
	for _, each := range loadEllipses(test).LevelRefused {
		got, err := nilgai.RadiusFactor(nilgai.Confidence(each.Level))
		if err == nil || got != none {
			test.Errorf("%s: %v, %v", each.Name, got, err)
		}
	}
}

func TestOneSigma(test *testing.T) {
	want := loadEllipses(test).OneSigma
	level := nilgai.OneSigma()
	if !near(Num(level), want.Want, want.Rel) {
		test.Errorf("one sigma holds %v", level)
	}
	radius, err := nilgai.RadiusFactor(level)
	if err != nil || !near(Num(radius), want.Radius, want.RadiusRel) {
		test.Errorf("one sigma is %v wide, %v", radius, err)
	}
}

func TestPercentToConfidence(test *testing.T) {
	for _, each := range loadEllipses(test).Percent {
		got, err := nilgai.PercentToConfidence(float64(each.Percent))
		if err != nil || !near(Num(got), each.Want, each.Rel) {
			test.Errorf("%s: %v, %v", each.Name, got, err)
		}
	}
}

func TestPercentToConfidenceRefuses(test *testing.T) {
	for _, each := range loadEllipses(test).PercentRefused {
		got, err := nilgai.PercentToConfidence(float64(each.Percent))
		if err == nil || got != none {
			test.Errorf("%s: %v, %v", each.Name, got, err)
		}
	}
}

func TestToOneSigma(test *testing.T) {
	for _, each := range loadEllipses(test).ToOneSigma {
		got, err := nilgai.ToOneSigma(float64(each.Stated),
			nilgai.Confidence(each.Level))
		if err != nil || !near(Num(got), each.Want, each.Rel) {
			test.Errorf("%s: %v, %v", each.Name, got, err)
		}
	}
}

func TestToOneSigmaRefuses(test *testing.T) {
	for _, each := range loadEllipses(test).ToOneSigmaRefused {
		got, err := nilgai.ToOneSigma(float64(each.Stated),
			nilgai.Confidence(each.Level))
		if err == nil || got != none {
			test.Errorf("%s: %v, %v", each.Name, got, err)
		}
	}
}

func TestNewErrorEllipse(test *testing.T) {
	for _, each := range loadEllipses(test).Ellipse {
		got, err := nilgai.NewErrorEllipse(float64(each.Major),
			float64(each.Minor), float64(each.Angle),
			nilgai.Confidence(each.Level))
		major, minor := Num(got.SemiMajor()), Num(got.SemiMinor())
		if err != nil || !near(major, each.OneMajor, each.Rel) ||
			!near(minor, each.OneMinor, each.Rel) ||
			Num(got.PositionAngle()) != each.Angle ||
			Num(got.StatedAt()) != each.Level ||
			Num(got.StatedSemiMajor()) != each.Major ||
			Num(got.StatedSemiMinor()) != each.Minor {
			test.Errorf("%s: %+v, %v", each.Name, got, err)
		}
	}
}

func TestNewErrorEllipseRefuses(test *testing.T) {
	for _, each := range loadEllipses(test).EllipseRefused {
		got, err := nilgai.NewErrorEllipse(float64(each.Major),
			float64(each.Minor), float64(each.Angle),
			nilgai.Confidence(each.Level))
		if err == nil || got != (nilgai.ErrorEllipse{}) {
			test.Errorf("%s was accepted: %+v", each.Name, got)
		}
	}
}
