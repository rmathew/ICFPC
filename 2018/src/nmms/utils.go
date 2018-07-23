// Package nmms contains libraries for the Nanobot Matter Manipulation System
// simulator.
package nmms

import (
	"fmt"
	"os"
)

func iAbs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func iMin(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func iMax(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

// ExitWithErrorMsg prints the given error-message to stderr and then exits with
// an error-code.
func ExitWithErrorMsg(errMsg string) {
	fmt.Fprintf(os.Stderr, "ERROR: %v\n", errMsg)
	os.Exit(1)
}

// Check exits with an error-message if there is an error.
func Check(e error) {
	if e != nil {
		ExitWithErrorMsg(e.Error())
	}
}
