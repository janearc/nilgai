// Package read parses Chandra, AllWISE, and Fermi catalogues into nilgai
// records. Each reader yields a Source per valid row or a RowError naming an
// invalid row. A row error allows reading to continue; an input-level failure
// ends the sequence.
//
// WiseFlux converts WISE Vega magnitudes to flux densities using the zero
// points in data/wise-zero-points.json. Unassociated filters Fermi sources by a
// blank CLASS1 and distance from the Galactic plane.
//
// testdata/4fgl-rows.fit contains seven rows from the real 4FGL catalogue. The
// fixture generator in testdata/mkfermi documents how to rebuild it; it is
// separate from the library.
package read

// none is the zero count.
const none = 0
