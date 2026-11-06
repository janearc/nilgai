# changelog

## v0.0.4 2026-11-06

- created `searches` component
- added first operational search, *`firang`*, which attempts to find evidence
  of planets formed in the active galactic nucleus phase of the MW during the
  period the fermi bubbles were created
- added support for gaia data (from `celestial-data`, separate release, ships
  with [celestial](https://github.com/janearc/celestial)).

## previous releases

since v0.0.1, which was the record types and nothing that read a
catalogue.

added:

- `nilgai orbits`: the fastest stars of a Gaia product, as celestial-data
  fetches it, run back through celestial's Milky Way on the GPU to
  their nearest pass by Sgr A*, one line a star sorted by that pass
- the stars kept by parallax over error, the share by speed against
  the Sun, the span, the steps and the batch are each a flag with its
  default and reason in `cmd/nilgai/orbits.yaml`
- `--sun frame` or `--sun model` is required: celestial's published
  frame and its MWPotential2014 give the Sun speeds 20 km/s apart
  along the rotation, and which the stars carry is a choice, not a
  default
- a star the frame refuses or whose orbit runs away is left out and
  named in the output's header. the solver names its runaways in a
  typed refusal (celestial's gpu/orbit Runaway), and the rest of the
  batch runs again without them
- any other refusal of the solver, a closed solver, a device gone,
  stops the run with its words and nothing written
- the share is counted by nearest integer, a half rounding up, and is
  at least one star: of ten rows, `--fastest 0.7` selects seven
- a row whose speed is not a number, a proper motion the archive left
  empty beside a radial velocity, is not kept
- the command builds on darwin only, since the kernel is Metal's
- `searches/`: a folder per question nilgai asks of the sky. the first
  is firang, stars thrown out of the galactic centre, read for the
  blanets they may have eaten there; its first pass, unreviewed, is
  written down beside it
- firang carries its own tool, `cmd/firang`, with a verb for each
  stage: screen, draws, validate and candidates, run by targets in the
  search's `.game` from the numbers in its `numbers.yaml`
- the tool reads Gaia sources through celestial-data's reader, draws
  them with its Spread and As, places them with its State, and sees
  thrown stars with celestial's galactocentric Seen; until those are
  released it builds in the same go workspace
- `internal/fitstest` writes a made-up Gaia product from GaiaStar
  builders; the orbits verb's tests and firang's share it
- the orbits verb places a source with celestial-data's State, the
  placing it had written for itself
- nilgai requires celestial-data now, for the Gaia reader; until
  celestial and celestial-data are released with it, the module
  builds in a go workspace with their branches
- three packages beside the record:
  - `band`: libtheme bands for a catalogue's filter, the measurement
    bands in data/filters.json, and the false-colour profiles in
    data/profiles.json, one libtheme `radiation.Profile` per instrument.
  - `read`: readers for a Chandra Source Catalog search (TSV), an
    AllWISE search (CSV) and the Fermi-LAT 4FGL catalogue (FITS).
  - `sky`: draws sources on an aqua ground, in inks chosen by measured
    contrast through libreadme, with a mark at a position the caller
    gives.
- `read.NewChandra(profiles)`, `read.NewAllWISE()` and `read.NewFermi()`
  return a reader, or nil and the error. `read.ReadAllWISE(filters)` and
  `read.ReadFermi(filters)` take the filter edges as data.
- AllWISE reads W1, W2, W3 and W4. `read.WiseBand` names all four, and
  `read.WiseZeroPoints` holds a cited zero point for each.
- `read.Unassociated(sources, limit)` keeps the 4FGL sources with a blank
  CLASS1 that lie within limit degrees of the Galactic plane. Rows that
  could not be read come through as their errors.
- `nilgai.Position.GalacticLatitude()`.
- `band.ProfileFile`, `band.ProfileEntry` and `band.Quantity`:
  data/profiles.json as a named file type, each band's ends written as
  the paper writes them ("1 keV", "84 GHz"). `band.NewProfileFile(data)`
  reads one, and refuses a key the file does not have.
- `band.ProfileSource`: what a caller needs from the profiles.
  `band.NewEmbedded()` is the one built into nilgai. `MockProfileSource`
  in mocks stands in for it in tests.
- `band.FilterBands(name)` and `band.ReadFilterBands(data, name)`.
- `sky.NewSky`, `sky.ReadSky` and `sky.Strip`.
- `nilgai.Finite`, `nilgai.Positive`, `nilgai.NotBelowNone`,
  `nilgai.Half`, and the angle constants (`ArcsecondsPerDegree`,
  `FullCircle`, `RightAngle` and the rest), defined once for every
  package.
- `nilgai show` draws searches as a png on standard output:
  - it reads `--chandra`, `--allwise` and `--fermi`.
  - it needs `--ra`, `--dec`, `--width` and `--pixels`.
  - each `--width` is a panel, so a list of widths shows a zoom.
  - rows that cannot be read are reported together, after the picture.

changed, breaking:

- `NewSource` and `NewLightCurve` return a pointer, and nil on a
  refusal.
- every reader's `Read` yields `*Source`, and nil beside a refused row.
- `BandsBetween` moved to `band.BandsBetween`.
- `WiseFlux`, `WiseBand` and `WiseZeroPoints` moved to `read`.
- the `version` verb is gone. pzm's `--age` says which build it is.
