# nilgai

A nilgai—from Hindi nīl (blue) and gāy (cow)—is a large blue-grey
antelope native to Asia, sharing the grasslands of India, Nepal, and
Pakistan. Supposedly there is a feral herd in Texas, but we don’t talk
about that.

Nilgai is also a Go library. I’m building it to look through the Milky
Way’s zone of avoidance for previously undocumented high-energy sources,
particularly blazar candidates. Radio, gamma-ray, X-ray, and infrared
catalogues each tell part of that story.

Basically, I’m looking for little blue dots (this is a joke, I am
hilarious) in the hardest place in the universe to actually look for
things. It’s fine.

## What it does

Nilgai converts catalogue rows into a common record: positions, positional
uncertainties, brightness, and quality flags in consistent units. That
gives a caller something useful to compare across instruments. Matching
sources and deciding whether a candidate is an AGN remain separate work.

| Catalogue | Coverage | Reader |
|---|---|---|
| Chandra Source Catalog | X-rays | `read.Chandra`, TAP text export |
| AllWISE | Mid-infrared, W1–W4: 3.4, 4.6, 12, and 22 µm | `read.AllWISE`, IRSA CSV export |
| Fermi-LAT 4FGL | Gamma rays | `read.Fermi`, FITS point-source table |
| MeerKAT | Radio, 1.28 GHz | Interface defined; reader planned |
| Neowise | Repeated mid-infrared observations at 3.4 and 4.6 µm | Interface defined; reader planned |

Readers yield a source or a `*nilgai.RowError` for each row. A row error
names the input, data row, and column, then reading continues. An error
that prevents reading the input is yielded once and ends the sequence.
Readers consume the supplied stream; callers open the files.

AllWISE requires W1 and can also read W2–W4. Optional bands need both
magnitude and uncertainty columns. A populated magnitude with an empty
uncertainty is an upper limit; an optional band with both cells empty
has no measurement and is omitted.

## The record

- `Source` holds the catalogue’s identifier, an ICRS position in degrees,
  an error ellipse, the position epoch when known, and measurements.
- `ErrorEllipse` stores one-sigma semi-axes in arcseconds and a position
  angle in degrees east of north. It also retains the catalogue’s stated
  axes and confidence level so the conversion can be checked.
- `Measurement` records a detection or upper limit through a named filter,
  such as W1, and the libtheme bands it covers. `FluxDensity` is in Jy;
  `EnergyFlux` is in erg/s/cm². Values and uncertainties must use the same
  unit. Original value text and catalogue flags can be retained alongside
  Nilgai’s quality assessment.
- `When` stores an MJD or an unknown time. `LightCurve` groups measurements
  at one position in time order, ready for the planned Neowise reader.

`RadiusFactor` and `ToOneSigma` convert positional confidence levels.
`read.WiseFlux` converts WISE Vega magnitudes using zero points cited in
the WISE Explanatory Supplement, IV.4.h, after Jarrett et al. 2011,
ApJ 735, 112. Its symmetric flux uncertainty is a first-order approximation.

Units and spectral conversions come from
[libtheme](https://github.com/janearc/libtheme-css). Wavelengths are in
nanometres. `band.BandsBetween` maps a filter’s bounds onto libtheme’s
`bands.Band`; Nilgai has no separate spectral band type. Fermi’s configured
colour profile spans 50 MeV to 1 TeV; the Energy_Flux100 measurement
filter covers 100 MeV to 100 GeV.

## Packages and command

The root package defines the record. Three packages build on it:

- `band` holds filter edges, citations, and false-colour profiles.
- `read` parses catalogues and converts their values.
- `sky` draws one-sigma error ellipses and arranges panels with `Strip`.

`read.Unassociated` selects Fermi sources with a blank CLASS1 within an
exclusive latitude bound around the Galactic plane. Sources classed unk
or bcu are excluded by this selection. Errors pass through unchanged.

The command uses [puzzlemouth](https://github.com/janearc/puzzlemouth).
`nilgai show` draws selected searches around a position as a PNG:

```sh
nilgai show --chandra chandra.tsv --allwise allwise.csv \
  --ra 266.4168371 --dec -29.0078106 --width 12 --pixels 600 \
  > sgra.png
```

Add `--fermi catalogue.fit` to include 4FGL. A width list such as
`--width 12,6` draws panels side by side. Each source has its instrument’s
colour, adjusted to the reader’s contrast requirements through libreadme.
Unreadable rows are reported on standard error after the image is written;
the image still contains the valid sources, and the command exits with an error.

## Building and checking

Use the Go version in `go.mod`, then run:

```sh
game build
game check
```

Filter edges, colour mappings, and drawing styles live in JSON data files
beside their packages. Measured bounds carry citations; chosen display
colours remain identified as choices. Tests include small catalogue
fixtures, real exported rows, and property checks for ellipse geometry.

## Author

Jane Michelle Arc · jmich@fastmail.nl
