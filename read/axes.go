package read

import (
	"math"

	"github.com/janearc/nilgai"
)

// orderAxes returns a catalogue's two semi-axes with the larger one first
// as the major axis, and the angle of that major axis. The angle passed in
// belongs to the first axis.
//
// When the second axis is the larger, orderAxes swaps them. The major axis
// then lies a quarter turn from where the first axis pointed, so the angle
// turns a quarter turn too, wrapping at the full circle.
func orderAxes(first, second, angle float64) (major, minor,
	turned float64) {
	if second > first {
		turned = math.Mod(angle+nilgai.RightAngle, nilgai.FullCircle)
		return second, first, turned
	}
	return first, second, angle
}
