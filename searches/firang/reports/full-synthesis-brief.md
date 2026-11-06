# Searching for planets and degenerate worlds formed in the Galactic Center

Full synthesis brief for the observatory team. Compiled 2026-11-06.

Status labels used throughout:

| Label | Meaning |
|---|---|
| established / demonstrated | measured in real data, or long-standing textbook physics |
| proposed | published model, simulation, or projected capability; no detection |
| contested | live peer-reviewed disagreement about whether the process works |
| speculative / no literature | premise has no published model or none found in the search |

Headline position: blanet *formation* is a small, contested theoretical
literature with zero detections. A hydrogen/helium-poor degenerate rocky world
near a solar mass has **one formation model**, worked for black holes heavier
than Sgr A* (section 1b), and its structure and its collapse limits are
standard white-dwarf physics. The observable we can act
on is the ejected-star population, and the actionable channel is unusual
surface chemistry in stars kinematically traced back to Sgr A*. The
"smoking gun" merger channel is well-posed physically but the expected signal
is much less extreme than the earlier draft assumed.

---

## 1. The two target populations

### 1a. Blanets: planets formed in an AGN disc

Proposed by Wada, Tsukamoto and Kokubo. The 2019 paper (ApJ 886, 107;
arXiv:1909.06748, plus corrigendum ApJ 907, 58) models sub-micron icy grains
growing through porous aggregates to decoupling sizes of 10 to 100 m, then a
gravitationally unstable dust layer fragmenting into planets of about 10 Earth
masses beyond the snow line, several parsecs from the black hole. The 2021
paper (ApJ 909, 96; arXiv:2007.15198) coins "blanet" and adds dust radial
advection, which raises predicted masses to **20 to 3000 Earth masses**
(roughly 6e-5 to 0.01 solar masses) inside r < 4 pc, for disc turbulence
alpha < 0.04 and a 1e6 solar-mass black hole, in about 70 to 80 Myr.

Key properties for our purposes, all model output, not observation:

- composition is **icy (water), not rock**, because formation is outside the AGN
  snow line (arXiv:1909.06748 full text; fiducial r_snow about 1.4 pc at
  1e6 solar masses, r_sub about 0.3 pc in the 1e7 solar-mass case at Eddington
  ratio 0.01)
- orbits are parsec-scale, separations about 1e-3 to 1e-2 pc, around a black
  hole rather than a star
- favored only around low-luminosity AGN (L_bol about 1e42 erg/s) with lifetimes
  under about 1e8 yr, and best at M_BH about 1e6 solar masses
  (arXiv:2007.15198). Sgr A* is 4.1e6 solar masses and today far below that
  luminosity, so the present-day center is a marginal host and any population
  would be a fossil of a past active phase
- the maximum predicted mass, 3000 Earth masses, is about 9.4 Jupiter masses, or
  roughly 1 percent of a solar mass. Blanets as published are **not degenerate**
  and are more than two orders of magnitude below a solar mass

Status: **proposed, and contested**. Giang, Hoang, Tram et al. (2022, ApJ 936,
108; arXiv:2111.11800) argue radiative torques spin up and disrupt large grains
and strip icy mantles, on timescales up to about 1e4 times shorter than grain
growth. Their conclusion is that growth survives only in a smooth torus and is
strongly constrained in a clumpy torus, which is the physically favored
geometry. This is the principal published challenge and it is unresolved.

The field is thin. OpenCitations and Crossref counts put roughly 20 citers on
the 2019 paper and roughly 9 on the 2021 paper through early 2026, mostly
astrobiology or AGN dust rather than dynamics. No dedicated blanet orbital
evolution, migration, or detection-signature paper was found. The closest
observational connection is Zhou, Zhang, Huang and Lin (2026, A&A 705, A58),
who fit the observed Sgr A* flare distribution with tidal disruption of
kilometer-scale rubble-pile bodies, a model match (predicted power-law index
1.83 against observed 1.65 to 1.9), not a detection.

### 1b. Hydrogen/helium-poor degenerate rocky worlds

This is the more exotic target, and the honesty check matters: **no published
work forms a near-solar-mass solid rocky body by pebble or solid accretion in an
H/He-poor environment.** The search found neither a formation paper nor a paper
that rules it out. The literature gap is explicit in
`degenerate_compact_objects.md` and is the single largest weakness in the
program's premise.

What is solid is the structure physics, which is textbook white-dwarf physics:

- a cold degenerate body obeys R proportional to M^(-1/3); a about 1 solar mass
  body is Earth-sized, roughly 5000 to 10000 km, with central density 1e6 to
  1e9 g/cm3 (Hamada and Salpeter 1961, ApJ 134, 683)
- the ideal Chandrasekhar limit is about 1.44 solar masses for mean molecular
  weight per electron mu_e = 2, which covers He, C, O and approximately Fe
  (Chandrasekhar 1931, ApJ 74, 81)
- before that limit, **electron capture truncates the sequence by composition**:
  iron at about **1.06 solar masses** (Hamada and Salpeter 1961), ONeMg at about
  **1.37 solar masses** via capture on 24Mg at 4e9 g/cm3 and on 20Ne at 2e10
  g/cm3 (Miyaji et al. 1980; Nomoto 1984; Hashimoto et al. 1993), while C/O is
  limited by carbon ignition rather than neutronization
- the maximum mass of a *solid planet* in the actual planet literature is about
  20 Earth masses (Seager et al. 2007, arXiv:0707.2895) with pure-rock radius
  models reaching 10 Jupiter masses (Fortney et al. 2007,
  arXiv:astro-ph/0612671); carbon planets cap near 60 Earth masses
  (Kuchner and Seager 2005, arXiv:astro-ph/0504214) and are themselves
  questionable (Nissen et al. 2014 find host C/O never above 0.8,
  arXiv:1406.5218; Deming and Seager 2017 list carbon planets among
  unconfirmed results, arXiv:1701.00493)

There is a genuine unmodeled band of about two orders of magnitude between the
top of solid-planet models and the bottom of the cold-sphere literature
(Zapolsky and Salpeter 1969, ApJ 158, 809). The phrase "degenerate planet"
returns zero arXiv hits.

Mishra, Lyra, McKernan, Mac Low, Ford and Cook (2026, ApJ 1005, 99;
arXiv:2605.19241) form them. Streaming instability in an AGN dust torus makes
filaments holding solar masses of dust, which collapse into tens of millions
of bodies from Earth to super-Jupiter masses.

Some are "directly formed above the hydrogen burning limit, yet of pure dust"
(their abstract), and "despite the high masses, these objects are pure dust"
(their section 5). The torus is gas-rich and gas accretion runs alongside:
some bodies reach the crossover mass and grow by gas toward stars.

But in their Figure 5 a body's dust accretion outruns its gas accretion more
than a thousandfold, at 1e7 to 1e9 solar masses and either fragmentation speed.
Our reading, not theirs: the pure-dust bodies stay mostly dust while pebbles
feed them. An earlier version of this brief said the paper forms no H/He-poor
bodies; it does.

Two limits: the model is worked only for black holes of 1e7 to 1e9 solar
masses, not Sgr A*'s 4.1e6, and its counts fall with mass, to about 2e5
planets at 1e7 (their Figure 6). Read 2026-11-06, pages 1 to 3 and 12 to 13
of 16; the copy in sources/ is arXiv v2, the draft as submitted.

In standard discs round young stars the ceiling is far
lower: pebble isolation mass is order 10 Earth masses (Bitsch et al. 2018,
arXiv:1801.02341; Chametla et al. 2022 reach tens to hundreds of Earth masses
only in high-viscosity discs, arXiv:2112.14214), and the critical core mass for
runaway gas accretion is 5 to 10 Earth masses (Piso and Youdin 2013,
arXiv:1301.1727). Burrows and Ostriker (2014, arXiv:1401.1814) make the same
point by dimensional analysis: a solar-mass self-gravitating body of rock or
iron is degenerate, not a solid planet.

Status: **structure established; formation proposed by one model**, worked for
black holes heavier than Sgr A*.

### 1c. Why the search stays clean even so

If an H/He-poor degenerate body exists, no normal star-forming channel explains
it, so a chemically or kinematically anomalous star points at the nucleus by
default rather than by coincidence. That logic is what makes a small candidate
list worth spectroscopically following up. It does not make the population
likely, and the null rate is a real risk (section 7).

---

## 2. Why the ejected-star population is the observable

The center itself is not searchable for these objects in the optical, and only
marginally in the near-infrared.

Demonstrated extinction toward Sgr A*:

- A_V about 38 to 42 mag from infrared extinction modeling (Gao and Li 2013,
  arXiv:1305.7137). Literature values span about 30 to 42 depending on method,
  so quote the range, not a single number
