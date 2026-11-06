// Package nilgai normalizes astronomy catalogue rows into a common record for
// finding sources through the dust of the Milky Way’s plane.
//
// Chandra, AllWISE, and Fermi-LAT 4FGL readers are implemented in read. MeerKAT
// and Neowise have reader interfaces; their implementations are still to come.
// Nilgai supplies records for comparison, leaving source matching and
// classification to callers.
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
// The root package defines Source, Measurement, ErrorEllipse, Position, flux
// types, Quality, Flag, When, LightCurve, and RowError. The following packages
// import the record; the record does not depend on them:
//
//   - band maps filters onto libtheme bands and supplies false-colour
//     profiles. Cited filter edges and chosen colour mappings live in
//     band/data/filters.json and band/data/profiles.json.
//   - read parses Chandra TAP searches, AllWISE IRSA searches, and 4FGL
//     FITS tables. Invalid rows yield RowError values with their location.
//   - sky draws sources as one-sigma error ellipses. Strip arranges panels
//     side by side; sky/data/sky.json supplies the drawing style.
//
// The nilgai command’s show verb renders these searches as a PNG.
package nilgai
