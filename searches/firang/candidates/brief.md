# firang: seven fast stars that trace back to Sgr A*

A brief on firang's candidates, written 2026-11-06. It is unreviewed,
and a reader is welcome to be the review. Write anything you make in
your own scratch, not in this repository, and delete nothing.

## The question

Seven stars in Gaia DR3 move faster than 400 km/s through the galaxy,
outward, and a fifth or more of their paths, drawn across their
measurement errors, pass within 100 pc of Sgr A*. Were they thrown from
Sgr A*? If so, their spectra are where the search reads them next.

Five are candidates and have folders here. Two are set aside: their
radial velocities are faint stars' huge ones, with errors past the
rule's 5 km/s, suspect for the reason [sixteen.md](../sixteen.md)
gives. One of the two is already published as a hypervelocity star.

## The run

- The tool is `cmd/firang` in this search, the verbs `draws` and
  `candidates`. The method and the control are the ones
  [sixteen.md](../sixteen.md) describes, with one change: 10,000
  draws per star, not 100.
- All 57,532 screened stars ran at 10,000 draws each, seed 1. At a
  share of 0.2, counting 10,000 draws holds it to about 0.004.
- The rule, from `numbers.yaml`: speed through the galaxy now of 400
  km/s or more, a share within 100 pc of 0.2 or more, and a radial
  velocity error of 5 km/s or less.
- The same seven clear the speed and the share at 10,000 draws as at
  100, and no other star does. Of 996 screened stars at 400 km/s or
  more, 18 sit between 0.1 and 0.2.
- `seven.tsv` beside this file holds the seven's rows, every column
  the draws verb writes. The folders still carry the first pass's
  figures at 100 draws.

## What a thrown star looks like here

The validation threw 20,000 stars from 1 pc off Sgr A*, at 500 to
1500 km/s, and ran the survivors through the same draws at 100 each.
Of the 4,192 kept at 3 to 7 kpc from the Sun, where the seven are, 771
reach a share of 0.2: 18%.

So a star really thrown from Sgr A*, seen from here with errors like
these, usually misses the cut. One that clears it is consistent with
having been thrown, not shown to have been. How often a fast star that
was not thrown clears it has not been measured.

## The seven

Shares are of draws passing within 100 pc of Sgr A*, at 10,000 draws,
with the first pass's share at 100 in brackets.

```
4094201527955913856   set aside, radial velocity error past the rule
place             ra 272.85252, dec -20.51619 degrees
distance          3.61 kpc, parallax 10.1 times its error
radial velocity   -677.2 +- 10.90 km/s
G, RUWE           14.59, 1.13
from the centre   4.77 kpc now, moving 659 km/s, 658 outward
share             0.442 (0.37)
nearest pass      median 109 pc, 16th to 84th percentile 71 to 190

4037334576111139712   candidate
place             ra 269.30578, dec -36.54110 degrees
distance          5.72 kpc, parallax 9.4 times its error
radial velocity   -423.1 +- 1.88 km/s
G, RUWE           12.44, 0.68
from the centre   2.73 kpc now, moving 467 km/s, 466 outward
share             0.383 (0.40)
nearest pass      median 123 pc, 16th to 84th percentile 65 to 227

4081136684118965760   candidate
place             ra 287.50086, dec -22.23481 degrees
distance          5.06 kpc, parallax 10.1 times its error
radial velocity   -366.9 +- 1.59 km/s
G, RUWE           13.01, 1.04
from the centre   3.93 kpc now, moving 405 km/s, 403 outward
share             0.319 (0.34)
nearest pass      median 163 pc, 16th to 84th percentile 53 to 374

6736906690592556672   candidate
place             ra 281.02536, dec -32.06512 degrees
distance          5.17 kpc, parallax 9.1 times its error
radial velocity   -377.4 +- 5.00 km/s
G, RUWE           14.00, 0.87
from the centre   3.44 kpc now, moving 401 km/s, 400 outward
share             0.305 (0.36)
nearest pass      median 158 pc, 16th to 84th percentile 69 to 342

4385671099714045824   candidate
place             ra 251.24756, dec 2.00623 degrees
distance          3.54 kpc, parallax 19.6 times its error
radial velocity   -381.5 +- 3.22 km/s
G, RUWE           13.20, 1.00
from the centre   5.71 kpc now, moving 487 km/s, 486 outward
share             0.288 (0.25)
nearest pass      median 168 pc, 16th to 84th percentile 63 to 373

4066702325525092736   set aside, radial velocity error past the rule
place             ra 272.21569, dec -22.70660 degrees
distance          3.90 kpc, parallax 6.3 times its error
radial velocity   -691.9 +- 6.77 km/s
G, RUWE           15.59, 0.92
from the centre   4.45 kpc now, moving 679 km/s, 678 outward
share             0.285 (0.24)
nearest pass      median 179 pc, 16th to 84th percentile 57 to 333

6760416924931944832   candidate, on the line
place             ra 283.58454, dec -30.67717 degrees
distance          5.91 kpc, parallax 12.1 times its error
radial velocity   -305.1 +- 1.01 km/s
G, RUWE           12.31, 0.93
from the centre   2.98 kpc now, moving 402 km/s, 400 outward
share             0.202 (0.17)
nearest pass      median 252 pc, 16th to 84th percentile 79 to 458
```