- A_H = 4.21 +/- 0.10, A_Ks = 2.42 +/- 0.10, A_L' = 1.09 +/- 0.13 measured on a
  direct line of sight using the VLA 2 cm continuum as the extinction-free
  reference, i.e. radio is unabsorbed (Fritz et al. 2011, arXiv:1105.2822)
- the extinction law is steep and non-standard: A(V):A(J):A(H):A(Ks) =
  1:0.188:0.108:0.062, so Ks is about 1/16 of visual (Nishiyama et al. 2008,
  arXiv:0802.3559), flat/gray from 3 to 8 microns (Chen et al. 2018,
  arXiv:1804.06807; von Fellenberg et al. 2025, arXiv:2511.14850)
- extinction is variable on sub-arcsecond scales near the central parsec
  (patchy dust screen, cross-section about 500 au; Haggard et al. 2024,
  arXiv:2408.03467), which is a first-order systematic for any time-domain or
  photometric search there

Practical consequence: visual light is dimmed by a factor of order 1e15 to 1e17
(Gao and Li 2013 quote an equivalent transmission of 1e-15 to 1e-17 in V), K band
by about 10x, radio not at all. A planet claim in the central parsec is NIR-only work
(demonstrated by UKIRT-2017-BLG-001Lb, detectable only in the near-infrared at
0.35 deg from the GC; Shvartzvald et al. 2018, arXiv:1802.06795), and no transit
survey reaches it at all.

What leaves, and how fast. In the Hills mechanism a tight binary that passes
close to the 4e6 solar-mass black hole is tidally broken up, one member captured,
the other flung outward at up to about 4000 km/s, far above the local escape
speed of roughly 500 to 600 km/s (Hills 1988, Nature 331, 687; Brown 2015, ARA&A
53, 15; local escape speed trending to about 600 km/s, implying a virial mass of
about 1.4e12 solar masses, Hattori et al. 2018, arXiv:1805.03194). So the fastest
ejecta are cleanly unbound and travel nearly straight, while slower ejections
form a bound runaway tail. Ejection rate is model-dependent and not directly
measured: about 1.5e-6 per year over the last 200 Myr for the B-star population
(Brown et al. 2014, arXiv:1401.7342) against an upper bound of about 8e-5 per
year from the Gaia DR3 search (Marchetti et al. 2022, arXiv:2206.06962).

So we look at what left. A star ejected at about 1000 km/s covers about 1 kpc
per Myr (1 km/s for 1 Myr = 1.02 pc), so at 5 to 15 kpc its flight time is about
5 to 15 Myr. The empirical anchor is S5-HVS1, a 2.35 solar-mass A-type star at
about 9 kpc with Galactic-frame velocity 1755 +/- 50 km/s, kicked at about
1800 km/s and traveling **4.8 Myr** from Sgr A* (Koposov et al. 2020,
arXiv:1907.11725; a GC origin over a random disc origin is favored with Bayes
factor K = 81 photometric and K = 354 spectro-photometric). A few-Myr to 9-Myr
lookback therefore selects ejecta currently within
roughly 5 to 9 kpc at about 1000 km/s, comfortably inside the halo volume where
the B-star hypervelocity population is found at 50 to 120 kpc.

Caveat to state plainly in any proposal: **no source in the corpus states a
"last 9 Myr" window.** The figure is assembled from the flight-time scaling plus
the S5-HVS1 anchor, and from the premise that the center's last active period
was about 9 Myr ago. Treat it as a program assumption to be justified, not as a
cited result.

Two further facts shape the design:

1. "Hypervelocity" is not one channel. Hill-mechanism ejections from Sgr A* are
   B/A stars (Hills 1988, Nature 331, 687; Brown 2015 review, ARA&A 53, 15).
   Hypervelocity white dwarfs D6-1, D6-2 and D6-3 at 1000 to 3000 km/s are
   surviving SN Ia companions (Shen et al. 2018, arXiv:1804.11163), as is the
   helium hot subdwarf US 708 at about 1200 km/s, whose GC origin is explicitly
   excluded (Geier et al. 2015, Science 347, 1126). PB 3877 is a wide sdB+K
   binary whose ejection has no known mechanism (Nemeth et al. 2016,
   arXiv:1604.03158). A credible "came from the core" list must exclude the
   supernova and satellite-debris channels by **origin inference**, not speed.
2. The catalogued B/A bias is a selection effect, not a property of the Hills
   mechanism (Brown et al. 2009 arXiv:0808.2469 targeted early-A/late-B colors;
   the complete MMT survey's 21 unbound stars are all 2.5 to 4 solar-mass
   main-sequence late-B stars at 50 to 120 kpc, Brown et al. 2014,
   arXiv:1401.7342). The chemical tests in section 3 want cool stars and white
   dwarfs, which is exactly the under-sampled end.

---

## 3. Detection channels, ranked for actionability

### 3a. Ejected stars that engulfed a planet: unusual spectra (rank 1)

This is the only channel where the technique is **demonstrated** rather than
proposed, and it runs on data already in public surveys.

The signature is a refractory-to-volatile abundance offset that correlates with
condensation temperature, plus lithium. Demonstrated instances:

- the Sun is depleted about 20 percent (about 0.08 dex) in refractories relative
  to volatiles compared with 11 solar twins, and the difference correlates with
  Tcond (Melendez et al. 2009, ApJL 704, L66; arXiv:0909.2299)
- Kronos and Krios, a 4 Gyr wide pair at about 100 pc: Kronos is enhanced about
  0.2 dex in elements with Tcond > 1200 K with only marginal C, N, O, Na, Mn
  enhancement, plus A(Li) = 2.75, about 0.5 dex above its twin, read as accretion
  of about 15 Earth masses of rock (Oh et al. 2018, ApJ 854, 138;
  arXiv:1709.05344). A 26-element re-analysis gives Delta[Fe/H] about 0.230 dex,
  Li excess 0.56 dex, engulfed mass about 27.8 Earth masses, and rules out
  diffusion or primordial inhomogeneity (Miquelarena et al. 2024, A&A 688, A73;
  arXiv:2406.06705)
- statistically, engulfment occurs in 20 to 35 percent of Sun-like stars from 107
  binaries (Spina et al. 2021, Nature Astronomy 5, 1163; arXiv:2108.12040), and
  at least one in a dozen from 91 co-natal pairs with 8 percent occurrence
  (Liu et al. 2024, Nature 627, 601; arXiv:2403.13209)
- lithium in giants: companions under 15 Jupiter masses dissolved in the
  convective envelope can raise A(Li) up to a ceiling of about 2.2, and 1 to
  3 percent of low-mass giants should exceed A(Li) = 1.5 (Aguilera-Gomez,
  Chaname, Pinsonneault and Carlberg 2016, ApJ 829, 127; arXiv:1605.05332), with
  Li I 6707.8 angstrom the diagnostic. Survey context: 1418 Li-bearing giants
  from GALAH, Li-rich fraction about 2x higher in the red clump (Castro-Tapia,
  Aguilera-Gomez and Chaname 2024, A&A 690, A367; arXiv:2401.00049)
- white dwarf pollution is the most established chemical tell of all: 25 percent
  of DA white dwarfs show Ca II (Zuckerman et al. 2003, ApJ 596, 477), an
  unbiased HST/COS survey finds at least 27 and possibly up to about 50 percent
  accreting planetary debris (Koester, Gaensicke and Farihi 2014, A&A 566, A34;
  arXiv:1404.2617), measured ratios match bulk Earth with Mg and Fe silicates
  (Gaensicke et al. 2012, MNRAS 424, 333), GD 362's 17 elements look like the
  Earth-Moon system (Zuckerman et al. 2007, ApJ 671, 872), and core-like Fe, Ni,
  S excess is seen in PG 0843+516. Warm dusty debris discs surround about 4
  percent of white dwarfs (Vanderburg et al. 2015, Nature 526, 546;
  arXiv:1510.06387), and WD 1145+017 is caught mid-accretion with 4.5 to 4.9 h
  transits up to 40 percent deep
- caught in the act: ZTF SLRN-2020, an infrared transient from a star engulfing a
  planet (De et al. 2023, Nature 617, 55; **no arXiv preprint**, DOI
  10.1038/s41586-023-05842-x). A 0.8 to 1.5 solar-mass star swallowed a 1 to
  10 Jupiter-mass planet; mid-infrared dust brightening began about 7 months
  before optical outburst; optical luminosity about 1e35 erg/s (only about 26
  L_sun, subluminous), total radiated about 6.5e41 erg, optical lasting about
  100 days with about 15 months of mid-IR emission, about 33 Earth masses of
  hydrogen and 0.33 Earth masses of dust ejected. Galaxy rate about 0.1 to
  several per year. JWST follow-up revises the progenitor to a about 0.7 solar
  mass K-type star not yet off the main sequence, so engulfment by orbital decay
  rather than stellar expansion (Lau et al. 2025, ApJ 983, 87;
  arXiv:2504.07275)

