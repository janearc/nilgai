package read

import (
	"errors"
	"fmt"
	"io"
	"iter"

	"github.com/janearc/libtheme-css/primitives/bands"
	"github.com/janearc/libtheme-css/spaces/radiation"
	"github.com/janearc/nilgai"
)

// Chandra reads a Chandra Source Catalog search as the CSC's TAP service
// writes it for text/csv. The file is tab-separated: column descriptions
// on lines starting with #, then a header row, then one row for each
// source. Chandra is a nilgai.ChandraReader. Make one with NewChandra.
type Chandra struct {
	first, last bands.Band
}

// Chandra reads what a nilgai.ChandraReader reads.
var _ nilgai.ChandraReader = Chandra{}

// chandraNumbers lists the numeric columns a search must select, by the
// CSC's names for them:
//   - ra and dec, the ICRS position in degrees
//   - err_ellipse_r0 and err_ellipse_r1, the radii of the 95% error
//     ellipse in arcseconds
//   - err_ellipse_ang, the ellipse's angle in degrees
//   - flux_aper_b, the broad-band energy flux
//   - flux_aper_lolim_b and flux_aper_hilim_b, its 68% limits
var chandraNumbers = []string{"ra", "dec", "err_ellipse_r0",
	"err_ellipse_r1", "err_ellipse_ang", "flux_aper_b",
	"flux_aper_lolim_b", "flux_aper_hilim_b"}

// chandraColumns lists every column a search must select: name, the
// source's name, then chandraNumbers.
var chandraColumns = append([]string{"name"}, chandraNumbers...)

// chandraStated is the confidence level at which the CSC states its
// ellipses. The CSC's own column descriptions call the ellipse a "95%
// confidence level position error ellipse".
const chandraStated = 0.95

// chandraBand is the filter name nilgai gives the CSC's broad band.
const chandraBand = "b"

// chandraProfile names the profile in band/data/profiles.json that holds
// the CSC's band edges, 0.5 to 7 keV.
const chandraProfile = "chandra"

// NewChandra returns a Chandra that takes its band from the profile named
// chandra among the profiles given, as band.Profiles reads them.
// NewChandra refuses profiles that have no profile of that name, and
// returns nil on a refusal.
func NewChandra(profiles []radiation.Profile) (*Chandra, error) {
	first, last, err := profileBands(profiles, chandraProfile)
	if err != nil {
		return nil, err
	}
	return &Chandra{first: first, last: last}, nil
}

// Read yields a *Source for each row of the search, or nil and a
// *RowError for a row it cannot read. A search without the columns it
// needs, or input that is not a table at all, is yielded once as nil and
// an error, and ends the read.
func (table Chandra) Read(name string,
	in io.Reader) iter.Seq2[*nilgai.Source, error] {
	reading := chandraReading{table: table, name: name, in: in}
	return reading.each
}

// chandraReading holds one call of Read on one input.
type chandraReading struct {
	table Chandra
	name  string
	in    io.Reader
}

// each yields the input's sources, in order, until the input ends or the
// caller stops.
func (reading chandraReading) each(yield func(*nilgai.Source, error) bool) {
	rows := tsv(reading.in)
	at, err := header(rows, chandraColumns)
	if err != nil {
		yield(nil, fmt.Errorf("%s: %w", reading.name, err))
		return
	}
	for row := firstDataRow; ; row++ {
		fields, err := rows.Read()
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			err = fmt.Errorf("%s: row %d: %w", reading.name, row,
				err)
			yield(nil, err)
			return
		}
		source, column, err := reading.table.source(cellsOf(fields, at))
		if err != nil {
			err = nilgai.NewRowError(reading.name, row, column, err)
		}
		if !yield(source, err) {
			return
		}
	}
}

// source converts one row of a search to a Source, or returns nil, the
// column that could not be read, and why. When the nilgai record refuses
// the Source, the error names no column.
func (table Chandra) source(
	cells map[string]string,
) (*nilgai.Source, string, error) {
	got, column, err := numbers(cells, chandraNumbers...)
	if err != nil {
		return nil, column, err
	}
	position, err := nilgai.NewPosition(got["ra"], got["dec"])
	if err != nil {
		return nil, "ra", err
	}
	ellipse, column, err := chandraEllipse(got)
	if err != nil {
		return nil, column, err
	}
	flux, err := table.flux(got)
	if err != nil {
		return nil, "flux_aper_b", err
	}
	var epoch nilgai.When // the CSC states no epoch for its positions
	source, err := nilgai.NewSource(nilgai.Chandra, cells["name"], position,
		ellipse, epoch, []nilgai.Measurement{*flux})
	return source, "", err
}

// chandraEllipse returns the row's 95% ellipse, scaled to one sigma. The
// CSC writes its two radii as r0, the major, and r1, the minor. They are
// equal for most sources, and r1 sometimes comes out larger in the last
// digits. The larger radius is taken as the major.
//
// chandraEllipse takes the CSC's angle as r0's. When r1 is the larger,
// the axes swap: the major axis is a quarter turn from r0's, so the angle
// turns a quarter turn too.
//
// chandraEllipse checks the angle against [0, 360) as the CSC gave it,
// before ordering the axes, so a bad angle is refused the same way
// whichever radius is larger, and on its own column.
func chandraEllipse(got map[string]float64) (ellipse nilgai.ErrorEllipse,
	column string, err error) {
	given := got["err_ellipse_ang"]
	if !nilgai.NotBelowNone(given) || given >= nilgai.FullCircle {
		return ellipse, "err_ellipse_ang", fmt.Errorf(
			"position angle %v deg is outside [0, 360)", given)
	}
	major, minor, angle := orderAxes(got["err_ellipse_r0"],
		got["err_ellipse_r1"], given)
	ellipse, err = nilgai.NewErrorEllipse(major, minor, angle,
		nilgai.Confidence(chandraStated))
	return ellipse, "err_ellipse_r0", err
}

// flux returns the broad-band flux as a detection. The CSC gives 68%
// limits, which are one sigma, and they are not quite even on either side
// of the flux. The error is half the width between them. The CSC gives
// all three in erg/s/cm^2.
func (table Chandra) flux(got map[string]float64) (*nilgai.Measurement,
	error) {
	width := got["flux_aper_hilim_b"] - got["flux_aper_lolim_b"]
	sigma := nilgai.ErgPerSecPerCm2(nilgai.Half(width))
	return nilgai.NewDetection(chandraBand, table.first, table.last,
		nilgai.ErgPerSecPerCm2(got["flux_aper_b"]), sigma)
}
