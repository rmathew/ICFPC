package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"poses/squeeze"
	"strings"
	"time"
)

var inpSol = flag.String("inp-sol", "", "file for reading in a solution")
var outSol = flag.String("out-sol", "", "file for writing out the best solution")

type bestSol struct {
	sol   *squeeze.Pose
	score int32
}

func readProblem() (*squeeze.Problem, error) {
	if flag.NArg() == 0 {
		return nil, fmt.Errorf("missing input problem-file")
	}
	pF := flag.Arg(0)
	log.Printf("Reading problem-file %q...", pF)
	p, err := squeeze.ReadProblem(pF)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func maybeReadSolution(prob *squeeze.Problem) (*squeeze.Pose, error) {
	sF := strings.TrimSpace(*inpSol)
	if len(sF) == 0 {
		return nil, nil
	}
	log.Printf("Reading solution-file %q...", sF)
	s, err := squeeze.ReadSolution(sF, prob)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func maybeWriteBestSol(best *bestSol, sol *squeeze.Pose,
	prob *squeeze.Problem) error {
	d := squeeze.GetDislikes(sol, prob)
	vs := "INVALID"
	ok := squeeze.IsValidSolution(sol, prob)
	if ok {
		vs = "valid"
	}
	log.Printf("Found %s solution with dislikes=%d", vs, d)
	if !ok || d > best.score {
		return nil
	}
	best.sol = sol
	best.score = d
	sF := strings.TrimSpace(*outSol)
	if len(sF) == 0 {
		return nil
	}
	if err := squeeze.WriteSolution(sol, sF); err != nil {
		return err
	}
	log.Printf("New best solution with dislikes=%d saved in %q.", d, sF)
	return nil
}

func main() {
	flag.Parse()

	prob, err := readProblem()
	if err != nil {
		log.Fatalf("Unable to read the problem: %v", err)
	}

	var tgtSol *squeeze.Pose
	if tgtSol, err = maybeReadSolution(prob); err != nil {
		log.Fatalf("Unable to read the solution: %v", err)
	}

	solver := squeeze.NewSolver(prob, tgtSol)
	solver.Reset()

	v := squeeze.Viewer{}
	if err = v.Init(prob); err != nil {
		log.Fatalf("Unable to create the viewer: %v", err)
	}
	defer v.Quit()
	v.UpdateView(nil)

	var inp *squeeze.UserInput
	best := bestSol{nil, math.MaxInt32}
	gotIt := false
	ok2Cont := true
	run := false
	for ok2Cont {
		t0 := time.Now()
		if run {
			sol := solver.GetNextSolution()
			v.UpdateView(sol)
			if !gotIt && solver.WasFinalSolution() {
				if err = maybeWriteBestSol(&best, sol, prob); err != nil {
					log.Fatalf("Unable to write the solution: %v", err)
				}
				gotIt = true
			}
		}

		const tickMs = 300 * time.Millisecond
		if dur := time.Since(t0); dur < tickMs {
			time.Sleep(tickMs - dur)
		}

		if inp, err = v.MaybeGetUserInput(); err != nil {
			log.Fatalf("Unable to get user-input: %v", err)
		}
		ok2Cont = !inp.Quit
		run = inp.Run
		if inp.Reset {
			solver.Reset()
			v.UpdateView(nil)
			gotIt = false
		}
	}

	if !gotIt {
		log.Printf("Could not find a solution.")
	}
}
