package painter

import (
	"fmt"
	"log"
)

type bruteForceSolver struct{}

func (b bruteForceSolver) String() string { return "BruteForce" }

func (b bruteForceSolver) solve(prob *Problem) (*Program, error) {
	// TODO: Implement this.
	tBlocks, err := findTargetBlocks(prob)
	if err != nil {
		return nil, err
	}
	log.Printf("Target-blocks: %s", tBlocks)
	return nil, fmt.Errorf("UNIMPLEMENTED.")
}
