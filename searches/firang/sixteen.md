# firang: sixteen very fast stars

Thrown from Sgr A*, or bad radial velocities? A brief for an
independent reviewer, written 2026-11-06.

Everything here comes from a scratch program written the same morning.
It is unreviewed, and you are the review. Write anything you make in
your own scratch, not in this repository, and delete nothing.

## The question

Sixteen stars in Gaia DR3 move 630 to 833 km/s through the galaxy and
lie within about 2.5 kpc of the Sun, toward the galactic centre. Each
traces back toward Sgr A*'s neighbourhood but misses it. Are they stars
thrown from Sgr A*, or is something wrong with their data?

The reading under review is that their radial velocities are bad, not
that they were thrown. Please try to break it. The parts trusted least
are listed under "Where to push".

## The data

- The product is Gaia DR3 gaia_source, every row with a radial
  velocity: 33,812,183 rows, 31 columns, on this machine; "Reproduce"
  says how to find it.
- celestial-data's `gaia/catalogue.go` names the columns. The file
  has no radial-velocity quality columns but the error.
- A star is kept when its parallax is positive and at least five times
  its error, and it has a radial velocity and a proper motion:
  29,947,046 stars.
- The sixteen, one line each with every number below, are in
  `sixteen.tsv` beside this file.

## The method

1. **Screen.** Every kept star is placed in celestial's published
   galactocentric frame and run back ten million years through
   MWPotential2014 (Bovy 2015) with Sgr A* as a point mass, on
   celestial's Metal kernel: kick-drift-kick, 32-bit, 10,000 steps.
2. The kernel records each star's nearest pass by Sgr A*. Each star
   runs twice, once per estimate of the Sun's speed along the rotation:
   ```
   frame (Sgr A*'s apparent motion, reflected)   251.548 km/s
   model (MWPotential2014 at the Sun, plus the
          Sun's own motion)                       231.476 km/s
   ```
   57,532 stars pass within 1 kpc under either speed.
3. **Draws.** Each of those 57,532 runs back 100 more times. Each draw
   takes the five astrometric parameters from their covariance, built
   from the errors and the ten correlations and drawn through a
   Cholesky factor.
4. In each draw the radial velocity is drawn from its own error, and
   the Sun's speed along the rotation uniformly between the two
   estimates. The Sun's other two components, Sgr A*'s distance, the
   Sun's height and the potential are held fixed. The seed is 1.
5. **Selection of the sixteen.** From the 57,532: speed through the
   galaxy over 600 km/s at central values with the frame's Sun,
   distance (one over the parallax) under 3 kpc, and parallax at least
   ten times its error. The line was drawn after seeing the list.

## The control

A star really thrown from Sgr A* has to be recognisable with errors
like these. To check that, 20,000 stars were thrown from 1 pc off Sgr
A*, straight out, at 500 to 1500 km/s. Their flights were 0.5 to 9 Myr,
rounded to a tenth of a Myr.

- The kernel carried each one forward: a back-run of the reversed
  motion, with the end state's motion reversed again.
- Each was seen from the Sun through the frame turned round. The
  frame's axes were found by placing stars along the three equatorial
  axes, and the motion was solved from three probes of the frame.
- Each borrowed the full covariance and radial-velocity error of a
  random screened star, and got one draw of that noise.
- The same quality floor was applied, keeping 10,392. Each then ran
  through the same 100 draws.

Run back from their true values, the thrown stars pass 0.23 pc from
Sgr A* at the median, and 0.42 pc at the 90th percentile. So the carry
and the seeing are right to within the kernel's step.

## What the numbers say

Thrown stars seen within 3 kpc of the Sun with parallax at least ten
times its error, 147 of them:
```
share of draws within 100 pc, mean                  0.42
stars with at least a fifth of draws within 100 pc  109 of 147
median of each star's median pass                   119 pc
```

The sixteen:
```
share of draws within 100 pc                0.00, every one
share of draws within 1 kpc, mean           0.88
median of each star's median pass           643 pc
each star's median pass                     239 to 1084 pc
widest 16th-to-84th spread of a star        217 pc
```
Their draws are tight around a miss. The miss is not a smear of large
errors.

Their properties:
```
speed through the galaxy   630 to 833 km/s, almost all outward
radial velocity            -868 to -641 km/s, errors 6.8 to 21.5
G magnitude                13.97 to 15.73
galactic longitude         -1.1 to 22.2 degrees
galactic latitude          -5.4 to +4.5 degrees
distance                   0.33 to 2.50 kpc
RUWE                       0.92 to 4.50; 4102928386097611392 is
                           4.50, 4107082822155354496 is 4.02
```

