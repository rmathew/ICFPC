package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"robovinci/painter"
	"strings"
)

var inProg = flag.String("inprog", "", "file for reading in a program")
var outProg = flag.String("outprog", "", "file for writing out the program")
var outImg = flag.String("outimg", "", "file for rendering the solution")

func updateUsage() {
	flag.Usage = func() {
		o := flag.CommandLine.Output()
		p := os.Args[0]
		fmt.Fprintf(o, "Usage of %s: %s [options] target_painting\n", p, p)
		flag.PrintDefaults()
	}
}

func validateArgs() {
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatalf("Unexpected number of arguments: %d (need 1).", flag.NArg())
	}
	if len(*inProg) > 0 && len(*outProg) > 0 {
		log.Fatalf("--inprog and --outprog are mutually exclusive.")
	}
	if len(*outProg) == 0 && len(*outImg) == 0 {
		log.Fatalf("Nothing to do - neither --outprog nor --outimg specified.")
	}
}

func readProblem() (*painter.Problem, error) {
	if flag.NArg() == 0 {
		return nil, fmt.Errorf("missing target painting")
	}
	tP := strings.TrimSpace(flag.Arg(0))
	log.Printf("Reading target painting %q...", tP)
	return painter.ReadProblem(tP)
}

func getProgram(prob *painter.Problem) (*painter.Program, error) {
	if len(*inProg) > 0 {
		return painter.ReadProgram(*inProg)
	}
	return painter.SolveProblem(prob)
}

func putProgram(prob *painter.Problem, prog *painter.Program) error {
	if len(*outProg) > 0 {
		// TODO: Save the program.
		// err := writeProgram()
	}
	if len(*outImg) > 0 {
		log.Printf("Interpreting program")
		res, err := painter.InterpretProgram(prob, prog)
		if err != nil {
			return err
		}
		log.Printf("Program completed successfully with score %d.", res.Score)
		log.Printf("Saving the result to %q", *outImg)
		err = painter.RenderResult(res, *outImg)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	updateUsage()
	validateArgs()

	prob, err := readProblem()
	if err != nil {
		log.Fatalf("Unable to read the problem: %v", err)
	}

	prog, err := getProgram(prob)
	if err != nil {
		log.Fatalf("Unable to get the program: %v", err)
	}

	err = putProgram(prob, prog)
	if err != nil {
		log.Fatalf("Unable to put the program: %v", err)
	}
}
