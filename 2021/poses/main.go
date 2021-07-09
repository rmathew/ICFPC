package main

import (
	"log"
	"os"
	"poses/squeeze"
	"time"
)

const (
	tickMs = 100 * time.Millisecond
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Missing input problem-file.\n")
		os.Exit(1)
	}
	probFile := os.Args[1]
	prob, err := squeeze.ReadProblem(probFile)
	if err != nil {
		log.Fatalf("Unable to read problem-file %q: %v\n", probFile, err)
	}
	log.Printf("Read problem-file %q.\n", probFile)

	v := squeeze.Viewer{}
	if err = v.Init(); err != nil {
		log.Fatalf("Unable to create viewer: %v", err)
	}
	defer v.Quit()

	var sol squeeze.Pose
	sol.Vertices = prob.Figure.Vertices
	goOn := true
	for goOn {
		v.UpdateView(prob, &sol)
		time.Sleep(tickMs)
		inp := v.MaybeGetUserInput()
		goOn = !inp.Quit
	}
}
