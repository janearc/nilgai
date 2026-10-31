# nilgai

nilgai: hindi nīl (blue) and gāy (cow), a nilgai is a large blue-grey antelope
native to asia, sharing the grasslands of india, nepal and pakistan.
supposedly there is a feral herd in texas, but we don't talk about that.

additionall, nilgai is a golang library. it exists to read five separate
astronomy catalogues:

- MeerKAT, radio, 1.28 GHz
- Fermi-LAT 4FGL, gamma rays
- Chandra Source Catalog, X-rays
- AllWISE, mid-infrared at 3.4, 4.6, 12 and 22 microns (W1 to W4), one row for
  each source
- Neowise, mid-infrared at 3.4 and 4.6 microns, with many looks over years

for the express purpose of looking across the 'avoidance zone' of the plane of
the milky way to find previously undocumented high energy xray and neutrino
sources (which is to say, blazars). 

each catalogue states positions, errors, brightness and quality flags in
its own way. nilgai normalizes to the same units, so a later library can
compare a gamma-ray source and an infrared dot as the same kind of thing.

basically, i'm looking for little blue dots (this is a joke, i am hilarious)
in the hardest place in the universe to actually look for things. it's fine.

## what it does not do

this library cannot determine what *is* an AGN, but it serves to find things
that might be by correlating various data sources.

## the record

- `Source`: one catalogue row. the catalogue's own identifier, a position
  (ICRS right ascension and declination in degrees), an error ellipse, and one
  or more `Measurement`s.
- `ErrorEllipse`: semi-axes at one sigma in arcseconds, position angle in
  degrees east of north. It also keeps the confidence the catalogue stated
  (Fermi and Chandra state 95%; MeerKAT states one sigma) and the catalogue's
  own axes, so the conversion can be checked.
- nilgai has no band type. one slice of the spectrum is libtheme's
  `bands.Band`, from my
  [libtheme](https://github.com/janearc/libtheme-css/primitives/bands) package.
  `band.BandsBetween` turns a catalogue filter's two edges in nm into the first
  and last libtheme bands it covers. Fermi's range runs from 50 MeV to 1 TeV.
  (more details on this in the libtheme source)
- `Measurement`: the brightness through one filter, with the catalogue's name
  for the filter (such as "W1") and the libtheme bands it covers. The unit is in
  the type: `FluxDensity` is Jy and `EnergyFlux` is erg/s/cm^2. Mixing them is
  refused. An upper limit is its own kind. The catalogue's flags are kept
  verbatim beside nilgai's reading of them.
- `LightCurve`: the Neowise looks at one position, in time order.

each conversion is a named function that cites where its numbers come
from: `RadiusFactor` and `ToOneSigma` for ellipse confidence,
`read.WiseFlux` for WISE Vega magnitudes (zero points from the WISE
explanatory supplement, section IV.4.h, after Jarrett et al. 2011, ApJ 735,
112).
units, constants and conversions between energy, frequency and
wavelength all come from libtheme (primitives/si, primitives/length and
spaces/radiation); nilgai defines none of its own. wavelengths are plain
nanometres, as libtheme has them.

## other stuff

`nilgai` is built on [puzzlemouth](https://github.com/janearc/puzzlemouth).

so far there are readers for chandra, allwise and fermi: `read.Chandra`,
`read.AllWISE` and `read.Fermi`. meerkat and neowise are still to come.
`nilgai show` draws their searches around a position as a png.

`read.Unassociated` picks out of a fermi read the sources that 4FGL has not
associated with anything (a blank class), within so many degrees of the
galactic plane. `Position.GalacticLatitude` is the latitude it asks.

the record is the root package. the rest is split by what it does, and each
imports the record, never the other way round:

- `band/`: the bands, the filters and the false colour profiles.
- `read/`: the readers.
- `sky/`: the picture, `Sky`, `Draw` and `Strip`.

## building

```sh
$ game build
$ game check
```

## author

`jane michelle arc <jmich@fastmail.nl>`
