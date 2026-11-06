//go:build darwin

package main

import "testing"

// the test and the anti-test: the kernel opens on this machine's Metal
// device over the published galaxy, with room for one star, and
// closes; room for none is refused. On a Mac without a Metal device
// this fails, saying so; it does not skip.
func TestTheMetalKernelOpensOverTheGalaxy(test *testing.T) {
	galaxy, err := published{}.Galaxy()
	if err != nil {
		test.Fatal(err)
	}
	made, err := metal{}.Open(*galaxy, oneStarRan)
	if err != nil {
		test.Fatal(err)
	}
	made.Close()
	none, err := (metal{}).Open(*galaxy, noRows)
	if err == nil || none != nil {
		test.Errorf("room for no star gave %v, %v", none, err)
	}
}
