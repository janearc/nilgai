//go:build darwin

package main

import "testing"

// The kernel tests' numbers, each defined.
const (
	roomForOne  = 1 // defined: a solver with room for one star
	roomForNone = 0 // defined: room for no stars, which the kernel refuses
)

// the test and the anti-test: the Metal kernel opens over the published
// galaxy with room for one star and closes; with room for none it is
// refused, with no solver beside the refusal.
func TestTheMetalKernelOpensWithRoom(test *testing.T) {
	galaxy, err := published{}.Galaxy()
	if err != nil {
		test.Fatal(err)
	}
	made, err := metal{}.Open(*galaxy, roomForOne)
	if err != nil {
		test.Fatal(err)
	}
	made.Close()
	if none, err := (metal{}).Open(*galaxy, roomForNone); err == nil ||
		none != nil {
		test.Errorf("room for no stars gave %v, %v", none, err)
	}
}
