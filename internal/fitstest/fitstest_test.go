package fitstest_test

import (
	"bytes"
	"errors"
	"math"
	"slices"
	"testing"

	"github.com/astrogo/fitsio"
	"github.com/janearc/celestial-data/gaia"
	"github.com/janearc/nilgai/internal/fitstest"
)

// The two sizes the test writes, each defined.
const (
	fractional = 5.0 / 2 // defined: two and a half, a size with a fraction
	whole      = 2.0 * 2 // defined: four, a size with no fraction
)

// fromTheTop is the row that a read of the table starts from.
const fromTheTop = 0 // defined: row zero

// Writing and reopening a FITS table must preserve every row’s values.
func TestWriteIsReadBack(test *testing.T) {
	type row struct {
		Size float32 `fits:"Size"`
	}
	var out bytes.Buffer
	columns := []fitsio.Column{{Name: "Size", Format: "E"}}
	err := fitstest.Write(&out, "T", columns, &row{fractional},
		&row{whole})
	if err != nil {
		test.Fatal(err)
	}
	file, err := fitsio.Open(&out)
	if err != nil {
		test.Fatal(err)
	}
	table, isTable := file.Get("T").(*fitsio.Table)
	if !isTable {
		test.Fatalf("read back %v", file.Get("T"))
	}
	rows, err := table.Read(fromTheTop, table.NumRows())
	if err != nil {
		test.Fatal(err)
	}
	defer rows.Close()
	var got []float32
	for rows.Next() {
		var each row
		if err := rows.Scan(&each); err != nil {
			test.Fatal(err)
		}
		got = append(got, each.Size)
	}
	if want := []float32{fractional, whole}; !slices.Equal(got, want) {
		test.Errorf("rows read back as %v, want %v", got, want)
	}
}

// The made-up stars of the gaia test, each defined.
const (
	firstID      = 11     // defined: the first star's id
	secondID     = 12     // defined: the second's
	someParallax = 0.5    // defined: mas, a star two kiloparsecs away
	firstRow     = 0      // defined: the first row
	secondRow    = 1      // defined: the second
	thousand     = 1000   // defined: milliarcseconds in an arcsecond
	twoRows      = 2      // defined: the rows written
	noError      = 0      // defined: mas, a parallax without error
	someRA       = 266.25 // defined: degrees, a binary fraction
	someDec      = -29.5  // defined: degrees
	somePMRA     = -3.25  // defined: mas a year
	somePMDec    = -5.5   // defined: mas a year
	someRadial   = 123.5  // defined: km/s, a float32 exactly
	someRVError  = 1.75   // defined: km/s
	placeError   = 0.25   // defined: mas, a float32 exactly
	motionError  = 0.375  // defined: mas a year
	raDecTie     = 0.5    // defined: correlations, each a float32 exactly
	raPlxTie     = 0.25   // defined: a quarter
	decPlxTie    = -0.125 // defined: an eighth, the other way
	roundRoom    = 1e-12  // defined: a unit's conversion, to rounding
)

// twoStars is a product WriteGaia wrote of two stars, the first with
// every value the builder sets, read by celestial-data's reader, which
// holds each column to the archive's name, kind and unit; or the test
// stops.
func twoStars(test *testing.T) *gaia.Catalogue {
	test.Helper()
	var out bytes.Buffer
	first := fitstest.NewGaiaStar(firstID)
	first.At(someRA, someDec)
	first.Distant(someParallax, noError)
	first.Moving(somePMRA, somePMDec)
	first.Receding(someRadial, someRVError)
	first.Erring(placeError, placeError, motionError, motionError)
	first.Tied(raDecTie, raPlxTie, decPlxTie)
	err := fitstest.WriteGaia(&out, first, fitstest.NewGaiaStar(secondID))
	if err != nil {
		test.Fatal(err)
	}
	reader := bytes.NewReader(out.Bytes())
	catalogue, err := gaia.OpenCatalogue(reader)
	if err != nil {
		test.Fatal(err)
	}
	return catalogue
}

