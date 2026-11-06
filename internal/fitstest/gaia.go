package fitstest

import (
	"io"

	"github.com/astrogo/fitsio"
)

// GaiaStar is a made-up star as a Gaia DR3 product of stars with a
// radial velocity holds it, for tests that read one. Each value is
// nothing until it is set.
type GaiaStar struct {
	id                                         int64
	ra, dec, parallax, pmRA, pmDec             float64
	raError, decError, parallaxError           float32
	pmRAError, pmDecError, radial, radialError float32
	raDec, raParallax, decParallax             float32
}

// NewGaiaStar is a star with the id given and every other value nothing.
func NewGaiaStar(id int64) *GaiaStar { return &GaiaStar{id: id} }

// At sets where the star is on the sky, in degrees.
func (star *GaiaStar) At(ascension, declination float64) {
	star.ra, star.dec = ascension, declination
}

// Distant sets the star's parallax and its error, in milliarcseconds.
func (star *GaiaStar) Distant(parallax, parallaxError float64) {
	star.parallax, star.parallaxError = parallax, float32(parallaxError)
}

// Moving sets the star's proper motion, in milliarcseconds a year.
func (star *GaiaStar) Moving(ascension, declination float64) {
	star.pmRA, star.pmDec = ascension, declination
}

// Receding sets the star's radial velocity and its error, in kilometres
// per second.
func (star *GaiaStar) Receding(radial, radialError float64) {
	star.radial, star.radialError = float32(radial), float32(radialError)
}

// Erring sets the errors of the star's place, in milliarcseconds, and of
// its proper motion, in milliarcseconds a year.
func (star *GaiaStar) Erring(ascension, declination, motionAscension,
	motionDeclination float64) {
	star.raError, star.decError = float32(ascension), float32(declination)
	star.pmRAError = float32(motionAscension)
	star.pmDecError = float32(motionDeclination)
}

// Tied sets three of the star's correlations: its ascension with its
// declination and with its parallax, and its declination with its
// parallax.
func (star *GaiaStar) Tied(ascensionDeclination, ascensionParallax,
	declinationParallax float64) {
	star.raDec = float32(ascensionDeclination)
	star.raParallax = float32(ascensionParallax)
	star.decParallax = float32(declinationParallax)
}

// values is the star as a row of the product: a value for each column
// in order, every column not set nothing of its kind.
func (star GaiaStar) values() []any {
	set := map[string]any{"source_id": star.id, "ra": star.ra,
		"dec": star.dec, "parallax": star.parallax, "pmra": star.pmRA,
		"pmdec": star.pmDec, "ra_error": star.raError,
		"dec_error": star.decError, "parallax_error": star.parallaxError,
		"pmra_error": star.pmRAError, "pmdec_error": star.pmDecError,
		"radial_velocity":       star.radial,
		"radial_velocity_error": star.radialError,
		"ra_dec_corr":           star.raDec, "ra_parallax_corr": star.raParallax,
		"dec_parallax_corr": star.decParallax}
	var values []any
	for _, column := range gaiaColumns() {
		value, given := set[column.Name]
		if !given {
			value = nothingOf[column.Format]
		}
		values = append(values, value)
	}
	return values
}

// nothingOf is nothing in each of the product's kinds of column, by
// form.
var nothingOf = map[string]any{"K": int64(nothing),
	"D": float64(nothing), "E": float32(nothing), "I": int16(nothing)}

// nothing is no value of any kind.
const nothing = 0 // defined: zero

// gaiaColumns are the product's columns as the archive writes them.
func gaiaColumns() []fitsio.Column {
	return []fitsio.Column{{Name: "source_id", Format: "K"},
		{Name: "ref_epoch", Format: "D", Unit: "yr"},
		{Name: "ra", Format: "D", Unit: "deg"},
		{Name: "ra_error", Format: "E", Unit: "mas"},
		{Name: "dec", Format: "D", Unit: "deg"},
		{Name: "dec_error", Format: "E", Unit: "mas"},
		{Name: "parallax", Format: "D", Unit: "mas"},
		{Name: "parallax_error", Format: "E", Unit: "mas"},
		{Name: "pmra", Format: "D", Unit: "mas.yr**-1"},
		{Name: "pmra_error", Format: "E", Unit: "mas.yr**-1"},
		{Name: "pmdec", Format: "D", Unit: "mas.yr**-1"},
		{Name: "pmdec_error", Format: "E", Unit: "mas.yr**-1"},
		{Name: "radial_velocity", Format: "E", Unit: "km.s**-1"},
		{Name: "radial_velocity_error", Format: "E", Unit: "km.s**-1"},
		{Name: "ra_dec_corr", Format: "E"},
		{Name: "ra_parallax_corr", Format: "E"},
		{Name: "ra_pmra_corr", Format: "E"},
		{Name: "ra_pmdec_corr", Format: "E"},
		{Name: "dec_parallax_corr", Format: "E"},
		{Name: "dec_pmra_corr", Format: "E"},
		{Name: "dec_pmdec_corr", Format: "E"},
		{Name: "parallax_pmra_corr", Format: "E"},
		{Name: "parallax_pmdec_corr", Format: "E"},
		{Name: "pmra_pmdec_corr", Format: "E"},
		{Name: "ruwe", Format: "E"},
		{Name: "phot_g_mean_mag", Format: "E", Unit: "mag"},
		{Name: "bp_rp", Format: "E", Unit: "mag"},
		{Name: "nu_eff_used_in_astrometry", Format: "E", Unit: "um**-1"},
		{Name: "pseudocolour", Format: "E", Unit: "um**-1"},
		{Name: "ecl_lat", Format: "D", Unit: "deg"},
		{Name: "astrometric_params_solved", Format: "I"}}
}

// WriteGaia writes a product of the stars given, as the archive writes
// one: a table named gaia_source of its columns, a row a star.
func WriteGaia(out io.Writer, stars ...*GaiaStar) error {
	var rows [][]any
	for _, star := range stars {
		rows = append(rows, star.values())
	}
	return writeRows(out, "gaia_source", gaiaColumns(), rows)
}
