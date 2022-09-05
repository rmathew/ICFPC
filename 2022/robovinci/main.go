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
}

func readProblem() (*painter.Problem, error) {
	if flag.NArg() == 0 {
		return nil, fmt.Errorf("missing target painting")
	}
	tP := strings.TrimSpace(flag.Arg(0))
	return painter.ReadProblem(tP)
}

func getProgram(prob *painter.Problem) (*painter.Program, error) {
	if len(*inProg) > 0 {
		return painter.ReadProgram(*inProg)
	}
	return painter.SolveProblem(prob)
}

func execProgram(prob *painter.Problem, prog *painter.Program) (*painter.ExecResult, error) {
	log.Printf("Executing program...")
	res, err := painter.InterpretProgram(prob, prog)
	if err == nil {
		log.Printf("Program completed successfully with score %d.", res.Score)
	}
	return res, err
}

func maybeSaveOutput(prog *painter.Program, res *painter.ExecResult) error {
	if len(*outProg) > 0 {
		log.Printf("Saving the generated program to %q", *outProg)
		err := painter.SaveProgram(prog, res, *outProg)
		if err != nil {
			return err
		}
	}
	if len(*outImg) > 0 {
		log.Printf("Saving the rendered result to %q", *outImg)
		err := painter.RenderResult(res, *outImg)
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

	res, err := execProgram(prob, prog)
	if err != nil {
		log.Fatalf("Unable to execute the program: %v", err)
	}

	err = maybeSaveOutput(prog, res)
	if err != nil {
		log.Fatalf("Unable to save the output: %v", err)
	}
}
