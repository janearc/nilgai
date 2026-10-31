// Package read holds nilgai's catalogue readers: Chandra, AllWISE and
// Fermi. Each reads one catalogue's file and yields a nilgai.Source for
// each row, or a *nilgai.RowError for a row it cannot read.
//
// Package read also holds what the AllWISE reader needs to convert a WISE
// Vega magnitude to a flux density: WiseFlux, and the zero points in
// data/wise-zero-points.json.
//
// Unassociated filters what Fermi reads. It keeps the sources near the
// Galactic plane that 4FGL has not associated with any known object.
//
// testdata/4fgl-rows.fit holds seven rows of the real 4FGL file. The
// program in testdata/mkfermi writes it and says how to run it. That
// program is not part of the library.
package read

// none is the count of nothing, zero.
const none = 0
