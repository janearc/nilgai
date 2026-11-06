package read_test

import (
	"errors"
	"io"
	"iter"
	"math"

	"github.com/janearc/nilgai"
)

// catalogueReader is anything that reads a catalogue, as each of the
// readers does.
type catalogueReader interface {
	Read(name string, in io.Reader) iter.Seq2[*nilgai.Source, error]
}

// errSourceBesideRefusal is what collect adds to its errors when a reader
// yields a source beside an error, so that any test counting errors fails.
var errSourceBesideRefusal = errors.New("a source came beside a refusal")

// collect reads everything the input holds. collect returns the sources
// and the errors that the reader yielded, each in order.
func collect(table catalogueReader, in io.Reader) (sources []nilgai.Source,
	errs []error) {
	for source, err := range table.Read("t", in) {
		if err != nil {
			if source != nil {
				errs = append(errs, errSourceBesideRefusal)
			}
			errs = append(errs, err)
			continue
		}
		sources = append(sources, *source)
	}
	return sources, errs
}

// janskys returns a flux density in Jy, the unit the fixtures work
// infrared fluxes in, and whether the flux is a flux density at all.
func janskys(each nilgai.Flux) (Num, bool) {
	density, isDensity := each.(nilgai.FluxDensity)
	return inJanskys(density), isDensity
}

// inJanskys returns a flux density in Jy.
func inJanskys(density nilgai.FluxDensity) Num {
	return Num(density / nilgai.Jansky(one))
}

// ergs returns an energy flux in erg/s/cm^2, the unit the fixtures work
// X-ray and gamma-ray fluxes in, and whether the flux is an energy flux.
func ergs(each nilgai.Flux) (Num, bool) {
	energy, isEnergy := each.(nilgai.EnergyFlux)
	return Num(energy / nilgai.ErgPerSecPerCm2(one)), isEnergy
}

// turnsApart returns how far apart two angles in degrees are, measured
// the short way round the circle.
func turnsApart(got, want float64) float64 {
	turn := float64(halves * degreesInHalfACircle)
	apart := math.Mod(math.Abs(got-want), turn)
	return min(apart, turn-apart)
}
