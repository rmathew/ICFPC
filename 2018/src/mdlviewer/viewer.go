package main

import (
	"fmt"
	"os"

	"nmms"
)

func main() {
	if len(os.Args) < 2 {
		nmms.ExitWithErrorMsg("Missing Model argument.")
	}
	fmt.Printf("Reading Model file \"%s\".\n", os.Args[1])
	mat := new(nmms.Matrix)
	nmms.Check(mat.ReadFromFile(os.Args[1]))
	res := mat.Resolution()
	fmt.Printf("Resolution=%d; ", res)
	numFilled := 0
	for x := 0; x < res; x++ {
		for y := 0; y < res; y++ {
			for z := 0; z < res; z++ {
				full, err := mat.IsFull(x, y, z)
				nmms.Check(err)
				if full {
					numFilled++
				}
			}
		}
	}
	fmt.Printf("Filled=%d\n", numFilled)

	viewer := new(nmms.Renderer)
	nmms.Check(viewer.Init())
	viewer.Update(mat)
	viewer.Quit()
}
