package main

import (
	"fmt"
	"log"
	"os"
	"poses/squeeze"
	"time"
)

func readProblem() (*squeeze.Problem, error) {
	if len(os.Args) < 2 {
		return nil, fmt.Errorf("missing input problem-file")
	}
	pF := os.Args[1]
	log.Printf("Reading problem-file %q...", pF)
	p, err := squeeze.ReadProblem(pF)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func maybeReadSolution(prob *squeeze.Problem) (*squeeze.Pose, error) {
	if len(os.Args) <= 2 {
		return nil, nil
	}
	sF := os.Args[2]
	log.Printf("Reading solution-file %q...", sF)
	s, err := squeeze.ReadSolution(sF, prob)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func main() {
	prob, err := readProblem()
	if err != nil {
		log.Fatalf("Unable to read the problem: %v", err)
	}

	var tgtSol *squeeze.Pose
	if tgtSol, err = maybeReadSolution(prob); err != nil {
		log.Fatalf("Unable to read the solution: %v", err)
	}

	solver := squeeze.NewSolver(prob, tgtSol)
	solver.InitSolver()

	v := squeeze.Viewer{}
	if err = v.Init(prob); err != nil {
		log.Fatalf("Unable to create the viewer: %v", err)
	}
	defer v.Quit()
	v.UpdateView(nil)

	var inp *squeeze.UserInput
	gotIt := false
	ok2Cont := true
	for ok2Cont {
		t0 := time.Now()
		sol := solver.GetNextSolution()
		v.UpdateView(sol)
		if !gotIt && solver.WasFinalSolution() {
			d := squeeze.GetDislikes(sol, prob)
			log.Printf("Found a solution with dislikes: %d", d)
			gotIt = true
		}

		const tickMs = 300 * time.Millisecond
		if dur := time.Since(t0); dur < tickMs {
			time.Sleep(tickMs - dur)
		}

		if inp, err = v.MaybeGetUserInput(); err != nil {
			log.Fatalf("Unable to get user-input: %v", err)
		}
		ok2Cont = !inp.Quit
	}

	if !gotIt {
		log.Printf("Could not find a solution.")
	}
}
