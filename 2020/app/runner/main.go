package main

import (
	"log"
	"os"

	"app/galaxy"
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	if len(os.Args) < 2 {
		log.Fatal("Missing input file-path.")
	}

	if err := galaxy.Evaluate(os.Args[1]); err != nil {
		log.Fatalf("Unable to evaluate %q: %v", os.Args[1], err)
	}
}