What we would actually look for, given the target physics. A blanet is icy, so
the expected fingerprint is volatile (C, N, O) *plus* refractory enrichment plus
lithium, not the refractory-only pattern of a rocky body. A merged
hydrogen/helium-poor rocky or iron body gives refractories with suppressed
hydrogen. The two are distinguishable in principle with one good spectrum.

Two honest limits on this channel.

First, dilution timescales. A 10 Earth-mass engulfment signature stays above
about 0.05 dex for only about 90 Myr in a 1 solar-mass star (Behmard et al.
2023, MNRAS 521, 2969; arXiv:2210.12121), and thermohaline mixing halves the
signature within 5 to 45 Myr, with solar-like stars keeping more than 0.05 dex
for about 20 Myr to 1.7 Gyr depending on structure (Behmard, Sevilla and Fuller
2022, MNRAS 518, 5465; arXiv:2210.11679). Both numbers are model-dependent and
somewhat inconsistent with each other, so treat the lifetime as
"tens of Myr, order-uncertain". The good news is that a 5 to 9 Myr flight time
is inside every published window, so a recent engulfment should still be
visible.

Second, and more important, **the calibration range does not extend to our
targets**. Every published ingestion case involves roughly 10 to 30 Earth
masses, and the observed amplitude is 0.05 to 0.23 dex. Blanets span 20 to 3000
Earth masses, so the upper end is one to two orders of magnitude beyond anything
calibrated. Peak transient luminosity scales steeply with engulfed mass
(planet about 1e35 erg/s against contact-binary merger about 1e6 L_sun, about
4e39 erg/s, from V1309 Sco and V838 Mon), so the abundance response to a
super-Jupiter-mass solid body is **not** an extrapolation anyone has published.
Expect it to be qualitatively different from the solar-twin template and do not
assume 0.2 dex is the ceiling.

Verdict: run this channel now. It is demonstrated, the data are public, and the
cost is one spectrum per candidate.

### 3b. Degenerate-body merger as a smoking gun (rank 2, with a major revision)

The earlier draft asserted that merging a degenerate rocky body into a normal
star would make the photosphere "tens of percent iron and rock". **The notes
refute this as stated, and the brief should not carry it forward.** Mass
balance: adding 10 to 20 Earth masses of iron to a solar-mass star changes the
whole-star metallicity Z by only about 1e-5, and even confined to a thin
convection zone the photospheric bump is at most a few tenths of a dex
(the Kronos calibration). A merged **white dwarf** delivers C, O and He, so the
product is hydrogen-poor and **carbon-rich**, not iron-rich (Sakurai's Object,
R Coronae Borealis stars, He-sdO/B stars). Reliable stellar metallicities stop
at about [Fe/H] = +0.56 (HD 126614 A, Howard et al. 2010, ApJ 721, 1467,
arXiv:1003.3488; the super-metal-rich threshold of [Fe/H] > +0.2 is Feltzing and
Gonzalez 2001, arXiv:astro-ph/0012369; five stars at [Fe/H] >= +0.4 in Castro et
al. 1997, arXiv:astro-ph/9704220; 14 Her about +0.43, HD 77338 +0.35, Jenkins et
al. 2012, arXiv:1207.1012; the metal-rich cluster NGC 6791 at [M/H] = +0.39 +/-
0.05, Carraro et al. 2006, arXiv:astro-ph/0512650). The extreme tail is
low-fidelity: Gaia BP/RP photometric metallicity defines an "extremely
metal-rich" GC knot above [M/H]_XP > +0.5 with a tail toward about +1.0, and the
authors themselves call for dust-penetrating spectroscopy to confirm
(Rix et al. 2024, arXiv:2406.01706). Bar stars average <[Fe/H]> = +0.30 (Wegg et
al. 2019, arXiv:1909.11105). Theory reaches only Z = 0.1, which is described as
the first such models in the literature (Karakas et al. 2021,
arXiv:2111.01308; Cinquegrana et al. 2023, arXiv:2308.06002; opacity tables to
Z = 0.2 in Farag et al. 2024, arXiv:2406.02845 is capability, not evidence).

What survives of the channel, and it is worth keeping:

1. **Composition decides the outcome, and each outcome has a distinct
   observable.** A C/O degenerate body pushed to the limit ends as a Type Ia
   supernova. An ONeMg body ends as electron-capture supernova (energy about
   1e50 erg, nickel below 0.015 solar masses, about 1.36 solar-mass neutron
   star, faint subluminous light curve under 100 days; Kitaura, Janka and
   Hillebrandt 2006, arXiv:astro-ph/0512065) or accretion-induced collapse
   (collapse in tens of ms, 3 to 4e49 erg, only about 1e-3 solar masses ejected;
   Dessart et al. 2006, arXiv:astro-ph/0601603). ECSN is the accepted channel
   for SN 1054 and now for SN 2018zd, identified on six independent indicators
   (Hiramatsu et al. 2021, Nature Astronomy 5, 903; arXiv:2011.02176). An iron
   body has only about 0.05 solar masses of headroom above its 1.06 limit and
   neutronizes before any carbon ignites, so it most plausibly gives a **silent**
   collapse to a neutron star: no bright transient, easy to miss. The AIC optical
   transient, if rotation drives a disc wind, is nickel-powered, peaks in about
   a day, and is about 100 times fainter than a Type Ia (Metzger, Piro and
   Quataert 2009, arXiv:0812.3656; Longo Micchi et al. 2023, arXiv:2306.04711,
   who put the electromagnetic counterpart about 2 orders below Type Ia).
   Gravitational waves from AIC are a generic bounce signal detectable for
   **Galactic** events with LIGO, to about 8 Mpc with third generation
   (Abdikamalov et al. 2010, arXiv:0910.2703). All of the transient predictions
   are simulation-based (proposed), the collapse physics is established.
2. **The Thorne-Zytkow route is the only theoretical path to an iron-peak and
   r-process-dominated photosphere from a degenerate merger, and no TZO is
   confirmed.** Predicted anomalies are Li plus the interrupted rp-process
   products Rb, Sr, Y, Zr, Mo (Podsiadlowski, Cannon and Rees 1995, MNRAS 274,
   485, population about 20 to 200 at any time; Cannon 1993, MNRAS 263, 817;
   Biehle 1991, ApJ 380, 167). The observational recipe is line ratios: Rb I
   7800.23 against Ni I 7798 and Fe I 7802, Mo I 5570.40 against Fe I 5569,
   Li I 6707.97 against K I 7699 and Ca I 6573 (Levesque, Massey, Zytkow and
   Morrell 2014, MNRAS Letters 443, L94; arXiv:1406.0001, for HV 2112). HV 2112
   has been largely refuted as an about 5 solar-mass AGB star (Beasor et al.
   2018, MNRAS 479, 3101; arXiv:1806.07399), 11 similar variables sit at 5 to
   10 solar masses, below the about 15 needed (O'Grady et al. 2020, ApJ 901,
   135; arXiv:2008.06563), with a 2023 re-reading that leaves HV 2112 ambiguous
   between a maximally massive super-AGB star and a TZO (O'Grady et al. 2023,
   ApJ 943, 18; arXiv:2211.12438). Note the rocky-core variant of a TZO in the
   draft is our own extension, not a published model.