## In the literature

SIMBAD was asked on 2026-11-06 for each of the seven by its Gaia DR3
id, and for the five it did not know, by their DR2 and EDR3 ids too.
Only the titles below were read; none of the papers has been.

**4094201527955913856** is a candidate eclipsing binary in SIMBAD, and
five papers name it:

- 2023ApJ...944L..39L, "Hypervelocity Stars Track Back to the Galactic
  Center in Gaia DR3".
- 2023RNAAS...7..267P, fifteen hypervelocity stars brighter than G 16
  with radial velocities below -600 km/s.
- 2023MNRAS.525..561E, constraints on the Galactic centre from Gaia
  hypervelocity stars, on a possible companion to Sgr A*.
- 2025ApJ...984...56A, the Galactic bar's and the LMC's effect on
  hypervelocity stars' paths.
- 2023A&A...674A..16M, Gaia DR3's catalogue of eclipsing-binary
  candidates.

So the tool found on its own a star others traced back to the centre.
The eclipsing-binary flag is the doubt: a binary's own orbit moves its
radial velocity, and one measurement of it can be far off.

**4081136684118965760** is UCAC4 339-178585 in SIMBAD. Four papers
name it: the APOGEE data releases in SDSS DR16 (2020AJ....160..120J)
and DR17 (2022ApJS..259...35A), APOGEE Net (2022AJ....163..152S), and
a paper on the outermost halo's metallicity (2020ApJ...894...34D).

That suggests APOGEE observed it, so a spectrum and abundances may
already exist. They have not been looked at.

**The other five** are not in SIMBAD. SIMBAD does not index every
large catalogue's table, so this says only that no paper it knows of
names them.

SIMBAD links three stars to 2023ApJ...944L..39L and fifteen to
2023RNAAS...7..267P. Of those, only 4094201527955913856 is among
firang's 57,532 screened stars, and none is among
[sixteen.md](../sixteen.md)'s sixteen.

## What the spectroscopic surveys hold

Each of the seven was looked up by its Gaia DR3 id on 2026-11-06 in
GALAH DR4's allstar and APOGEE DR17's allStar, through celestial-data's
readers of those files. GALAH holds none of them. APOGEE holds one.

**4081136684118965760** is 2M19100020-2214052 in APOGEE, observed from
Las Campanas. Every element below has a flag of 0.

```
radial velocity   -364.8 km/s, its visits scattering 0.1 km/s
                  (Gaia: -366.9 +- 1.59 km/s)
temperature       4823 +- 11 K
gravity           log g 1.98 +- 0.03
iron              [Fe/H] -0.81 +- 0.01, [M/H] -0.82
fit               ASPCAPFLAG, STARFLAG and EXTRATARG all 0; S/N 179
[X/Fe]            O +0.31  Ca +0.16  K +0.11  N +0.11  Si +0.09
                  Mg +0.06  S +0.06  TiII +0.04  Ti +0.03  Na -0.01
                  V -0.02  Ni -0.02  Co -0.06  Ce -0.08  Cr -0.10
                  CI -0.15  Al -0.21  C -0.24  Mn -0.25
```

