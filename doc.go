// Package nilgai reads four catalogues of sources seen through the dust
// of the Milky Way's centre and hands back one common record for each.
//
// The catalogues are MeerKAT (radio, 1.28 GHz), Fermi-LAT 4FGL (gamma
// rays), the Chandra Source Catalog (X-rays) and Neowise (mid-infrared,
// 3.4 and 4.6 microns, many looks over years).
//
// The package reads and normalises. It does not match, score or decide;
// matching belongs to a later library.
//
// Quantities carry their unit in their type. A flux density in Jy
// (FluxDensity) and an energy flux in erg/s/cm^2 (EnergyFlux) are
// different types, and a Measurement refuses a value and an error that
// disagree.
//
// An upper limit is a Kind of Measurement, not a special
// number. Every value that was converted keeps the catalogue's own
// value beside it.
//
// This version has the types, the conversions and the reader interfaces.
// It has no reader yet.
package nilgai
