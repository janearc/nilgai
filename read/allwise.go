package read

import (
	"errors"
	"fmt"
	"io"
	"iter"
	"math"
	"strconv"

	"github.com/janearc/libtheme-css/primitives/bands"
	"github.com/janearc/nilgai"
	"github.com/janearc/nilgai/band"
)

// AllWISE reads an AllWISE source catalogue search as IRSA's TAP service
// writes it for csv: comma-separated, with a header row and then one row
// for each source. Make one with NewAllWISE.
type AllWISE struct {
	edges [len(allwiseBands)]wiseEdges
}

// wiseEdges holds the first and last libtheme band of one WISE filter.
type wiseEdges struct{ first, last bands.Band }

// allwiseBand describes one of AllWISE's four bands with:
//   - band, which WISE band it is
//   - filter, its filter's name in band/data/filters.json, which is also
//     the name its measurement carries
//   - magnitude, the name of IRSA's profile-fit magnitude column
//   - sigma, the name of IRSA's column for that magnitude's error
//
// allwiseBand's place in allwiseBands is its place in cc_flags and
// ph_qual, which have one character for each band.
type allwiseBand struct {
	band             WiseBand
	filter           string
	magnitude, sigma string
}

// allwiseBands lists the four bands in the order AllWISE gives them. A
// search must select W1. It may select W2 to W4, which are read when it
// does.
var allwiseBands = [...]allwiseBand{
	{W1, "W1", "w1mpro", "w1sigmpro"},
	{W2, "W2", "w2mpro", "w2sigmpro"},
	{W3, "W3", "w3mpro", "w3sigmpro"},
	{W4, "W4", "w4mpro", "w4sigmpro"},
}

// allwiseNumbers lists the numeric columns a search must select, by
// IRSA's names for them:
//   - ra and dec, the ICRS position in degrees
//   - sigra and sigdec, the one-sigma errors of the position in arcseconds
//   - sigradec, the co-sigma of those two errors
//   - w1mpro, the W1 profile-fit magnitude
//
// allwiseBands holds the columns of the other bands.
var allwiseNumbers = []string{"ra", "dec", "sigra", "sigdec", "sigradec",
	"w1mpro"}

// allwiseColumns lists every column a search must select:
//   - designation, the source's name
//   - allwiseNumbers
//   - w1sigmpro, the W1 magnitude's error, which is empty when the
//     magnitude is an upper limit
//   - cc_flags and ph_qual, WISE's own flags
var allwiseColumns = append(append([]string{"designation"},
	allwiseNumbers...), "w1sigmpro", "cc_flags", "ph_qual")

// allwiseFlags lists the columns kept as flags: cc_flags, the
// contamination flags, and ph_qual, the photometric quality, each with one
// character for every band.
var allwiseFlags = []string{"cc_flags", "ph_qual"}

// cleanContamination is the cc_flags character that says AllWISE found no
// artefact affecting that band's photometry.
const cleanContamination = '0'

// NewAllWISE returns an AllWISE whose fluxes carry the libtheme bands of
// their own band's filter, W1 to W4, from band/data/filters.json, or nil
// on a refusal. A flux never carries another band's filter.
func NewAllWISE() (*AllWISE, error) {
	return allwiseOver(band.FilterBands)
}

// ReadAllWISE does what NewAllWISE does, with the filters given as JSON
// shaped like band/data/filters.json. ReadAllWISE refuses filters that
// band.ReadFilterBands refuses, and filters that lack any of W1 to W4,
// and returns nil on a refusal.
func ReadAllWISE(filters []byte) (*AllWISE, error) {
	return allwiseOver(filtersJSON(filters).edges)
}

// filtersJSON holds JSON shaped like band/data/filters.json.
type filtersJSON []byte

// edges returns the first and last libtheme band of the named filter.
func (filters filtersJSON) edges(name string) (first, last bands.Band,
	err error) {
	return band.ReadFilterBands(filters, name)
}

// allwiseOver returns an AllWISE built from the bands that edges returns
// for each of the four filters, looked up by name. If edges fails for any
// filter, allwiseOver returns nil and the first error.
func allwiseOver(edges func(name string) (bands.Band, bands.Band,
	error)) (*AllWISE, error) {
	table := &AllWISE{}
	for place, each := range allwiseBands {
		first, last, err := edges(each.filter)
		if err != nil {
			return nil, err
		}
		table.edges[place] = wiseEdges{first: first, last: last}
	}
	return table, nil
}

