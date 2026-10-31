// Package nilgai reads catalogues of sources seen through the dust of the
// Milky Way's centre and hands back one common record for each.
//
// Package nilgai reads these catalogues:
//   - MeerKAT, radio at 1.28 GHz
//   - Fermi-LAT 4FGL, gamma rays
//   - the Chandra Source Catalog, X-rays
//   - AllWISE, mid-infrared in W1 to W4 at 3.4, 4.6, 12 and 22 microns,
//     with one row for each source
//   - Neowise, mid-infrared at 3.4 and 4.6 microns, with many looks over
//     years
//
// Package nilgai reads and normalises. It does not match, score or decide.
// Matching belongs to a later library.
//
// Quantities carry their unit in their type, held in SI: FluxDensity and
// EnergyFlux are different types, which Jansky and ErgPerSecPerCm2 make
// from the units catalogues give. A Measurement's value and error are one
// type, so a pair that disagrees does not compile.
//
// Kind marks a Measurement as an upper limit; a limit is never a special
// number. Every converted value keeps the catalogue's own beside it.
//
// # Packages
//
// Package nilgai is the record. It holds Source, Measurement, ErrorEllipse,
// Position, the flux types, Quality, Flag, When, LightCurve and RowError.
// Three packages beside it import it, and it imports none of them:
//   - band holds the bands and false-colour profiles, and the edges of
//     the filters that readers label a flux with. The filters are in
//     band/data/filters.json, with their citations and whether they are
//     approximate. The profiles are in band/data/profiles.json.
//   - read holds the readers. Three catalogues have them: read.Chandra
//     reads a Chandra Source Catalog search, read.AllWISE reads an
//     AllWISE search from IRSA, and read.Fermi reads the 4FGL FITS file.
//     MeerKAT and Neowise have reader interfaces and no reader yet.
//   - sky draws what the readers return: each source as its one-sigma
//     error ellipse around a centre, in an ink for its instrument. Strip
//     puts several skies side by side. The colours are in
//     sky/data/sky.json. The nilgai command's show verb draws searches
//     this way as a png.
package nilgai
