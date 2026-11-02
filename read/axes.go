package read

import (
	"math"

	"github.com/janearc/nilgai"
)

// orderAxes puts the larger semi-axis first and returns its position angle. The
// input angle belongs to the first axis.
//
// When the axes swap, the angle rotates a quarter-turn and wraps at a full
// circle, preserving the orientation of the same ellipse.
func orderAxes(first, second, angle float64) (major, minor,
	turned float64) {
	if second > first {
		turned = math.Mod(angle+nilgai.RightAngle, nilgai.FullCircle)
		return second, first, turned
	}
	return first, second, angle
}