3. **The real detection hook is pipeline failure, not a big reported number.**
   An automated search for a metal-dominated star should look for **grid-edge
   pinning and flags**, because that is what the pipelines do. APOGEE DR17
   abundance grids all cap at [M/H] = +0.5; a more metal-rich star is pinned at
   the edge and flagged GRIDEDGE_BAD/GRIDEDGE_WARN and M_H_BAD/M_H_WARN, and
   since DR17 a metallicity grid-edge star is *not* set STAR_BAD but the
   calibrated M_H tag is withheld. GALAH sets flag_sp bit 1024 ("MARCS grid limit
   reached or outside reasonable parameter range") and flag_fe_h = 16 ("no
   reliable measurement"), and DR3 deliberately dropped data-driven labels
   because they are invalid outside the training space. So: query for the flags
   and the pinning, plus poor chi-squared from saturated blended Fe I lines, not
   for a high [Fe/H] value that will never be reported. A useful opposite-sign
   contrast is the lambda Boötis class, unusually *low* surface iron-peak
   abundances, plausibly from accreting metal-poor gas or hot-Jupiter mass loss
   (Jura 2015, AJ 150, 166; arXiv:1509.04672).

Verdict: keep as the highest-value single test, but re-scope the target from
"iron-dominated star" to (i) hydrogen-poor, carbon- or refractory-rich anomalies
with GC-traced kinematics, (ii) grid-edge pinning and flag overrides in APOGEE
and GALAH, and (iii) the Rb/Mo/Li ratio test in a cool luminous giant. One
spectrum still settles it; what changes is what we ask the spectrum to show.
Also note the gap that no dedicated search for a genuinely iron-peak-dominated
photosphere has been published, so its absence is partly unsearched rather than
purely excluded.

### 3c. Microlensing free-floating planets toward the bulge (rank 3)

The most mature channel and the one where a null result is already on the
books. Demonstrated results:

- the first claim of a large Jupiter-mass free-floating population, about 1.8
  (-0.8, +1.7) per main-sequence star from 2 years of MOA-II (Sumi et al. 2011,
  Nature 473, 349; arXiv:1105.3544), later contradicted in normalization
- a 6x larger OGLE sample finding no 1 to 2 day excess: **95 percent upper limit
  of 0.25 Jupiter-mass free-floating or wide-orbit planets per main-sequence
  star**, i.e. the Jupiter-mass population is excluded (Mroz et al. 2017, Nature
  548, 183; arXiv:1707.07634)
- the census is **bottom-heavy**: MOA-II 9-year short-event sample gives
  dN/dlogM = 2.18 (+0.52, -1.40) x (M/8 Earth masses)^(-0.96) per star below
  0.02 solar masses, about 21 (+23, -13) free-floating or wide-orbit planets per
  star over 0.33 to 6660 Earth masses, about 19 times more free-floating than
  wide-orbit, and predicts Roman will find 988 (+1848, -566) down to Mars mass
  (Sumi et al. 2023, AJ; arXiv:2303.08280)
- individual masses are degenerate until parallax is measured. The **first and
  only directly measured free-floating planet mass** is
  0.219 (+0.075, -0.046) Jupiter masses (about 0.7 Saturn) for
  KMT-2024-BLG-0792/OGLE-2024-BLG-0516 from ground plus Gaia space parallax,
  argued to be an ejected planet rather than an isolated brown dwarf (Dong et
  al. 2025, Science; arXiv:2601.00057). Other theta_E-measured candidates:
  KMT-2024-BLG-3237 at 67 (+103, -43) Earth masses (Inyanya et al. 2026,
  arXiv:2602.22709), OGLE-2023-BLG-0524 with a record 25-year HST baseline
  rejecting 25 to 48 percent of stellar companions (Kapusta et al. 2025, A&A
  706, A110; arXiv:2507.01109), and KMT-2024-BLG-0816 (Poleski et al. 2026,
  arXiv:2601.00325). Two homed ultra-short events in Mróz et al. 2019, A&A 622,
  A201 (arXiv:1811.00441)
- host rejection is now demonstrated interferometrically: GRAVITY observations
  two years after peak found no luminous host out to 90 au projected separation
  with a 5 sigma limit of 0.5 to 0.6 solar masses (Mroz et al. 2026,
  arXiv:2609.12743)
- caveats from the same data: up to about half of reported "free-floating
  Neptunes" may simply be wide-orbit (Hadden and Wu 2025, arXiv:2507.08968),
  masses above about 1 Jupiter mass may often be bound to massive hosts or the
  low-mass stellar tail (Yee and Kenyon 2025, arXiv:2503.11597), and short
  low-amplitude candidates can be stellar flares (Mroz 2023, arXiv:2404.16480)
- near-infrared microlensing toward the **central** bulge, inaccessible to
  optical surveys, began operationally with PRIME in 2024 February, 486
  candidates through mid-2025 (Sumi et al. 2025, AJ 170, 338;
  arXiv:2508.14474). KMTNet's full-frame pipeline reaches theta_E about 1
  microarcsecond, i.e. sub-Earth lenses, over 483068 stars (Qian et al. 2025,
  PASP; arXiv:2503.24097), and DREAMS projects Mars-mass sensitivity in the
  bulge and Moon-mass in the disc (Yang et al. 2026, AJ; arXiv:2601.11469, with
  the first data release at arXiv:2605.27364). An independent non-microlensing
  bound: NANOGrav 15-year pulsar timing gives about 6.03e5 pc^-3 for
  Jupiter-mass free-floating objects (Dey et al. 2025; arXiv:2507.19475)

Read-across to our targets. Blanets as published run 20 to 3000 Earth masses,
which overlaps the measured bottom-heavy FFP population almost entirely, so a
Blanet-origin claim from event counts alone is not separable from the general
rogue-planet population. What *is* distinct is a Jovian-to-super-Jovian excess in
the bulge direction specifically, since the demonstrated census excludes a large
Jupiter-mass population field-wide. That makes a detection cleanly diagnostic if
it happens, and it is a separate, larger project than this brief.

Verdict: watch, do not fund as the primary. The surveys that will build the
Jovian census (PRIME, KMTNet full-frame, DREAMS, Roman GBTDS) are already
running or projected (Roman yields are projected, not demonstrated; Liveoak et
al. 2026, arXiv:2609.06112; Terry et al. 2025, arXiv:2510.13974).

### 3d. Direct and astrometric detection near Sgr A* (rank 4)

Proposed, unproven. Demonstrated capability exists: VLTI/GRAVITY reaches K
fainter than 19 mag with sub-milliarcsecond resolution and microarcsecond
astrometry (Eisenhauer and Monnier 2023, ARAA; arXiv:2303.00453), and S2 was
astrometrically measured to about 30 microarcseconds per epoch, enough to detect
Schwarzschild precession (GRAVITY Collaboration 2020, A&A 636, L5;
arXiv:2004.07187). Keck/GSAOI-class adaptive optics achieves 0.2 to 0.6 mas
astrometry in the crowded field (Neichel et al. 2014, arXiv:1409.0719; Haggard
et al. 2024 tracks about 145 high-proper-motion stars across the central parsec
over about 10 years).

But no published planet search, and no published companion limit, exists for the
nuclear S-stars. A Jupiter at about 1 au around a about 10 solar-mass S-star at
8 kpc induces a wobble of tens of microarcseconds or less, near or below the
noise floor, so only super-Jupiter or near-brown-dwarf companions are accessible
at all (inference from the quoted precision, not a published limit). Radial
velocity is effectively foreclosed: K about 14 to 18 targets, A_Ks 2.4, and
near-infrared precision around 100 m/s is about two orders of magnitude short of
what a Jupiter analog needs. ELT/METIS imaging of the inner arcsecond is a
projected channel and, per the notes, no single citable primary source gives its
sensitivity in a 1e5 pc^-2 field; the binding constraint inside about 1 pc is
crowding, not raw contrast. On the theory side, isolated rogue planets near the
GC are "hardly detectable by current or forthcoming facilities unless tidally
disrupted" (Mapelli and Ripamonti 2015, arXiv:1504.04624; a tidally stripped
photoevaporating embryo could reach Br gamma about 1e31 erg/s, comparable to G2),
and tidally captured planets were proposed to explain G1/G2 orbits (Trani et
al. 2016, arXiv:1608.07230), an interpretation that has not succeeded: G2 is
standardly read as a dusty gas cloud or stellar-wind shell.

Verdict: park. Worth one low-cost action, which is to ask the GRAVITY and Keck
groups whether unpublished S-star companion limits exist, since the absence in
the literature may be an absence of publication rather than of data.

### 3e. The bodies themselves, without a star (rank 5; two sub-channels)

**Microlensing of a dark solar-mass body (established machinery).** For
0.5 to 1 solar masses at 4 kpc toward the bulge (source at 8 kpc, mu_rel 4 to
6 mas/yr), theta_E = 0.7 to 1.0 mas and **t_E = 50 to 75 days**, since t_E
scales as sqrt(M) (kappa = 8.144 mas/solar mass). Such an event is trivially
detectable in ongoing OGLE/MOA/KMTNet streams, which find thousands per year.
The difficulty is not sensitivity but distinguishing it from an ordinary stellar
or brown-dwarf lens, which needs parallax plus finite-source or astrometric mass
measurement. Population bounds are already tight: MACHO found 13 to 17 events in
5.7 years toward the LMC with optical depth 1.2e-7 and about 20 percent halo
fraction (95 percent CI 8 to 50 percent), excluding a 100 percent MACHO halo
(Alcock et al. 2000, ApJ 542, 281; arXiv:astro-ph/0001272); EROS-2 found 1
candidate against about 39 expected for a full halo, giving halo fraction
**< 8 percent at 95 percent CL** and ruling out compact objects from 0.6e-7 to
15 solar masses as the primary halo occupants (Tisserand et al. 2007, A&A 469,
387; arXiv:astro-ph/0607207). Primordial black holes are separately constrained
by M31 monitoring (Niikura et al. 2019, Nature Astronomy 3, 524;
arXiv:1701.02151).

