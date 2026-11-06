# firang: the stages

1. **Candidate selection by orbit traceback.** Every Gaia DR3 star with
   a radial velocity runs back ten million years through the Milky
   Way. A candidate's path comes from Sgr A* across its measurement
   errors, and it moves fast and outward through the galaxy now.
2. **Injection-recovery.** Fake stars thrown from Sgr A* are observed
   with real Gaia errors and go through the same selection. That says
   what the selection would find, and what it would miss.
3. **Spectroscopic follow-up.** Each candidate's abundances: refractory
   elements against volatile ones, and lithium.

cmd/firang runs the stages, a verb each, by the numbers in numbers.yaml:

- screen: every kept star run back at both estimates of the Sun's speed.
- draws: each screened star run back across its errors.
- validate: stars thrown from Sgr A*, drawn as Gaia would see them.
- candidates: a folder for each star the numbers choose.

The errors drawn are each star's five astrometric parameters, from
their covariance, and its radial velocity. With them, the Sun's speed
along the rotation is drawn between its two estimates: the frame's,
from Sgr A*'s apparent motion, and the one the galaxy model implies.
