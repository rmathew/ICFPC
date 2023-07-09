package main

import (
	"concert/plan"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var outSoln = flag.String("outsoln", "", "file for storing the solution")

func updateUsage() {
	flag.Usage = func() {
		o := flag.CommandLine.Output()
		p := os.Args[0]
		fmt.Fprintf(o, "Usage of %s: %s [options] problem_spec\n", p, p)
		flag.PrintDefaults()
	}
}

func validateArgs() {
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatalf("Invalid number of arguments: %d (need 1).", flag.NArg())
	}
}

func readProblem() (*plan.Problem, error) {
	if flag.NArg() == 0 {
		return nil, fmt.Errorf("missing problem specification")
	}
	tP := strings.TrimSpace(flag.Arg(0))
	prob, err := plan.ReadProblem(tP)
	if err != nil {
		return nil, err
	}
	log.Printf("Room: %q", &prob.Room)
	log.Printf("Stage: %q", &prob.Stage)
	log.Printf("%d musicians will be playing for %d attendees.",
		len(prob.Musicians), len(prob.Attendees))
	return prob, nil
}

func maybeWriteSolution(prob *plan.Problem, soln *plan.Solution) error {
	sF := strings.TrimSpace(*outSoln)
	if len(sF) == 0 {
		log.Print("No file specified for saving the solution.")
		return nil
	}
	if err := plan.WriteSolution(soln, sF); err != nil {
		return err
	}
	log.Printf("Solution saved in %q.", sF)
	return nil
}

func main() {
	log.SetFlags(0) // We just use `log` as a convenient printer.

	updateUsage()
	validateArgs()

	prob, err := readProblem()
	if err != nil {
		log.Fatalf("Unable to read the problem: %v", err)
	}

	solver := plan.NewSolver(prob)
	solver.Reset()
	gotIt := false
	for i := 0; i < 100000; i++ {
		soln := solver.GetNextSolution()
		if solver.WasFinalSolution() {
			log.Printf("Found a solution with score %0.2f.", soln.Score)
			if err = maybeWriteSolution(prob, soln); err != nil {
				log.Fatalf("Unable to write out the solution: %v", err)
			}
			gotIt = true
			break
		}
	}
	if !gotIt {
		log.Fatal("Could not find a solution.")
	}
}