// a gaia product written by WriteGaia is read back as the archive would
// give it: two stars, in order, with their ids, the first's parallax.
func TestAGaiaProductIsReadBackAsTheArchiveWouldGiveIt(test *testing.T) {
	catalogue := twoStars(test)
	read, err := catalogue.Source(firstRow)
	if err != nil {
		test.Fatal(err)
	}
	second, err := catalogue.Source(secondRow)
	if err != nil {
		test.Fatal(err)
	}
	mas := read.Parallax().InArcseconds() * thousand
	if catalogue.Rows() != twoRows || read.ID() != firstID ||
		second.ID() != secondID || mas != someParallax {
		test.Errorf("read back %d rows: %d at %g mas, then %d",
			catalogue.Rows(), read.ID(), mas, second.ID())
	}
}

// every value the builder sets is read back as it was set, to the
// rounding of the reader's units: the place, the motions, the radial
// velocity, the errors, and the three correlations; a column not set,
// such as the RUWE or the magnitude, is nothing.
func TestEveryValueTheBuilderSetsIsReadBack(test *testing.T) {
	read, err := twoStars(test).Source(firstRow)
	if err != nil {
		test.Fatal(err)
	}
	raError, decError := read.PlaceError()
	pmRAError, pmDecError := read.MotionError()
	got := []float64{read.Place().Ascension().InDegrees(),
		read.Place().Declination().InDegrees(), read.Motion().Ascension(),
		read.Motion().Declination(), read.RadialVelocity(),
		read.RadialVelocityError(), raError.InArcseconds() * thousand,
		decError.InArcseconds() * thousand, pmRAError, pmDecError}
	want := []float64{someRA, someDec, somePMRA, somePMDec, someRadial,
		someRVError, placeError, placeError, motionError, motionError}
	for _, pair := range []struct {
		one, other gaia.Parameter
		want       float64
	}{{gaia.RA, gaia.Dec, raDecTie}, {gaia.RA, gaia.Parallax, raPlxTie},
		{gaia.Dec, gaia.Parallax, decPlxTie}} {
		between, err := read.Correlations().Between(pair.one, pair.other)
		if err != nil {
			test.Fatal(err)
		}
		got, want = append(got, between), append(want, pair.want)
	}
	for which := range want {
		if !(math.Abs(got[which]-want[which]) <= roundRoom) {
			test.Errorf("value %d read back %g, want %g", which, got[which],
				want[which])
		}
	}
	if read.RUWE() != noError || read.Magnitude() != noError {
		test.Errorf("columns not set read back as %g and %g", read.RUWE(),
			read.Magnitude())
	}
}

// The writer tests' numbers, each defined.
const (
	block     = 2880 // defined: a FITS block's bytes, the primary HDU's
	noBytes   = 0    // defined: a writer that takes nothing
	badFormat = "Q?" // defined: a column form fitsio cannot make
)

// shortWriter is a writer that takes so many bytes, then refuses.
type shortWriter struct{ left int }

// Write takes what it has room for, then refuses.
func (out *shortWriter) Write(given []byte) (int, error) {
	if out.left < len(given) {
		return noBytes, errors.New("no room")
	}
	out.left -= len(given)
	return len(given), nil
}

// the anti-test: a table is not written to a writer that refuses its
// first block or its second, nor with a column fitsio cannot make; each
// is refused. fitsio takes any row it is given, so no row is refused.
func TestATableThatCannotBeWrittenIsRefused(test *testing.T) {
	columns := []fitsio.Column{{Name: "Size", Format: "E"}}
	type row struct {
		Size float32 `fits:"Size"`
	}
	for name, out := range map[string]*shortWriter{"first block": {},
		"second block": {left: block}} {
		if err := fitstest.Write(out, "T", columns, &row{}); err == nil {
			test.Errorf("a writer refusing its %s took the table", name)
		}
	}
	var out bytes.Buffer
	bad := []fitsio.Column{{Name: "Size", Format: badFormat}}
	if err := fitstest.Write(&out, "T", bad, &row{}); err == nil {
		test.Error("a column fitsio cannot make was written")
	}
}
