package nilgai

import "fmt"

// Unit names the unit of a flux. Two fluxes of different units are
// never added, compared or paired.
type Unit string

const (
	// Jansky is the unit of flux density, Jy: 1e-26 W/m^2/Hz.
	Jansky Unit = "Jy"
	// ErgPerSecPerCm2 is the unit of energy flux, erg/s/cm^2.
	ErgPerSecPerCm2 Unit = "erg/s/cm^2"
)

// Flux is a brightness with its unit in its type. Only FluxDensity and
// EnergyFlux are Fluxes. Radio and infrared catalogues give flux
// density; gamma-ray and X-ray catalogues give energy flux. Converting
// between the two needs a bandwidth and nilgai does not guess one.
type Flux interface {
	// Unit says which unit Value is in.
	Unit() Unit
	// Value is the number, in Unit.
	Value() float64
}

// FluxDensity is a flux density in Jy.
type FluxDensity float64

// Unit is Jansky.
func (FluxDensity) Unit() Unit { return Jansky }

// Value is the flux density in Jy.
func (density FluxDensity) Value() float64 { return float64(density) }

// EnergyFlux is an energy flux in erg/s/cm^2.
type EnergyFlux float64

// Unit is ErgPerSecPerCm2.
func (EnergyFlux) Unit() Unit { return ErgPerSecPerCm2 }

// Value is the energy flux in erg/s/cm^2.
func (energy EnergyFlux) Value() float64 { return float64(energy) }

// checkFlux refuses a missing flux, one that is not a finite number, or
// one below zero. what names the flux in the message.
func checkFlux(what string, flux Flux) error {
	if flux == nil {
		return fmt.Errorf("%s is missing", what)
	}
	if !notBelowNone(flux.Value()) {
		return fmt.Errorf("%s %v %s must be finite and not below zero",
			what, flux.Value(), flux.Unit())
	}
	return nil
}