**EMRI / LISA (loud if rare).** A stellar-mass compact object (1 to 100 solar
masses) inspiraling into a 1e4 to 1e7 solar-mass black hole is an EMRI by
definition, with mass ratio 1e-4 to 1e-7; sub-stellar objects into Sgr A*-class
holes are "XMRIs" (Amaro-Seoane 2022, arXiv:2011.03059). For a solar-mass body
into Sgr A* (4.1e6 solar masses, 8.3 kpc), the frequency at ISCO is about
0.54 mHz with the signal in roughly 0.1 to 0.55 mHz, at the low-frequency edge
of the LISA band (1e-4 to 1e-1 Hz; Amaro-Seoane et al. 2017, arXiv:1702.00786;
Amaro-Seoane et al. 2023, Living Rev. Rel. 26, 2, arXiv:2203.06016), and the
strain is about 2.5e-18, giving **SNR about 1e4 to 1e6** over 4 years against a
threshold of about 20. The lesson that a Galactic-center inspiral would be the
loudest source on the sky goes back to Hils and Bender (1995, ApJ 445, L7).
Expected LISA EMRI detection rates are at least a few per year up to a few
thousand per year under optimistic assumptions (Babak et al. 2017, PRD 95,
103012, arXiv:1703.09722; Gair et al. 2004, arXiv:gr-qc/0405137).

The catch is occurrence, not sensitivity: capture and in-band inspiral rates are
about 1e-6 to 1e-4 per year, so we almost certainly do not have one live during
LISA. A further physical selection also kills the "watch Sgr A* flare up"
variant of this idea. A 0.6 solar-mass white dwarf has r_t about 2.6 r_g, near
the ISCO, so a white-dwarf EMRI terminates in tidal disruption near the horizon
rather than a clean plunge (arXiv:2011.03059; Hils and Bender 1995). And a
solar-mass degenerate body does **not** make an observable disruption flare at
Sgr A*: with R_S = 1.21e7 km (about 0.081 au), a solar-radius white dwarf has
r_t about 1.6e6 km, so r_t/R_S about 0.13, and it is **swallowed whole**; a
neutron star gives r_t/R_S about 1e-4. A main-sequence star, with r_t about
1.1e8 km, about 9 R_S, is disrupted as a classical TDE. White-dwarf tidal
disruption is therefore an intermediate-mass-black-hole phenomenon
(M_BH < about 2e5 solar masses), where deep passages compress the dwarf into
explosive burning releasing about 1e51 erg plus a near-Eddington soft X-ray flare
(L_X about 1e41 (M_BH/1e3 solar masses) erg/s) lasting months to a year and an
underluminous optical transient with Doppler widths of about 10000 km/s
(Rosswog, Ramirez-Ruiz and Hix 2009, ApJ 695, 404, arXiv:0808.2143, and the
earlier conference version arXiv:0811.2129; MacLeod et al. 2016, ApJ 819, 3,
arXiv:1508.02399; low-mass dwarfs with extended hydrogen envelopes can be
stripped by 1e5 to 1e7 solar-mass holes, giving 0.6 to 11 day flares such as
Dougie and PTF10iya, Law-Smith et al. 2017, arXiv:1701.08162; review: Maguire et
al. 2020, Space Sci. Rev., arXiv:2004.00146). No WD-TDE has been conclusively
detected. Foundational dynamics: Luminet and Marck 1985, MNRAS 212, 57.

Also in this bucket, and explicitly deprioritized: blanet tidal disruption as a
detection channel. For a 100 Earth-mass, 6 Earth-radius blanet around a 1e7
solar-mass hole, the tidal radius is 0.82 au (8.3 r_g), frozen-in fallback peaks
at about 3.7e40 erg/s (about 3e-5 Eddington) as a **century-long UV transient**,
at a rate of only 1e-7 to 1e-6 per AGN per year from scalar resonant relaxation,
with debris gravitational waves about 9 orders below LISA (Pandey and Singh
2026, arXiv:2606.06884). The paper itself concludes these are not promising
targets. Context on the host: past AGN activity may have photoevaporated bulge
planet atmospheres, and the post-AGN center is not itself a practical planet
detector (Ishibashi 2024, MNRAS 533, 455, arXiv:2410.22428; Wislocka, Kovacevic
and Balbi 2019, A&A 624, A71, arXiv:1902.07950). No paper proposing quasar
microlensing by planets as a detection method was found, so treat that idea as
unsupported.

Verdict: the MACHO-style sub-channel is nearly free (re-mining existing event
catalogs for 50 to 75 day events with no lens flux and an astrometric mass) and
should be assigned, but a population is already bounded to under about 8 to 10
percent of halo dark matter. LISA is a decade-scale bet on a rate of about 1e-5
per year.

### 3f. Summary table

| Channel | Status | Blocking issue | Cost to start |
|---|---|---|---|
| a. engulfment spectra in GC-traced stars | demonstrated technique | B/A bias in the kinematic sample; calibration does not reach blanet masses | lowest: public data, 1 spectrum/star |
| b. degenerate-body merger | physics established, counterparts proposed | "iron-dominated star" is refuted; re-scope to H-poor/C-rich, flags, TZO ratios | low |
| c. microlensing FFP toward bulge | demonstrated survey, mature limits | field-wide rogue census is bottom-heavy and not separable by origin | medium, separate project |
| d. direct/astrometric near Sgr A* | proposed only, no published limits | sensitivity floor plus crowding | low as an inquiry, high as a program |
| e. bodies themselves | established machinery (microlensing), proposed occurrence (EMRI) | such bodies already bounded to under 8 to 10 percent of halo dark matter; EMRI rate 1e-6 to 1e-4/yr | low for catalog re-mining |

---

## 4. What to pull, and the concrete first screening action

### 4.1 Data and tools

For the kinematic backbone (channels a and b):

1. **Gaia DR3 astrometry**: parallax and proper motions with full covariance,
   about 1.47 billion stars (Vallenari et al. 2023, A&A 674, A1;
   arXiv:2208.00211). From the ESA archive.
2. **Radial velocities**: Gaia DR3 RVS gives 33,812,183 stars pushed to
   G_RVS = 14 mag (Katz et al. 2023, A&A 674, A5; arXiv:2206.05902). This RV
   availability, not the astrometry, is the completeness bottleneck for any 6D
   selection.
3. **Distances**: Bailer-Jones et al. 2021 photogeometric catalog, 1.47 billion
   geometric and 1.35 billion photogeometric (AJ 161, 147; arXiv:2012.05220).
   Use these, not inverse parallax.
4. **Orbit integration**: galpy (Bovy 2015, ApJS 216, 29; arXiv:1412.3451) or
   gala, with the McMillan 2017 potential (R0 = 8.20 kpc, v0 = 232.8 km/s,
   M_vir about 1.3e12; MNRAS 465, 76, arXiv:1608.00971), plus emcee for the
   Monte Carlo. The recipe is established practice: 100 samples from a 4D
   Gaussian over (parallax, mu_alpha*, mu_delta, v_rad) with uncertainties and
   correlations, backward-integrate 11 Gyr in 0.1 Myr steps, record minimum
   disc-crossing radius r_min and flight time, screen at r_min < 1 kpc, and
   confirm with a pericenter within a few to tens of parsecs (the sphere of
   influence is only about 1 to 10 pc) or a Bayes factor as in Koposov et al.
   2020 and Marchetti et al. 2022 (arXiv:2206.06962).
5. **Do not rebuild the candidate list from scratch.** Start from published
   selections: Brown 2014 (21 unbound late-B, arXiv:1401.7342), Brown et al.
   2018 (arXiv:1805.04184), Marchetti, Rossi and Brown 2019 DR2 in 6D (125
   sources above 50 percent unbound, arXiv:1804.10607), Marchetti 2021 EDR3
   (94 above 50 percent, arXiv:2012.02123), Marchetti et al. 2022 DR3 in 6D
   (arXiv:2206.06962), Boubert et al. 2018 Open Fast Stars Catalog
   (arXiv:1804.10179; note faststars.space no longer resolves, content is in the
   Internet Archive), Hattori et al. 2018 (30 metal-poor extreme-velocity
   stars, arXiv:1805.03194), Li et al. 2023 (12784 high-velocity stars, 52
   marginally unbound late-type candidates, 40 new; arXiv:2207.04406),
   Liao et al. 2023 (ApJL 944, L39: two DR3 RVS stars passing within 1 kpc), and
   Koposov et al. 2020 (S5-HVS1, the only confident GC link).
6. **Chemistry**: APOGEE DR17 (near-infrared, reads refractories and CNO through
   the dust), GALAH (optical, reads Li; southern coverage, which matters for the
   bulge), LAMOST (cheap RVs for fainter stars), Gaia-ESO. Flag handling is the
   actual signal here (section 3b).
7. **White dwarf channel**: the Gaia white dwarf catalog (Gentile Fusillo et
   al.; not yet pinned in these notes, flagged as a to-do), crossed with the
   same kinematics, then checked for metal pollution. Note that known
   hypervelocity white dwarfs are SN Ia ejecta (Shen et al. 2018), so origin
   inference must be applied before they count.
8. **Reddening**: Bayestar19 plus a GC-specific map (Surot et al.; verify) to
   define where the optical zone of avoidance actually ends.

### 4.2 First screening action (concrete, low cost)

