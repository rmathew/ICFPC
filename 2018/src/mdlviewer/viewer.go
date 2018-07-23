// Usage: go run viewer.go /path/to/file.mdl
package main

import (
	"fmt"
	"os"
	"time"

	"nmms"
)

func main() {
	if len(os.Args) < 2 {
		nmms.ExitWithErrorMsg("Missing Model argument.")
	}
	var nSys nmms.NmmSystem
	fmt.Printf("Reading Model file \"%s\".\n", os.Args[1])
	nmms.Check(nSys.Mat.ReadFromFile(os.Args[1]))
	res := nSys.Mat.Resolution()
	fmt.Printf("Resolution=%d; ", res)
	numFilled := 0
	for x := 0; x < res; x++ {
		for y := 0; y < res; y++ {
			for z := 0; z < res; z++ {
				if nSys.Mat.IsFull(x, y, z) {
					numFilled++
				}
			}
		}
	}
	fmt.Printf("Filled=%d\n", numFilled)

	var viewer nmms.Renderer
	nmms.Check(viewer.Init(&nSys))
	for viewer.Update(&nSys) {
		time.Sleep(250 * time.Millisecond)
	}
	viewer.Quit()
}
