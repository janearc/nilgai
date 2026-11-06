# Searching for planets and degenerate worlds formed in the Galactic Center

Briefing for the observatory team. 

Status: synthesis from a literature search plus the working conversation so far.
A few references are marked *verify on ADS* where I have the author and year
but have not yet finalized the exact citation; the detection-methods citations
carry arXiv IDs and are solid. Do not treat a *verify* flag as a confirmed
citation.

---

## 1. What we are looking for: two populations

The driving idea is that the Galactic Center's last active period, roughly the
last 9 million years, could have built objects that normal star-forming
regions cannot. Two kinds are on the table, with very different confidence.

### 1a. Blanets (planets around Sgr A* / in the AGN disk)

Planets formed in the accretion disk of a supermassive black hole, proposed by
Wada, Tsukamoto and Kokubo (2019, ApJ 886, 107; and the follow-up that names
them "blanets", 2021, ApJ 909, 96; *verify on ADS*). The predicted population:
roughly 20 to 3000 Earth masses, icy and rocky, forming beyond a sublimation
radius of a few parsecs, over timescales up to ~100 Myr, favored in
lower-luminosity AGN. The formation is contested: a radiative-barrier argument
(RATD) by Giang, Hoang et al. 2022 (arXiv:2111.11800) suggests grain growth to
planet scales is suppressed in AGN disks. Treat blanet formation as an
uncertain but active hypothesis.

### 1b. Degenerate rocky worlds (hydrogen/helium-poor solid bodies)

Speculation: in an H/He-poor, dust-rich environment, pebble accretion could
build rocky bodies approaching a solar mass. Such an object is supported by
electron degeneracy pressure, is roughly Earth-sized at ~1 solar mass (density
~1e5 to 1e6 g/cc), cannot fuse hydrogen, and is essentially an iron/rock
"world" with no star. There is no formation paper we have found; this remains
speculation. The mass limits are the familiar ones: iron-rich degenerate
matter becomes unstable to electron capture near ~1.06 solar masses, while
carbon/oxygen matter holds to the Chandrasekhar limit near ~1.4 solar masses
(*verify exact numbers on ADS*).

The AGN-only-forms-these argument is what makes this search clean: if such a
body exists, its host or its orbital history points at the nucleus, so a
grossly anomalous star on a Galactic-Center-traced orbit is a detection, not a
coincidence.

---

## 2. Why the ejected population is the observable

The core itself cannot be searched in the optical or near-infrared for planets.
Measured extinction toward Sgr A*: A_V ~ 38-42 mag (Gao & Li 2013,
arXiv:1305.7137); the near-IR is workable but still deep (A_Ks = 2.42,
A_H = 4.21; Fritz et al. 2011, arXiv:1105.2822); only the radio is essentially
unobscured. The optical attenuation toward the center is enormous, so there is
a genuine zone of avoidance.

The escape is to look for stars that were ejected from the region and are now
in clean, low-extinction fields. A ~1000 km/s star covers ~1 kpc per Myr, so a
star found 5-15 kpc out has a flight time of a few to ~10 Myr, comfortably
inside the 9 Myr window that brackets the active episodes (the event that may
have inflated the Fermi bubbles, a few Myr ago, and a more recent flare).

The observable is therefore: ejected stars, kinematically traced back to a
close approach to Sgr A*, examined for chemical or transient fingerprints of
having encountered one of the objects above.

---

## 3. Detection channels, ranked for actionability

### 3a. Ejected stars that engulfed a planet (the "unusual spectra" route)

This is the most immediately actionable channel because the technique is now
demonstrated, not theoretical.

- Main-sequence planet ingestion is a real, named technique. Melendez 2009
  (ApJL 691, L47; *verify*) showed the Sun is depleted in refractory elements
  relative to solar twins, interpreted as planet formation sequestering them.
  Spina et al. 2021 (Nature Astronomy; arXiv:2108.12040) found a wide-binary
  pair where one star is refractory-rich and lithium-rich versus its twin,
  read as a swallowed planet. Liu et al. 2024 (arXiv:2403.13209) sharpened the
  same kind of signature.
- White dwarf pollution is the established workhorse: a quarter to half of
  cool white dwarfs show heavy metals from accreted planetary material
  (Zuckerman, Koester, and the follow-up literature; *verify*).
- Lithium-rich giants are partly attributed to engulfed planets or brown
  dwarfs (Aguilera-Gomez et al. 2016, Nature Astronomy; *verify*), and those
  surveys are already public.
- Caught in the act: a star visibly swallowing a planet has been observed
  (ZTF SLRN-2020; De et al. 2023, Nature). Merger transients in general follow
  the luminous-red-nova template (V1309 Sco; Tylenda 2011; *verify*).

