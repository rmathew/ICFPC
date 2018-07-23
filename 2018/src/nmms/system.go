package nmms

import (
	"fmt"
)

type NmmSystem struct {
	Energy        int
	HighHarmonics bool
	Mat           Matrix
	Bots          []Nanobot
	Trc           Tracer
}

func (n *NmmSystem) ExecuteStep() error {
	numBots := len(n.Bots)
	if numBots == 0 {
		return nil
	}
	if numBots > 1 {
		return fmt.Errorf("cannot handle more than one Nanobot right now")
	}
	cmds, err := n.Trc.TakeCommands(numBots)
	if err != nil {
		return err
	}
	if len(cmds) == 0 {
		return nil
	}

	energyCost := 0
	resCubed := n.Mat.Resolution() * n.Mat.Resolution() * n.Mat.Resolution()
	if n.HighHarmonics {
		energyCost += 30 * resCubed
	} else {
		energyCost += 3 * resCubed
	}
	energyCost += 20 * len(n.Bots)

	if err = cmds[0].Execute(n, 0); err != nil {
		return err
	}
	n.Energy += energyCost

	return nil
}