Take the published kinematically validated GC-traced candidates, not a new
all-sky search. The realistic starting set is small: S5-HVS1, the two Liao et
al. 2023 candidates within 1 kpc, and the handful of r_min < 1 kpc survivors
from Marchetti et al. 2022 plus Scholz 2024 astrometry-only extremes
(A&A 685, A162; arXiv:2402.10714; note Scholz found that none of the 72 most
extreme DR3 candidates passed all quality checks, so apply artifact rejection
hard, especially near the GC where parallax and proper motion are contaminated).

For each star, answer one gross question instead of a full abundance analysis:
**does this spectrum fit any stellar model, and is its iron abundance sane?**

Cross-match to APOGEE/GALAH and flag on:

- grid-edge pinning or flags: `GRIDEDGE_BAD`/`GRIDEDGE_WARN`,
  `M_H_BAD`/`M_H_WARN`, `CALRANGE_BAD/WARN`, withheld `M_H`; GALAH `flag_sp` bit
  1024, `flag_fe_h` = 16, and the recommended cuts `flag_sp == 0` and
  `flag_fe_h == 0` inverted, i.e. we want the rejects
- a Tcond-correlated refractory excess of any size, and enhanced lithium
  (A(Li) above the twin/cluster baseline)
- enhanced Rb and Mo against Fe and Ni in a cool luminous source, the TZO ratio
  test (Rb I 7800.23, Mo I 5570.40, Li I 6707.97)
