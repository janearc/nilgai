package read

import (
	"errors"
	"fmt"
	"io"
	"iter"
	"math"
	"strconv"
	"strings"

	"github.com/astrogo/fitsio"
	"github.com/janearc/libtheme-css/primitives/bands"
	"github.com/janearc/nilgai"
	"github.com/janearc/nilgai/band"
)

// Fermi reads the Fermi-LAT 4FGL catalogue, the FITS file that the Fermi
// Science Support Center provides (gll_psc_v35.fit for DR4). It reads the
// point source table, which has one row for each source, unassociated
// sources included. Make one with NewFermi.
type Fermi struct {
	first, last bands.Band
}

// fermiTableName is the name that 4FGL gives its point source table.
const fermiTableName = "LAT_Point_Source_Catalog"

// fermiRow holds the columns nilgai reads from a 4FGL row. Positions are
// J2000 in degrees, which nilgai takes as ICRS. The two agree to a few
// hundredths of an arcsecond, far inside any Fermi ellipse. The 95%
// ellipse's axes and angle are in degrees.
type fermiRow struct {
	name                string
	ra, dec             float32
	major, minor, angle float32
	flux, fluxError     float32
	class, assoc        string
	flags               int16
}

// fermiStated is the confidence level at which 4FGL states its ellipses.
// Conf_95_SemiMajor and the other ellipse columns carry it in their
// names.
const fermiStated = 0.95

// fermiFilter is the label nilgai gives 4FGL's energy flux. fermiFilterData
// is the filter's name in band/data/filters.json, which holds its edges
// and their citation.
const (
	fermiFilter     = "100 MeV to 100 GeV"
	fermiFilterData = "Energy_Flux100"
)

// 4FGL's own names for the columns nilgai reads or keeps.
const (
	fermiFlagsColumn   = "Flags"
	fermiClassColumn   = "CLASS1"
	fermiAssocColumn   = "ASSOC1"
	fermiEllipseColumn = "Conf_95_SemiMajor"
)

// noFlags is the Flags value that 4FGL uses to say no analysis flag is
// set on the source.
const noFlags = 0

// extendedEnd is the letter that ends the name of an extended source in
// 4FGL.
const extendedEnd = "e"

// NewFermi constructs a 4FGL reader using the Energy_Flux100 filter edges in
// band/data/filters.json. Invalid filter data returns nil and an error.
func NewFermi() (*Fermi, error) {
	return fermiOver(band.FilterBands(fermiFilterData))
}

// ReadFermi constructs a reader with filter JSON in band/data/filters.json’s
// format. Invalid filters or a missing Energy_Flux100 return nil and an error
// from band.ReadFilterBands.
func ReadFermi(filters []byte) (*Fermi, error) {
	return fermiOver(band.ReadFilterBands(filters, fermiFilterData))
}

// fermiOver returns a Fermi built from the energy flux bands given. If
// the bands arrive with an error, fermiOver returns nil and that error
// instead.
func fermiOver(first, last bands.Band, err error) (*Fermi, error) {
	if err != nil {
		return nil, err
	}
	return &Fermi{first: first, last: last}, nil
}

// Read yields a *Source for each row of the point source table, or nil
// and a *RowError for a row it cannot read. Two kinds of row have no
// position ellipse and cannot be read:
//   - an extended source
//   - a source whose position 4FGL fixes from outside
//
// Input that is not a 4FGL file is yielded once as nil and an error, and
// ends the read.
func (table Fermi) Read(name string,
	in io.Reader) iter.Seq2[*nilgai.Source, error] {
	reading := fermiReading{table: table, name: name, in: in}
	return reading.each
}

// fermiReading holds one call of Read on one input.
type fermiReading struct {
	table Fermi
	name  string
	in    io.Reader
}

// each yields the file's sources, in order, until the file ends or the
// caller stops.
func (reading fermiReading) each(yield func(*nilgai.Source, error) bool) {
	rows, err := fermiRows(reading.in)
	if err != nil {
		yield(nil, fmt.Errorf("%s: %w", reading.name, err))
		return
	}
	defer rows.Close()
	for row := firstDataRow; rows.Next(); row++ {
		var source *nilgai.Source
		var column string
		var got fermiRow
		err := scanFermi(rows, &got)
		if err == nil {
			source, column, err = reading.table.source(got)
		}
		if err != nil {
			err = nilgai.NewRowError(reading.name, row, column, err)
		}
		if !yield(source, err) {
			return
		}
	}
}

// scanFermi fills the row with the current row's columns, by 4FGL's names
// for them. If fitsio could not read them it returns the error, and the
// caller must not use the row.
func scanFermi(rows *fitsio.Rows, into *fermiRow) error {
	var scanned struct {
		Name      string  `fits:"Source_Name"`
		RA        float32 `fits:"RAJ2000"`
		Dec       float32 `fits:"DEJ2000"`
		Major     float32 `fits:"Conf_95_SemiMajor"`
		Minor     float32 `fits:"Conf_95_SemiMinor"`
		Angle     float32 `fits:"Conf_95_PosAng"`
		Flux      float32 `fits:"Energy_Flux100"`
		FluxError float32 `fits:"Unc_Energy_Flux100"`
		Class     string  `fits:"CLASS1"`
		Assoc     string  `fits:"ASSOC1"`
		Flags     int16   `fits:"Flags"`
	}
	err := rows.Scan(&scanned)
	*into = fermiRow{
		name:      scanned.Name,
		ra:        scanned.RA,
		dec:       scanned.Dec,
		major:     scanned.Major,
		minor:     scanned.Minor,
		angle:     scanned.Angle,
		flux:      scanned.Flux,
		fluxError: scanned.FluxError,
		class:     scanned.Class,
		assoc:     scanned.Assoc,
		flags:     scanned.Flags,
	}
	return err
}

