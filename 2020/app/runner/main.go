package main

import (
	"log"
	"flag"

	"app/galaxy"
)

var bUrl = flag.String("base_url", "https://icfpc2020-api.testkontur.ru/",
	"Base URL for the Alien Proxy server.");
var aKey = flag.String("api_key", "", "API-key for the Alien Proxy server.");

func main() {
	flag.Parse()
	log.SetFlags(log.Ltime | log.Lshortfile)

	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("Missing input file-path.")
	}
	f := args[0]

	fds, err := galaxy.ParseFunctions(f)
	if err != nil {
		log.Fatalf("Unable to load & parse %q: %v", f, err)
	}

	ctx := &galaxy.InterCtx{BaseUrl: *bUrl, ApiKey: *aKey, Protocol: fds}
	if err = galaxy.DoInteraction(ctx); err != nil {
		log.Fatalf("Unable to interact using %q: %v", f, err)
	}
}
