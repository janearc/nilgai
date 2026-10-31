package nilgai

import (
	"fmt"

	"github.com/janearc/libtheme-css/primitives/length"
	"github.com/janearc/libtheme-css/primitives/si"
)

// Flux is a FluxDensity or an EnergyFlux, and nothing else can be one.
// A caller reads a Flux by switching on its type, so its number never
// leaves it without saying which of the two it is.
//
// Radio and infrared catalogues give flux density, and gamma-ray and
// X-ray catalogues give energy flux. Converting between the two needs a
// bandwidth, and nilgai does not guess one.
type Flux interface{ flux() }

// FluxDensity is a flux density in watts a square metre a hertz. Jansky
// makes one from the janskys a catalogue gives.
type FluxDensity float64

// EnergyFlux is an energy flux in watts a square metre. ErgPerSecPerCm2
// makes one from the erg/s/cm^2 a catalogue gives.
type EnergyFlux float64

// flux makes FluxDensity a Flux.
func (FluxDensity) flux() {}

// flux makes EnergyFlux a Flux.
func (EnergyFlux) flux() {}

// oneFlux is FluxDensity or EnergyFlux as a type parameter. A constructor
// that takes a flux and its error as one oneFlux takes two of a kind.
type oneFlux interface {
	FluxDensity | EnergyFlux
	Flux
}

// wattsPerSquareMetreHertzInAJansky is defined: the IAU defines the
// jansky as 1e-26 W/m^2/Hz.
const wattsPerSquareMetreHertzInAJansky = 1e-26

// joulesInAnErg is defined: the erg is the cgs unit of energy, 1e-7 J.
const joulesInAnErg = 1e-7

// Jansky returns janskys as a FluxDensity.
func Jansky(janskys float64) FluxDensity {
	return FluxDensity(janskys * wattsPerSquareMetreHertzInAJansky)
}

// ErgPerSecPerCm2 returns ergs a second a square centimetre as an
// EnergyFlux.
func ErgPerSecPerCm2(ergs float64) EnergyFlux {
	centimetre := si.Centi(length.Metre)
	return EnergyFlux(ergs * joulesInAnErg / (centimetre * centimetre))
}

// whyNotAFlux returns why a flux cannot be measured: it is not a finite
// number, or it is below zero. It returns nil for a flux that can be.
func whyNotAFlux(flux float64) error {
	if !Finite(flux) {
		return fmt.Errorf("%v is not a finite number", flux)
	}
	if flux < none {
		return fmt.Errorf("%v is below zero", flux)
	}
	return nil
}
