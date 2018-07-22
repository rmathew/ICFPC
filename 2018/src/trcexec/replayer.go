package main

import (
	"fmt"
	"os"

	"nmms"
)

func main() {
	if len(os.Args) < 3 {
		nmms.ExitWithErrorMsg("Missing Trace and target-Model argument.")
	}
	var nSys nmms.NmmSystem
	nSys.Bots = make([]nmms.Nanobot, 1)
	nSys.Bots[0].Pos = nmms.Coordinate{}

	fmt.Printf("Reading Trace file \"%s\".\n", os.Args[1])
	nmms.Check(nSys.Trc.ReadFromFile(os.Args[1]))
	fmt.Printf("Reading Model file \"%s\".\n", os.Args[2])
	nmms.Check(nSys.Mat.ReadFromFile(os.Args[2]))
	nSys.Mat.Clear()
	res := nSys.Mat.Resolution()
	fmt.Printf("Resolution=%d\n", res)

	var viewer nmms.Renderer
	nmms.Check(viewer.Init(&nSys))
	for viewer.Update(&nSys) {
		if err := nSys.ExecuteStep(); err != nil {
			fmt.Printf("ERROR: %v\n", err)
			break
		}
	}
	viewer.Quit()
}
