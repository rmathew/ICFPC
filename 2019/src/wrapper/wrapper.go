// Usage: go run wrapper.go /path/to/prob.desc
package main

import (
	"fmt"
	"os"

	"wwabr"
)

func main() {
	if len(os.Args) < 2 {
		wwabr.ExitWithErrorMsg(
			"Missing problem-description file-path argument.")
	}
	fmt.Printf("Reading problem-description file \"%s\".\n", os.Args[1])
	_, err := wwabr.NewFromFile(os.Args[1])
	wwabr.Check(err)
}