So a second instrument confirms its radial velocity, and APOGEE saw no
binary in its visits. It is a red giant, metal-poor, and these numbers
show no excess of the refractory elements over the volatile ones.
APOGEE's H band holds no lithium line.

A reading, recalled and not read: stars born at the Galactic Centre
are mostly metal-rich, and a star at [Fe/H] -0.8 with magnesium near
iron looks more like a halo star accreted from another galaxy. Its
fourth paper in SIMBAD is on the outer halo's metallicity.

That paper, Dietz, Yoon, Beers and Placco 2020 (arXiv:1911.11140),
takes its metallicities from SDSS spectra and SkyMapper photometry, by
a search's summary of it; it was not read.

**Lithium** for it needs a spectrum at 6708 A, and none was found on
2026-11-06. GALAH DR4 lacks the star; APOGEE's band and Gaia's RVS
window hold no lithium line. ESO's archive and Gemini's hold nothing
within 5 arcseconds of it, and a web search found no paper.

ESO's archive holds the Gaia-ESO Survey and every VLT spectrograph's
products; the same query finds Proxima Centauri's HARPS spectra, so it
works. One new optical spectrum at high resolution would answer it.

None of the sixteen in [sixteen.md](../sixteen.md) is in either
survey.

## Where each claim comes from

- Measured in these runs: every number in "The run", "What a thrown
  star looks like here" and "The seven".
- Read on 2026-11-06: SIMBAD's ids, object types and paper titles;
  every number in "What the spectroscopic surveys hold", from the
  survey files; the archives' answers.
- Inferred, then confirmed by the file: that a reference to the APOGEE
  data releases means APOGEE observed the star.
- Recalled, not read: that Gaia DR3's parallax zero point is a large
  part of parallaxes this small, and that its formal errors are
  underestimated; that Galactic Centre stars are mostly metal-rich, and
  what low magnesium at [Fe/H] -0.8 says. Treat them as hypotheses.

## Where to push

1. **No control for stars not thrown.** The validation shows how
   often a thrown star clears the rule, not how often an ordinary fast
   star does. Without that, a share of 0.3 is not evidence of origin.
2. **The parallax.** The draws use Gaia's formal errors only. The
   seven's parallaxes are 0.17 to 0.28 mas. Apply DR3's zero point
   and an error inflation, and see whether the shares hold.
3. **The radial velocities.** Gaia DR3's own quality columns for the
   six APOGEE has not confirmed, and the eclipsing-binary flag for
   4094201527955913856. Fetching them is an archive query, which waits
   on the observatory owner's yes.
4. **4094201527955913856 in the papers.** Read 2023ApJ...944L..39L
   and compare its traceback with this one's.
5. **The rule.** It was picked with the first pass. 6760416924931944832
   clears it at 0.202. 4111098272950981632 sits at 0.186 with a radial
   velocity error of 5.08 km/s, just past the rule.
6. **The Sun and the potential**, as [sixteen.md](../sixteen.md) puts
   them: the Sun's speed range, and MWPotential2014 alone.
7. **4081136684118965760's chemistry.** Is a red giant at [Fe/H] -0.81
   with [Mg/Fe] +0.06 and [Al/Fe] -0.21 a star of the Galactic Centre
   at all? Compare it with the centre's stars and the halo's in the
   literature before reading its lithium.

## Reproduce

The tool builds only inside a workspace holding celestial's galaxy
branch at 2760660 and celestial-data's gaia branch at eb0d25b, until
both are released. It needs macOS for the Metal kernel. Whoever handed
you this brief has the workspace and the Gaia product.

The `screen` target in this search's `.game` makes `data/screen.tsv`.
Then, from `searches/firang`, with a copy of `numbers.yaml` whose
`draws` is 10000:

```sh
go run ./cmd/firang draws --numbers YOUR_FOLDER/numbers.yaml \
  --gaia data/gaia/gaia_source_0_0_r180.fits \
  --screened data/screen.tsv --into YOUR_FOLDER/draws.tsv
go run ./cmd/firang candidates --numbers YOUR_FOLDER/numbers.yaml \
  --drawn YOUR_FOLDER/draws.tsv --into YOUR_FOLDER/candidates
```

The draws held the GPU for 16.5 minutes and 11 GB of memory at peak.
Run one heavy job at a time.