// fermiColumns lists required columns by the FITS tags used in scanFermi.
// fitsio silently leaves absent columns at zero, so validate them before
// reading: zero positions or flags must come from the catalogue itself.
var fermiColumns = []string{"Source_Name", "RAJ2000", "DEJ2000",
	fermiEllipseColumn, "Conf_95_SemiMinor", "Conf_95_PosAng",
	"Energy_Flux100", "Unc_Energy_Flux100", fermiClassColumn,
	fermiAssocColumn, fermiFlagsColumn}

// fermiRows opens the point-source table after checking required columns.
func fermiRows(in io.Reader) (*fitsio.Rows, error) {
	file, err := fitsio.Open(in)
	if err != nil {
		return nil, err
	}
	table, isTable := file.Get(fermiTableName).(*fitsio.Table)
	if !isTable {
		return nil, fmt.Errorf("no table %s", fermiTableName)
	}
	for _, name := range fermiColumns {
		if table.Index(name) < none {
			return nil, fmt.Errorf("table %s has no column %s",
				fermiTableName, name)
		}
	}
	return table.Read(none, table.NumRows())
}

// source converts one row into the common record. On failure it returns nil,
// the affected column, and the error. Record-level validation errors have no
// column name.
func (table Fermi) source(got fermiRow) (*nilgai.Source, string, error) {
	position, err := nilgai.NewPosition(float64(got.ra), float64(got.dec))
	if err != nil {
		return nil, "RAJ2000", err
	}
	ellipse, column, err := fermiEllipse(got)
	if err != nil {
		return nil, column, err
	}
	energy, energyError := float64(got.flux), float64(got.fluxError)
	flux, err := nilgai.NewDetection(fermiFilter, table.first, table.last,
		nilgai.ErgPerSecPerCm2(energy),
		nilgai.ErgPerSecPerCm2(energyError))
	if err != nil {
		return nil, "Energy_Flux100", err
	}
	name := strings.TrimSpace(got.name)
	var epoch nilgai.When // 4FGL states no epoch for its positions
	flagged, unkept := fermiFlagged(*flux, got)
	source, err := nilgai.NewSource(nilgai.Fermi, name, position, ellipse,
		epoch, flagged)
	return source, "", errors.Join(unkept, err)
}

// fermiEllipse converts the row’s 95% ellipse from degrees to arcseconds and
// scales it to one sigma. Its angle wraps into [0, 360), east of north.
//
// 4FGL omits all three ellipse fields for extended sources and sources whose
// positions are fixed externally, such as novae or pulsars. These rows return
// an error identifying the reason; no ellipse is invented.
//
// Partially missing ellipse fields are treated as invalid data. The error names
// the first field that is not numeric.
func fermiEllipse(got fermiRow) (ellipse nilgai.ErrorEllipse, column string,
	err error) {
	given := []struct {
		column string
		value  float64
	}{{fermiEllipseColumn, float64(got.major)},
		{"Conf_95_SemiMinor", float64(got.minor)},
		{"Conf_95_PosAng", float64(got.angle)}}
	missing, first := none, ""
	for _, each := range given {
		if !nilgai.Finite(each.value) {
			missing++
			if first == "" {
				first = each.column
			}
		}
	}
	if missing == len(given) {
		return ellipse, fermiEllipseColumn, noFermiEllipse(got.name)
	}
	if missing > none {
		return ellipse, first, fmt.Errorf("%s is not a number, but "+
			"the rest of the ellipse is", first)
	}
	const arcsec = nilgai.ArcsecondsPerDegree
	major, minor, angle := orderAxes(float64(got.major)*arcsec,
		float64(got.minor)*arcsec, float64(got.angle))
	angle = math.Mod(angle+nilgai.FullCircle, nilgai.FullCircle)
	ellipse, err = nilgai.NewErrorEllipse(major, minor, angle,
		nilgai.Confidence(fermiStated))
	return ellipse, fermiEllipseColumn, err
}

// noFermiEllipse returns the reason a 4FGL source has no ellipse, judging
// by its name.
func noFermiEllipse(name string) error {
	if strings.HasSuffix(strings.TrimSpace(name), extendedEnd) {
		return errors.New("an extended source has no point position, " +
			"so 4FGL gives it no ellipse")
	}
	return errors.New("a point source whose position is fixed from " +
		"outside 4FGL has no ellipse in 4FGL")
}

// fermiFlagged preserves analysis flags, class, and association on the
// measurement, returning a one-element slice through kept. Any analysis flag
// makes quality Doubtful; no flags means Good. Validation errors return nil and
// an error.
func fermiFlagged(measurement nilgai.Measurement,
	got fermiRow) ([]nilgai.Measurement, error) {
	quality := nilgai.Good
	if got.flags != noFlags {
		quality = nilgai.Doubtful
	}
	analysis := strconv.Itoa(int(got.flags))
	return kept(measurement, quality, cell{fermiFlagsColumn, analysis},
		cell{fermiClassColumn, strings.TrimSpace(got.class)},
		cell{fermiAssocColumn, strings.TrimSpace(got.assoc)})
}
