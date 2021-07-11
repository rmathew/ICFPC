package main

import (
	"log"
	"os"
	"poses/squeeze"
	"time"
)

const (
	tickMs = 300 * time.Millisecond
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

	var tgtSol *squeeze.Pose
	if len(os.Args) > 2 {
		solFile := os.Args[2]
		if tgtSol, err = squeeze.ReadSolution(solFile, prob); err != nil {
			log.Fatalf("Unable to read solution-file %q: %v\n", solFile, err)
		}
		log.Printf("Read solution-file %q.\n", solFile)
	}

	v := squeeze.Viewer{}
	if err = v.Init(prob); err != nil {
		log.Fatalf("Unable to create viewer: %v", err)
	}
	defer v.Quit()

	var currSol *squeeze.Pose
	if tgtSol != nil {
		currSol = &squeeze.Pose{}
		currSol.Vertices = make([]squeeze.Point, len(prob.Figure.Vertices))
		copy(currSol.Vertices, prob.Figure.Vertices)
	}

	v.UpdateView(nil)
	const maxSteps = 16
	currStep := 0
	ok2Cont := true
	for ok2Cont {
		if currStep < maxSteps && currSol != nil {
			currStep++
			for i, v := range currSol.Vertices {
				deltaX := (tgtSol.Vertices[i].X - v.X) * int32(currStep) /
					int32(maxSteps)
				deltaY := (tgtSol.Vertices[i].Y - v.Y) * int32(currStep) /
					int32(maxSteps)
				currSol.Vertices[i].X += deltaX
				currSol.Vertices[i].Y += deltaY
			}
			v.UpdateView(currSol)
		} else {
			// OK even if `tgtSol` is nil - draws the original figure then.
			v.UpdateView(tgtSol)
		}
		time.Sleep(tickMs)
		inp, err := v.MaybeGetUserInput()
		if err != nil {
			log.Fatalf("Unable to get user-input: %v", err)
		}
		ok2Cont = !inp.Quit
	}
}