Expected fingerprint for a blanet: blanets are icy and rocky, so the pollution
signature is volatile (C, N, O) plus refractory enrichment plus lithium, not
metals alone. The caveat that matters most: this reads cleanly in a cool star
with a deep convection zone, and poorly in hot stars, yet the published
ejected/hypervelocity sample is dominated by B and A stars (a selection effect,
they are the bright ones). So push the cross-match toward late-type and, above
all, white dwarf candidates.

### 3b. Degenerate-body merger (the smoking gun)

If a hostile solar-mass iron/rock body merges with a star, the signature is not
subtle. Mixing even a sub-solar-mass degenerate body into a solar-mass star
makes the surface tens of percent iron and rock, versus the ~1.4 percent cosmic
metallicity. No stellar evolution produces that, and an abundance pipeline
(APOGEE, GALAH) would flag it or fail on it. One spectrum settles it.

The physical outcome splits by mass and composition, and the split sets the
search:

1. Body stays sub-critical and is shredded in the envelope or swallowed to the
   core without collapsing: the star survives as an extreme, metal-saturated
   object. If the core ends as a degenerate kernel under a surviving hydrogen
   envelope, that is a Thorne-Zytkow object with a rocky core instead of a
   neutron star core; the published TZO signature is enhanced rubidium,
   molybdenum and lithium in a cool luminous giant (Levesque et al. 2014,
   MNRAS; *verify*), with HV 2112 in the SMC as their candidate.
2. Body reaches the core and tips past the mass limit: it collapses by
   electron capture or accretion-induced collapse toward a neutron star,
   leaving a faint transient and no star behind to find.
3. The plunge itself is bright before either outcome: a cool, red luminous
   transient, days to months, fainter than a supernova, brighter than a nova.

So the "find a persistent star" channel favors the sub-collapse mass case, and
the discriminator to steal is the TZO rubidium-molybdenum-lithium test added to
an absurd-metallicity test.

### 3c. Microlensing rogue planets toward the bulge (mature, and has a gap)

Free-floating planetary-mass objects are best counted by microlensing toward
the bulge. Sumi et al. 2011 (Nature) first claimed Jupiter-mass free-floating
planets; Mroz et al. 2017 (Nature) contradicted that with a strict limit; Sumi
et al. 2023 (arXiv:2303.08280) settled on a bottom-heavy (Earth-to-Neptune)
mass function, roughly 21 free-floating planets per star, with no
Jupiter-mass population. The first directly measured free-floating planet mass,
0.219 Jupiter masses, is Dong et al. 2025 (Science; arXiv:2601.00057).

The implication for us: if blanets are Jupiter-to-ten-Jupiter-mass objects,
they would stand out as an excess precisely where the current census finds
nothing. The catch is that no Jovian-free-floating census exists yet; it is the
ongoing near-IR bulge surveys (PRIME, arXiv:2508.14474; KMTNet full-frame,
arXiv:2503.24097; DREAMS, arXiv:2601.11469 / 2605.27364) that will build one.
This is a larger, separate project but is the most "physics already built"
channel.

### 3d. Direct and astrometric detection near Sgr A*

GRAVITY/VLTI delivers tens-of-microarcsecond astrometry and has measured the
S2 orbit precisely enough to see Schwarzschild precession (GRAVITY 2020,
arXiv:2004.07187; review Eisenhauer & Monnier 2023, arXiv:2303.00453), but
there is no published planet-search campaign or companion-mass limit on the
nuclear S-stars. That is a genuine gap, and forward-looking: ELT/METIS imaging
of the inner arcsecond is proposed but has no published sensitivity paper. The
near-term truth is that direct detection of a planet near Sgr A* is not yet
on the table; astrometric wobble is a future channel.

### 3e. The bodies themselves, without a star

Two handles exist for free-flying degenerate/solar-mass bodies:

- MACHO-style microlensing: a dark ~0.5-1 solar-mass object in the bulge
  lenses a background star into a months-long event with no luminous
  counterpart at the lens position. This is the modern compact-object search,
  and the event catalogs already exist.
- An extreme-mass-ratio inspiral (EMRI): a solar-mass compact body inspiraling
  into Sgr A* is exactly the LISA-band source class the detector is being built
  to hear. This is a decade-scale channel but a clean one.

There are also tidal-disruption proposals specific to blanets; the published
rate is discouraging (Pandey & Singh 2026, arXiv:2606.06884, gives a
century-long, UV-peaked, very low-rate signal), so do not prioritize it.

---

## 4. What to pull, and how to build the candidate list

For the ejected-star candidate list (the heart of 3a and 3b):

1. Gaia DR3 astrometry (positions, proper motions, parallax), plus the
   Bailer-Jones photogeometric distance catalog (parallax is unusable at these
   distances) and the RVS radial velocities where present. The source is the
   ESA archive, which the request already has access to.