The population they come from is every kept star with a radial
velocity of 600 km/s or more either way, in `rv-tails.tsv` beside this
file (which lists 500 km/s and more). "Toward the bulge" means within
30 degrees of the centre in longitude and 15 in latitude.
```
                     negative   positive   most within 3 kpc of us
toward the bulge        91        114      yes
elsewhere              145        206      yes
mean G                 about 15 in every group
```

A star near us that moves toward the centre at over 600 km/s cannot
have come from it. Huge radial velocities in both signs, all over the
sky, in faint stars close to us, look like a measurement population.

The sixteen are the part of it whose negative sign, toward the bulge,
happens to point their paths back near Sgr A*.

## Since, in SIMBAD and the surveys

Looked up by Gaia DR3 id on 2026-11-06. Two of the sixteen are in
SIMBAD, each named by one paper: 4090667766374029184 as a candidate
eclipsing binary, and 4069681937636687232 as a young stellar object.
Both are kinds of star a single radial velocity can mislead on.

None of the sixteen is among the stars SIMBAD ties to two papers on
fast Gaia DR3 stars: 2023ApJ...944L..39L (three stars) and
2023RNAAS...7..267P (fifteen brighter than G 16 with radial velocities
below -600 km/s). The two picked other stars than these.

None is in APOGEE DR17's allStar or GALAH DR4's allstar either, so no
second radial velocity was found for any of them.

## Where each claim comes from

- Measured in these runs: every number above but in "Since".
- Read on 2026-11-06: SIMBAD's ids, object types and paper titles,
  and the two survey files.
- Recalled, not read: that Gaia DR3 has spurious large radial
  velocities for faint stars, worst in crowded fields. Also that
  hundreds of real stars this fast this close would far exceed the
  known hypervelocity stars. Treat both as hypotheses.

## Where to push

1. **The control's errors.** The control borrows radial-velocity
   errors from the screened stars, about 2.7 km/s on average. The
   sixteen's are 6.8 to 21.5 km/s. Rerun the control borrowing only
   from stars like the sixteen.
2. The draws already include each star's own error, and the sixteen
   still miss outside their spread, but check that this holds.
3. **The selection.** The cut was drawn after looking. Does the
   conclusion survive other cuts?
4. **The Sun.** The Sun's speed range moves a pass by roughly the 20
   km/s difference times the flight, about 180 pc over 9 Myr. It
   applies to the control and the sixteen alike. Is it the right
   uncertainty to draw?
5. **The potential.** Only MWPotential2014 was tried. Would a
   different bulge or bar bend these paths onto Sgr A*?
6. **The covariance.** In the scratch program's `covariance()` and
   `noisy()`, ra_error is taken as the error of right ascension times
   the cosine of the declination, in mas, and converted to an angle
   through the cosine. Check the units and the conversion.
7. **The decisive check, not yet done.** Gaia DR3's own
   radial-velocity quality columns for these sixteen: the number of
   transits, the template temperature, the fit statistics. The column
   names are recalled, and the product here does not carry them.
8. Fetching those columns is an archive query, which waits on the
   observatory owner's yes.

## Reproduce

The scratch program is one Go package in a scratch workspace on this
machine, built on celestial's galaxy branch at 11570f6 and
celestial-data's gaia branch at 9d1ba2a. It needs macOS for the Metal
kernel, and each full run holds the GPU for about a minute and a half.

Two paths on this machine are not written here: the scratch workspace
that holds the program, and the Gaia product. Whoever handed you this
brief has both; set them as `work` and `product`.

```sh
cd $work/draws
GOWORK=$work/go.work go build -o YOUR_FOLDER/draws .
YOUR_FOLDER/draws -gaia $product -out YOUR_FOLDER            # screen, draws
YOUR_FOLDER/draws -gaia $product -out YOUR_FOLDER -tails 500 # the tails
YOUR_FOLDER/draws -gaia $product -out YOUR_FOLDER \
  -inject 20000 -lenders YOUR_FOLDER/screen.tsv -seed 2      # the control
```

Run one heavy job at a time. The screen's and the draws' full output
(`draws.tsv`, 9 MB) and the control's (`injected.tsv`, 1 MB) are not
kept in the repository; the commands above make them again from the
same seeds.
