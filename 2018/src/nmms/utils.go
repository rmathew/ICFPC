// Package nmms contains libraries for the Nanobot Matter Manipulation System
// simulator.
package nmms

import (
	"fmt"
	"os"
)

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