- a hydrogen-poor, carbon-rich pattern (Sakurai's Object / R CrB / He-sdO-like),
  which is what a merged degenerate body of C/O composition should actually look
  like
- near-infrared or WISE excess, or a broad/composite spectrum that resists
  classification

Then extend the same query down the velocity distribution to late-type (G/K/M)
and white dwarf candidates, because that is where the chemical tests are
sensitive and where the existing hypervelocity census is thin. Budget: one
dedicated spectrum per flagged candidate.

---

## 5. Open gaps and risks

1. **The degenerate rocky worlds rest on one model.** Mishra et al. 2026
   (arXiv:2605.19241) form pure-dust bodies above the hydrogen-burning limit in
   AGN tori, for black holes of 1e7 to 1e9 solar masses only. Nothing yet
   says the same at Sgr A*'s 4.1e6. This is still the premise's weakest leg.

   Outside that model, solid accretion tops out at about 3000 Earth masses
   (0.01 solar masses), and the band from about 10 Earth masses to about 1
   solar mass is modelled nowhere else.
2. **Blanet formation is contested.** The RATD critique (arXiv:2111.11800)
   argues grain growth fails in the physically favored clumpy torus. If correct,
   channel (a) has no source population.
3. **Sgr A* is a marginal AGN for the model.** Formation is favored at
   M_BH about 1e6 solar masses and L_bol about 1e42 erg/s; Sgr A* is 4.1e6 solar
   masses and much fainter today.
4. **Timescale mismatch.** Blanet formation takes 1e7 to 1e8 yr, and 70 to 80
   Myr at the snow line, while the flight-time window we can observe is a few to
   about 9 Myr and the assumed last active period is about 9 Myr. The notes also
   flag that the 9 Myr window is our own construction, not a cited result. If the
   active phase was shorter than the formation time, the population may not have
   had time to assemble.

   Mishra et al. 2026 ease this. In their model grains grow to millimetres and
   centimetres well within active phases of 1 to 10 Myr, and pebble accretion
   doubles a body's mass in 1e3 to 1e7 yr where grains survive 10 m/s
   collisions, though 1e6 to 1e9 yr where they survive only 1 m/s (Figure 5).
5. **The GC-traced sample is tiny and shrinking.** Gaia DR3 in 6D found **no**
   HVS candidate above 700 km/s with robust kinematics, bounding the Hills
   ejection rate to less than about 8e-5 per year (arXiv:2206.06962); an
   astrometry-only extreme check passed none of 72 candidates
   (arXiv:2402.10714); backward integration yields two candidates within 1 kpc
   (Liao 2023). Any null result from a 5-star sample constrains almost nothing,
   so design for a growing sample (future DR4) rather than a one-shot test.
6. **Spectral-type mismatch.** The ejected census is B/A-dominated by selection
   (arXiv:0808.2469, arXiv:1401.7342), which is the wrong class for
   photospheric-chemistry tests; late-type engulfment candidates recovered from
   earlier surveys were mostly shown to be bound (Ziegerer et al. 2015,
   arXiv:1503.04164; Boubert 2018; Hattori 2018).
7. **Engulfment amplitude is uncalibrated above tens of Earth masses**, so an
   observed offset may be larger than the solar-twin template or may saturate
   differently. Do not apply 0.05 to 0.23 dex as a prior.
8. **The draft's iron-dominated-star expectation is refuted by mass balance** and
   by the absence of any such star; reliable metallicities stop at [Fe/H] about
   +0.56 and models at Z = 0.1. Related: no dedicated published search for an
   iron-peak-dominated photosphere exists, so its absence is partly unsearched.
9. **No TZO is confirmed** and the one candidate is now read as an AGB or
   super-AGB star (arXiv:1806.07399, arXiv:2008.06563).
10. **No Jupiter-to-super-Jovian rogue census exists**, and the demonstrated
    short-event data cut against a Jovian population (arXiv:1707.07634;
    arXiv:2303.08280), so channel (c) currently has no positive expectation.
11. **No S-star planet limits**, and no blanet detection-proposal paper at all;
    observability is essentially unstudied as of 2026.
12. **Numbers still to verify against primary sources before quoting**: the
    sublimation and snow-line prefactors in arXiv:1909.06748 disagree between two
    full-text extractions (L_UV 1e43 vs 1e46, L_X 1.3e42 vs 1.3e41), so quote the
    fiducial radii (r_sub about 0.3 pc, r_snow about 1.4 to 4.7 pc) rather than
    the formulas; the exact iron neutronization threshold density spans
    about 1.0 to 1.3e9 g/cm3 with maximum mass 1.06 +/- 0.02 solar masses per
    Hamada and Salpeter tables not re-read this session; the Zapolsky and Salpeter
    maximum-mass ceiling is metadata-verified only; the about 10 Earth-radius
    blanet figure comes from an ar5iv rendering, not the IOP full text; the
    Gentile Fusillo and Surot citations need pinning; and quantitative
    predictions of rp-element enhancement factors in TZO models could not be
    verified against the primary nucleosynthesis papers.
13. **Corpus completeness.** The citation inventory came from OpenCitations,
    Crossref and one search pass, not an exhaustive ADS full-text sweep, so a very
    recent preprint could be missing.

---

## 6. Recommended sequencing

1. Assign the section 4.2 screening immediately; it is public-data work on a
   small candidate list, and one spectrum per flag.
2. In parallel, re-mine the bulge microlensing catalogs for 50 to 75 day events
   with no lens flux and pursue an astrometric or space-parallax mass for the
   best one or two (channel 3e). Cheap, bounded by EROS-2 but a confirmed
   dark solar-mass lens would still be a discovery.
3. Ask the GRAVITY and Keck GC groups whether unpublished S-star companion
   limits exist before budgeting anything for channel 3d.
4. Treat the Jovian free-floating census (3c) as a separate proposal with its own
   timeline, tied to PRIME, KMTNet full-frame, DREAMS and Roman.
5. Before the program is made public, commission a short theory task on the two
   load-bearing premise gaps: whether Mishra et al.'s pure-dust bodies form at
   Sgr A*'s mass, and whether a 9 Myr AGN episode at 4.1e6 solar masses grows
   any solid body at all. These are where a reviewer can stop the whole case.

---

## 7. Corrections against the earlier draft

1. **Melendez et al. 2009** is ApJL **704, L66**, not 691, L47.
2. **The wide-binary ingestion case is Oh et al. 2018** (ApJ 854, 138,
   Kronos and Krios), not Spina et al. 2021. Spina et al. 2021 (Nature
   Astronomy 5, 1163) is the *statistical 107-binary* study finding engulfment
   in 20 to 35 percent of Sun-like stars. The draft conflated them.
3. **Aguilera-Gomez et al. 2016** is ApJ 829, 127, not Nature Astronomy.
4. **De et al. 2023 (ZTF SLRN-2020) has no arXiv preprint.** Cite by DOI
   10.1038/s41586-023-05842-x. The draft's suggested about 1e4 L_sun and
   T about 5000 to 10000 K are unsupported: the Nature abstract gives
   about 1e35 erg/s, about 26 L_sun, and JWST follow-up finds about 700 K and
   about 300 K dust components. The "red giant swallowing a planet" framing was
   itself revised by Lau et al. 2025.
5. **The "surface tens of percent iron and rock" claim is not supportable.** The
   strongest version we can state: H/He-poor, carbon- and refractory-rich
   anomalies, plus pipeline grid-edge pinning and flag overrides. See 3b.
6. **Blanets are icy, not "icy and rocky"**, in the published models, and they do
   **not** reach degeneracy (max about 0.01 solar masses).
7. The 1b density figure should read central density 1e6 to 1e9 g/cm3 with an
   Earth-like radius, not 1e5 to 1e6 g/cc.
8. **A solar-mass body disrupted by Sgr A* does not flare**; it is swallowed
   whole (r_t/R_S about 0.13). Disruption flares need M_BH below about 2e5 solar
   masses. Any proposal should not promise a Sgr A* WD-TDE.
9. The 9 Myr window, the Wada 2021 and Brown 2015 references, the ZTF and TZO
   citations, and the HVS catalog IDs now have confirmed bibliographic records;
   the *verify on ADS* flags in the draft are resolved except where listed in
   section 5, item 12.
10. The draft's link of the ~9 Myr window to the Fermi bubbles ("a few Myr ago,
   plus a more recent flare") has **no citation anywhere in the corpus**. It is
   plausible framing but it is currently unsourced, so any proposal text that
   keeps it must either find a primary reference or present it as a hypothesis.

---

## 8. Consolidated references

Blanets and AGN planet formation:

- Wada, Tsukamoto and Kokubo 2019, ApJ 886, 107; arXiv:1909.06748 (origin;
  corrigendum ApJ 907, 58)
- Wada, Tsukamoto and Kokubo 2021, ApJ 909, 96; arXiv:2007.15198 (coins
  "blanet"; 20 to 3000 Earth masses)
- Giang, Hoang, Tram et al. 2022, ApJ 936, 108; arXiv:2111.11800 (RATD
  disruption barrier)
- Mishra, Lyra, McKernan, Mac Low, Ford and Cook 2026, ApJ 1005, 99;
  arXiv:2605.19241 (AGN tori as birthplace of millions of planets, and of
  pure-dust bodies above the hydrogen-burning limit; sources/2605.19241.pdf)
- Zhou, Zhang, Huang and Lin 2026, A&A 705, A58 (tidal disruption of planetary
  bodies and Sgr A* flares)
- Arakawa et al. 2023, A&A 670, L21 (porous-aggregate collisional growth
  threshold)

Degenerate body structure and collapse:

- Chandrasekhar 1931, ApJ 74, 81 (ideal limit 1.44 solar masses)
- Hamada and Salpeter 1961, ApJ 134, 683 (mass-radius tables; iron 1.06 solar
  masses)
- Miyaji et al. 1980, PASJ 32, 623; Nomoto 1984, ApJ 277, 791; Hashimoto,
  Iwamoto and Nomoto 1993 (ONeMg electron capture, 1.37 solar masses)
- Zapolsky and Salpeter 1969, ApJ 158, 809 (cold spheres)
- Burrows and Ostriker 2014, PNAS 111, 3454; arXiv:1401.1814
- Seager et al. 2007, ApJ 669, 1279; arXiv:0707.2895 (solid planets to
  20 Earth masses)
- Fortney, Marley and Barnes 2007, ApJ 659, 1661; arXiv:astro-ph/0612671
- Kuchner and Seager 2005, arXiv:astro-ph/0504214 (carbon planets, preprint);
  Mashian and Loeb 2016, MNRAS 464, 4522; arXiv:1603.06943; counter-evidence
  Nissen et al. 2014, A&A 568, A25; arXiv:1406.5218 and Deming and Seager 2017;
  arXiv:1701.00493
- Bashi et al. 2017, A&A 604, A83; arXiv:1701.07654; Odrzywolek and Rafelski
  2018; arXiv:1612.03556; Gottstein et al. 2026; arXiv:2605.18950 (degeneracy as
  the upper boundary of the planet population)
- Kitaura et al. 2006, A&A 450, 345; arXiv:astro-ph/0512065 (ECSN)
- Dessart et al. 2006, ApJ 644, 1063; arXiv:astro-ph/0601603 (AIC)
- Metzger, Piro and Quataert 2009, MNRAS 396, 1659; arXiv:0812.3656; Longo
  Micchi et al. 2023; arXiv:2306.04711 (AIC counterparts)
- Hiramatsu et al. 2021, Nature Astronomy 5, 903; arXiv:2011.02176 (SN 2018zd as
  ECSN)
- Abdikamalov et al. 2010, PRD 81, 044012; arXiv:0910.2703 (AIC gravitational
  waves)
- Canal, Isern and Labay 1990, ARA&A 28, 183; Nomoto and Kondo 1991, ApJ 367,
  L19; Gutierrez et al. 1996, ApJ 459, 543
- Bitsch et al. 2018, A&A 612, A30; arXiv:1801.02341; Piso and Youdin 2013,
  ApJ 765, 33; arXiv:1301.1727; Chametla et al. 2022; arXiv:2112.14214 (pebble
  isolation and core-mass ceilings)

Engulfment and merger signatures:

- Melendez et al. 2009, ApJL 704, L66; arXiv:0909.2299
- Oh et al. 2018, ApJ 854, 138; arXiv:1709.05344 (Kronos and Krios)
- Miquelarena et al. 2024, A&A 688, A73; arXiv:2406.06705
- Spina et al. 2021, Nature Astronomy 5, 1163; arXiv:2108.12040
- Liu et al. 2024, Nature 627, 601; arXiv:2403.13209
- Behmard et al. 2023, MNRAS 521, 2969; arXiv:2210.12121; Behmard, Sevilla and
  Fuller 2022, MNRAS 518, 5465; arXiv:2210.11679; Sevilla et al. 2022, MNRAS
  516, 3354
- Aguilera-Gomez et al. 2016, ApJ 829, 127; arXiv:1605.05332; Aguilera-Gomez et
  al. 2020, ApJL; arXiv:2005.14209; Castro-Tapia et al. 2024, A&A 690, A367;
  arXiv:2401.00049; Charalambous et al. 2026; arXiv:2607.05566
- Zuckerman et al. 2003, ApJ 596, 477; Zuckerman et al. 2010, ApJ 722, 725;
  arXiv:1007.2252; Koester, Gaensicke and Farihi 2014, A&A 566, A34;
  arXiv:1404.2617; Gaensicke et al. 2012, MNRAS 424, 333; Zuckerman et al. 2007,
  ApJ 671, 872; Klein et al. 2010, ApJ 709, 950; arXiv:0912.1422; Klein et al.
  2011, ApJ 741, 64; arXiv:1108.1565; Farihi et al. 2016; arXiv:1608.07278
- Vanderburg et al. 2015, Nature 526, 546; arXiv:1510.06387 (WD 1145+017)
- De et al. 2023, Nature 617, 55, DOI 10.1038/s41586-023-05842-x (**no arXiv
  preprint**); Lau et al. 2025, ApJ 983, 87; arXiv:2504.07275
- Tylenda et al. 2011, A&A 528, A114; arXiv:1012.0163 (V1309 Sco); V838 Mon:
  see notes, Wikipedia-sourced summary in `engulfment_merger_signatures.md`
- Levesque et al. 2014, MNRAS Letters 443, L94; arXiv:1406.0001 (HV 2112);
  Beasor et al. 2018, MNRAS 479, 3101; arXiv:1806.07399; O'Grady et al. 2020,
  ApJ 901, 135; arXiv:2008.06563; O'Grady et al. 2023, ApJ 943, 18;
  arXiv:2211.12438
- Podsiadlowski, Cannon and Rees 1995, MNRAS 274, 485; Cannon 1993, MNRAS 263,
  817; Biehle 1991, ApJ 380, 167; Thorne and Zytkow 1975/1977, ApJ 199, L19 and
  212, 832
- Metallicity records: Feltzing and Gonzalez 2001; arXiv:astro-ph/0012369;
  Castro et al. 1997; arXiv:astro-ph/9704220; Howard et al. 2010, ApJ 721, 1467;
  arXiv:1003.3488; Jenkins et al. 2012, ApJ 766, 67; arXiv:1207.1012; Carraro et
  al. 2006, ApJ 643, 1151; arXiv:astro-ph/0512650; Rix et al. 2024;
  arXiv:2406.01706; Wegg et al. 2019, A&A 632, A121; arXiv:1909.11105; Karakas
  et al. 2021; arXiv:2111.01308; Cinquegrana et al. 2023; arXiv:2308.06002;
  Farag et al. 2024; arXiv:2406.02845; Moedas and Di Mauro 2026;
  arXiv:2605.31060; Jura 2015, AJ 150, 166; arXiv:1509.04672
- Pipeline behavior: SDSS APOGEE DR17 library and ASPCAP documentation
  (https://www.sdss.org/dr17/irspec/apogee-libraries/,
  https://www.sdss.org/dr17/irspec/aspcap/,
  https://www.sdss.org/dr17/irspec/apogee-bitmasks/); GALAH DR3 flags
  (https://www.galah-survey.org/dr3/flags/,
  https://www.galah-survey.org/dr3/using_the_data/,
  https://www.galah-survey.org/dr3/overview/)

Ejected and hypervelocity stars:

- Hills 1988, Nature 331, 687
- Brown 2015, ARA&A 53, 15 (review)
- Brown et al. 2005, ApJ 622, L33; arXiv:astro-ph/0501177
- Brown et al. 2006, ApJ 640, L35; arXiv:astro-ph/0601580; ApJ 647, 303;
  arXiv:astro-ph/0604111; ApJ 660, 311; arXiv:astro-ph/0701600
- Brown, Geller and Kenyon 2009, ApJ 690, 1639; arXiv:0808.2469; 2012, ApJ 751,
  55; arXiv:1203.3543; 2014, ApJ 787, 89; arXiv:1401.7342
- Brown et al. 2015, ApJ 804, 49; arXiv:1502.05069; Brown et al. 2018;
  arXiv:1805.04184
- Marchetti, Rossi and Brown 2019; arXiv:1804.10607; Marchetti 2021, MNRAS 503,
  1374; arXiv:2012.02123; Marchetti, Evans and Rossi 2022, MNRAS 515, 767;
  arXiv:2206.06962
- Boubert et al. 2018; arXiv:1804.10179 (Open Fast Stars Catalog)
- Hattori et al. 2018, ApJ 866, 121; arXiv:1805.03194
- Palladino et al. 2014, ApJ 780, 7; arXiv:1308.3495; Ziegerer et al. 2015, A&A
  576, L14; arXiv:1503.04164
- Koposov et al. 2020, MNRAS; arXiv:1907.11725 (S5-HVS1)
- Liao, Du, Li, Ma and Shi 2023, ApJL 944, L39
- Scholz 2024, A&A 685, A162; arXiv:2402.10714
- Li et al. 2023, AJ; arXiv:2207.04406
- Nemeth et al. 2016, ApJL 821, L13; arXiv:1604.03158 (PB 3877)
- Shen et al. 2018, ApJ 865, 15; arXiv:1804.11163 (hypervelocity white dwarfs)
- Geier et al. 2015, Science 347, 1126 (US 708)
- Tooling and potentials: Bovy 2015, ApJS 216, 29; arXiv:1412.3451; McMillan
  2017, MNRAS 465, 76; arXiv:1608.00971; Bailer-Jones et al. 2021, AJ 161, 147;
  arXiv:2012.05220; Katz et al. 2023, A&A 674, A5; arXiv:2206.05902; Vallenari et
  al. 2023, A&A 674, A1; arXiv:2208.00211

Detection channels toward the bulge and center:

- Extinction: Fritz et al. 2011, ApJ 737, 73; arXiv:1105.2822; Nishiyama et al.
  2008; arXiv:0802.3559; Nishiyama et al. 2009, ApJ 696, 1407;
  arXiv:0902.3095; Gao and Li 2013; arXiv:1305.7137; Chen et al. 2018;
  arXiv:1804.06807; von Fellenberg et al. 2025; arXiv:2511.14850; Haggard et al.
  2024, AJ; arXiv:2408.03467; Shvartzvald et al. 2018, ApJL 857, L2;
  arXiv:1802.06795
- Interferometry and astrometry: Eisenhauer and Monnier 2023, ARAA;
  arXiv:2303.00453; GRAVITY Collaboration 2020, A&A 636, L5; arXiv:2004.07187;
  Neichel et al. 2014; arXiv:1409.0719; O'Neil et al. 2019, AJ 158, 4;
  arXiv:1809.05490; Mroz et al. 2026; arXiv:2609.12743
- Microlensing free-floating planets: Sumi et al. 2011, Nature 473, 349;
  arXiv:1105.3544; Mroz et al. 2017, Nature 548, 183; arXiv:1707.07634; Mroz et
  al. 2019, A&A 622, A201; arXiv:1811.00441; Sumi et al. 2023, AJ;
  arXiv:2303.08280; Dong et al. 2025, Science; arXiv:2601.00057; Inyanya et al.
  2026; arXiv:2602.22709; Kapusta et al. 2025, A&A 706, A110;
  arXiv:2507.01109; Poleski et al. 2026; arXiv:2601.00325; Yee and Kenyon 2025;
  arXiv:2503.11597; Hadden and Wu 2025; arXiv:2507.08968; Sumi et al. 2025, AJ
  170, 338 (PRIME); arXiv:2508.14474; Qian et al. 2025, PASP (KMTNet);
  arXiv:2503.24097; Yang et al. 2026, AJ (DREAMS); arXiv:2601.11469 and
  arXiv:2605.27364; Mroz 2023; arXiv:2404.16480; Dey et al. 2025 (NANOGrav);
  arXiv:2507.19475; Liveoak et al. 2026 (Roman); arXiv:2609.06112; Terry et al.
  2025 (Roman RGES); arXiv:2510.13974; Konacki et al. 2003, ApJ 597, 1076;
  arXiv:astro-ph/0306542 and 2005, ApJ 624, 372; arXiv:astro-ph/0412400; Torres
  et al. 2004, ApJ 614, 979; arXiv:astro-ph/0406627
- Compact-object lensing constraints: Alcock et al. 2000, ApJ 542, 281;
  arXiv:astro-ph/0001272 (MACHO); Tisserand et al. 2007, A&A 469, 387;
  arXiv:astro-ph/0607207 (EROS-2); Niikura et al. 2019, Nature Astronomy 3, 524;
  arXiv:1701.02151
- EMRI and LISA: Amaro-Seoane 2022; arXiv:2011.03059; Amaro-Seoane et al. 2017
  (LISA proposal); arXiv:1702.00786; Amaro-Seoane et al. 2023, Living Rev. Rel.
  26, 2; arXiv:2203.06016; Babak et al. 2017, PRD 95, 103012;
  arXiv:1703.09722; Gair et al. 2004, CQG 21, S1595; arXiv:gr-qc/0405137; Hils
  and Bender 1995, ApJ 445, L7
- Tidal disruption: Rosswog, Ramirez-Ruiz and Hix 2009, ApJ 695, 404;
  arXiv:0808.2143 and J. Phys. Conf. Ser. 172, 012036; arXiv:0811.2129; MacLeod
  et al. 2016, ApJ 819, 3; arXiv:1508.02399; Law-Smith et al. 2017, ApJ 841,
  132; arXiv:1701.08162; Maguire et al. 2020, Space Sci. Rev.;
  arXiv:2004.00146; Luminet and Marck 1985, MNRAS 212, 57
- Blanet and GC planet observability: Pandey and Singh 2026;
  arXiv:2606.06884 (disruption rates 1e-7 to 1e-6 per AGN per year); Mapelli and
  Ripamonti 2015; arXiv:1504.04624; Trani et al. 2016; arXiv:1608.07230;
  Davari and Capuzzo-Dolcetta 2019; arXiv:1906.03738; Ishibashi 2024, MNRAS 533,
  455; arXiv:2410.22428; Wislocka, Kovacevic and Balbi 2019; arXiv:1902.07950
- Habitability context (uses blanets as a premise, not formation or detection):
  Iorio 2020, ApJ 896, 82 and ApJ 889, 152; Iorio 2021, AJ 162, 51; Bakala et
  al. 2020, ApJ 889, 41; Balbi, Hami and Kovacevic 2020, Life 10, 132; Rodener
  et al. 2022, Galaxies 10, 101; Scherf et al. 2024, Astrobiology 24, e916

---

## 9. One-paragraph answer for a reviewer

We cannot see planets around Sgr A* (extinction of 38 to 42 mag in V, dimming
light by 15 or more orders of magnitude, and no demonstrated near-infrared
technique), so the program tests the fossil record carried
out by stars ejected from the center within the last about 9 Myr. That channel,
chemical traces of engulfment, is demonstrated on real stars, runs on public
spectroscopic and astrometric data, and costs one spectrum per candidate from a
list of order ten objects. The two premises that would need a theory push before
this becomes a discovery program are that solid bodies can form at all in a
few-Myr AGN episode at Sgr A*'s mass (contested on grain-disruption grounds;
one model eases the timescale) and that an H/He-poor body near a solar mass
forms at that mass (one model forms them, Mishra et al. 2026, for heavier black
holes only). The spectacular versions of the idea, a metal
dominated star or a white dwarf disruption flare at Sgr A*, do not survive the
physics and should not be promised.

## 10. Amended 2026-11-06

Mishra et al. 2026 (arXiv:2605.19241, sources/2605.19241.pdf) was read,
pages 1 to 3 and 12 to 13 of 16. It forms pure-dust bodies above the
hydrogen-burning limit in AGN tori, so sections 1b, 5 (gaps 1 and 4), 6
and 9, the headline and the reference entry now say so.

What still stands: the model is worked for black holes of 1e7 to 1e9 solar
masses, not Sgr A*'s, and nothing read yet places a torus at Sgr A* when it
was last active. Pages 4 to 11, its methods, were not read.
