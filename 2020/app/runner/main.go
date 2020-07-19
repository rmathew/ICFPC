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
	f := os.Args[1]

	fds, err := galaxy.ParseFunctions(f)
	if err != nil {
		log.Fatalf("Unable to load & parse %q: %v", f, err)
	}

	if err = galaxy.DoInteraction(fds); err != nil {
		log.Fatalf("Unable to interact using %q: %v", f, err)
	}
}