// Read yields a *Source for each row of the search, or nil and a
// *RowError for a row it cannot read. A search without the columns it
// needs, or input that is not a table at all, is yielded once as nil and
// an error, and ends the read.
//
// Read refuses a row, naming the column, when:
//   - a number cell is empty or not a number
//   - ra or dec is outside the sky
//   - sigra or sigdec is negative, NaN or infinite
//   - the errors give no ellipse
//   - w1sigmpro is negative, NaN or infinite
//
// Read also takes the magnitude and error of W2, W3 and W4 when a search
// selects them (w2mpro and w2sigmpro, and so on), both columns of a band
// or neither. Each source then has a measurement in each band selected,
// W1 to W4 in order.
//
// Read refuses the other bands in the same way as W1, naming the column,
// with one exception. A band whose magnitude and error are both empty is
// left out, because AllWISE has no measurement in that band.
func (table AllWISE) Read(name string,
	in io.Reader) iter.Seq2[*nilgai.Source, error] {
	reading := allwiseReading{table: table, name: name, in: in}
	return reading.each
}

// allwiseReading holds one call of Read on one input.
type allwiseReading struct {
	table AllWISE
	name  string
	in    io.Reader
}

// each yields the input's sources, in order, until the input ends or the
// caller stops.
func (reading allwiseReading) each(yield func(*nilgai.Source, error) bool) {
	rows := separated(reading.in)
	at, err := header(rows, allwiseColumns)
	if err == nil {
		err = bothColumnsOfEachBand(at)
	}
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

// bothColumnsOfEachBand returns an error for a search that selects only
// one of a band's two columns, the magnitude or its error. The error
// names the missing column. A search that selects neither has not asked
// for the band, and only W1 is required.
func bothColumnsOfEachBand(at map[string]int) error {
	for _, each := range allwiseBands {
		if each.band == W1 {
			continue
		}
		_, hasMagnitude := at[each.magnitude]
		_, hasSigma := at[each.sigma]
		switch {
		case hasMagnitude && !hasSigma:
			return noColumn(each.sigma)
		case hasSigma && !hasMagnitude:
			return noColumn(each.magnitude)
		}
	}
	return nil
}

// source converts one row to a Source, or returns nil, the column that
// could not be read, and why. When the nilgai record refuses the Source,
// the error names no column.
func (table AllWISE) source(
	cells map[string]string,
) (*nilgai.Source, string, error) {
	got, column, err := numbers(cells, allwiseNumbers...)
	if err != nil {
		return nil, column, err
	}
	position, err := nilgai.NewPosition(got["ra"], got["dec"])
	if err != nil {
		return nil, "ra", err
	}
	for _, column := range []string{"sigra", "sigdec"} {
		if !nilgai.NotBelowNone(got[column]) {
			return nil, column, fmt.Errorf("%s %v must be "+
				"finite and not below zero", column,
				got[column])
		}
	}
	ellipse, err := allwiseEllipse(got)
	if err != nil {
		return nil, "sigra", err
	}
	measurements, column, err := table.measurements(cells)
	if err != nil {
		return nil, column, err
	}
	var epoch nilgai.When // AllWISE states no epoch for its positions
	source, err := nilgai.NewSource(nilgai.AllWISE, cells["designation"],
		position, ellipse, epoch, measurements)
	return source, "", err
}

// measurements returns the row's brightness in each band, W1 first, with
// WISE's flags kept and read. If a column cannot be read it returns that
// column and why.
//
// W1 is always there. Another band is there when the search selected it
// and the row has a value in it.
//
// AllWISE leaves both a band's magnitude and its error empty when no
// profile-fit measurement was possible there. Its ph_qual is then X (AllWISE
// Explanatory Supplement, section II.1.a). A source with no measurement in
// the band is not a bad row, so the band is left out.
//
// measurements refuses a band with an error and no magnitude as a bad row.
func (table AllWISE) measurements(cells map[string]string) (
	made []nilgai.Measurement, column string, err error) {
	var unkept error
	for place, each := range allwiseBands {
		if each.band != W1 && absent(cells, each) {
			continue
		}
		magnitude, column, err := numbers(cells, each.magnitude)
		if err != nil {
			return nil, column, err
		}
		got, column, err := table.brightness(place,
			magnitude[each.magnitude], cells[each.sigma])
		if err != nil {
			return nil, column, err
		}
		read, failed := flagged(*got, cells, place)
		made = append(made, read...)
		unkept = errors.Join(unkept, failed)
	}
	return made, "", unkept
}

// allwiseEllipse returns the one-sigma ellipse for AllWISE's errors in RA
// and Dec and their co-sigma. The co-sigma squared, keeping its sign, is
// the covariance. The axes are the square roots of the covariance
// matrix's eigenvalues.
//
// allwiseEllipse turns the major axis's direction from an angle measured
// from the RA axis to the position angle, measured east of north.
func allwiseEllipse(got map[string]float64) (nilgai.ErrorEllipse, error) {
	alongRA, alongDec := got["sigra"], got["sigdec"]
	cosigma := got["sigradec"]
	covariance := cosigma * math.Abs(cosigma)
	raVariance, decVariance := alongRA*alongRA, alongDec*alongDec
	mean := nilgai.Half(raVariance + decVariance)
	reach := math.Hypot(nilgai.Half(raVariance-decVariance), covariance)
	major, minor := math.Sqrt(mean+reach), math.Sqrt(mean-reach)
	fromEast := nilgai.Half(math.Atan2(nilgai.Halves*covariance,
		raVariance-decVariance))
	eastOfNorth := nilgai.RightAngle -
		fromEast*nilgai.DegreesInHalfACircle/math.Pi
	angle := math.Mod(eastOfNorth+nilgai.FullCircle, nilgai.FullCircle)
	return nilgai.NewErrorEllipse(major, minor, angle, nilgai.OneSigma())
}

// absent says whether the row has no measurement in the band, because
// either the search did not select the band or both its cells are empty.
func absent(cells map[string]string, each allwiseBand) bool {
	magnitude, selected := cells[each.magnitude]
	return !selected || magnitude == "" && cells[each.sigma] == ""
}

// brightness returns the brightness in the band at the place in
// allwiseBands. It is a detection when AllWISE gives an error, and an
// upper limit when the error is empty, which is how AllWISE writes a
// limit.
//
// If a cell cannot be read, brightness returns nil and names which of the
// two columns it was.
func (table AllWISE) brightness(place int, magnitude float64,
	magnitudeError string) (*nilgai.Measurement, string, error) {
	each, edges := allwiseBands[place], table.edges[place]
	if magnitudeError == "" {
		limit, _, err := WiseFlux(each.band, magnitude, none)
		if err != nil {
			return nil, each.magnitude, err
		}
		measurement, err := nilgai.NewUpperLimit(each.filter,
			edges.first, edges.last, limit)
		return measurement, each.magnitude, err
	}
	spread, err := strconv.ParseFloat(magnitudeError, 64)
	if err != nil {
		return nil, each.sigma, err
	}
	if !nilgai.NotBelowNone(spread) {
		return nil, each.sigma, fmt.Errorf("magnitude error "+
			"%v must be finite and not below zero", spread)
	}
	flux, sigma, err := WiseFlux(each.band, magnitude, spread)
	if err != nil {
		return nil, each.magnitude, err
	}
	measurement, err := nilgai.NewDetection(each.filter, edges.first,
		edges.last, flux, sigma)
	return measurement, each.magnitude, err
}

// flagged returns the measurement with WISE's flags kept and read. The
// quality is doubtful when the band's character of cc_flags names an
// artefact, and good when it does not. The flags are kept whole, for every
// band.
func flagged(measurement nilgai.Measurement, cells map[string]string,
	place int) ([]nilgai.Measurement, error) {
	var flags []cell
	for _, name := range allwiseFlags {
		flags = append(flags, cell{name, cells[name]})
	}
	quality := nilgai.Good
	contamination := cells["cc_flags"]
	if len(contamination) > place &&
		contamination[place] != cleanContamination {
		quality = nilgai.Doubtful
	}
	return kept(measurement, quality, flags...)
}