2. Radial velocities and abundances from spectroscopy, to close the 6D
   phase-space vector and to test chemistry: APOGEE DR17 (near-IR, reads
   refractories and CNO through the dust), GALAH DR4 (optical, reads lithium,
   southern coverage, which is where the center sits), LAMOST (cheap radial
   velocities for fainter stars), and Gaia-ESO. This is also where the
   engulfment and TZO discriminators get measured.
3. Orbit integration backward in time: a Milky Way potential (galpy is the
   standard), then integrate and Monte Carlo over the parallax, proper-motion
   and radial-velocity errors to get a probability that a star's pericenter
   came within a few parsecs of Sgr A* within the last 9 Myr.
4. Do not re-derive the population from scratch: start from the published
   hypervelocity and runaway candidate catalogs (Brown 2015, ARA&A 53, *verify*;
   the Gaia DR2/DR3 surveys by Marchetti, Boubert and others, *exact IDs
   pending*).
5. For the white dwarf channel: the Gaia white dwarf catalog (Gentile Fusillo
   et al.; *verify*), crossed with the same kinematics, to find candidate
   ejected white dwarfs, then check for metal pollution.
6. Reddening: Bayestar19 or the Galactic-Center-specific map (Surot et al.;
   *verify*) to define the true edge of the zone of avoidance.

---

## 5. First screening action

Take the kinematically validated ejected list and, for each star, ask one gross
question rather than doing a full abundance analysis: does this spectrum fit
any stellar model, and is its iron abundance sane?

Triggers to flag a star for a dedicated spectrum:
- unreasonably high [Fe/H] or a pipeline that fails to converge on it;
- enhanced lithium;
- enhanced rubidium and molybdenum (the TZO test), especially in a cool,
  luminous source;
- near-IR/WISE excess or an unusual broad or composite spectrum.

A star that is even a fraction iron-dominated will not hide; the cost is one
spectrum per candidate.

---

## 6. Open gaps and risks

- Blanet formation is contested (RATD barrier); the entire 3a chain rests on
  the formation being real.
- The degenerate-rocky-world formation step has no published support; section
  3b is detection logic, not established theory.
- There is no Jupiter-to-super-Jupiter free-floating planet census yet; the
  current microlensing result cuts against a Jovian population, which either
  kills 3c or makes a Jovian excess the cleanest imaginable signal, depending
  on what blanets actually weigh.
- No S-star planet-search limit exists, so 3d is unquantified.
- The ejected/hypervelocity sample is B/A-biased, which is the wrong spectral
  class for the chemical tests; the faint, late-type end is under-sampled and is
  where the signal lives.

## 7. References (bold = arXiv ID confirmed by search)

Blanets / AGN planets:
- Wada, Tsukamoto & Kokubo 2019, ApJ 886, 107; and 2021, ApJ 909, 96 (blanets). *verify*
- Giang, Hoang et al. 2022 (RATD barrier), arXiv:2111.11800.

Detection methods / extinction:
- Fritz et al. 2011, arXiv:1105.2822 (GC extinction A_Ks etc.).
- Nishiyama et al. 2008, arXiv:0802.3559 (extinction law toward GC).
- Gao & Li 2013, arXiv:1305.7137 (A_V toward Sgr A*).
- Eisenhauer & Monnier 2023, arXiv:2303.00453 (GRAVITY review).
- GRAVITY Collaboration 2020, arXiv:2004.07187 (S2 Schwarzschild precession).
- Haggard et al. 2024, arXiv:2408.03467 (sub-arcsec extinction variability).

Microlensing free-floating planets:
- Sumi et al. 2011, Nature 473, 349. *verify*
- Mroz et al. 2017, Nature 548, 183. *verify*
- Sumi et al. 2023, arXiv:2303.08280 (MOA-II mass function).
- Dong et al. 2025, Science, arXiv:2601.00057 (first measured FFP mass).
- PRIME, arXiv:2508.14474; KMTNet, arXiv:2503.24097; DREAMS, arXiv:2601.11469 and 2605.27364.

Engulfment / merger signatures:
- Melendez et al. 2009, ApJL 691, L47. *verify*
- Spina et al. 2021, Nature Astronomy, arXiv:2108.12040.
- Liu et al. 2024, arXiv:2403.13209.
- Aguilera-Gomez et al. 2016, Nature Astronomy (Li-rich giants). *verify*
- De et al. 2023, Nature 617, 55 (ZTF SLRN-2020 star engulfs planet). *verify*
- Tylenda et al. 2011 (V1309 Sco merger transient). *verify*
- Levesque et al. 2014, MNRAS (Thorne-Zytkow object search, HV 2112). *verify*

Blanet tidal disruption:
- Pandey & Singh 2026, arXiv:2606.06884.
