package main

import (
	"flag"
	"io/ioutil"
	"log"
	"strings"

	"app/galaxy"
)

var bUrl = flag.String("base_url", "https://icfpc2020-api.testkontur.ru/",
	"The base URL for the Alien Proxy server.")

var aKeyFile = flag.String("api_key_file", "",
	"A file containing an API-key for the Alien Proxy server.")

var flipY = flag.Bool("flip_y", false,
	"Flip the Y-axis to have the origin at bottom-left instead of top-left.")

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

	var akf []byte
	if akf, err = ioutil.ReadFile(*aKeyFile); err != nil {
		log.Fatalf("Unable to read API-key from %q: %v", aKeyFile, err)
	}
	aKey := strings.TrimSpace(string(akf))

	gv := &galaxy.GalaxyViewer{FlipY: *flipY}
	if err = gv.Init(); err != nil {
		log.Fatalf("Unable to create Galaxy Viewer: %v", err)
	}
	defer gv.Quit()

	ctx := &galaxy.InterCtx{
		BaseUrl:  *bUrl,
		ApiKey:   aKey,
		Protocol: fds,
		Viewer:   gv,
	}
	if err = galaxy.DoInteraction(ctx); err != nil {
		log.Fatalf("Unable to interact using %q: %v", f, err)
	}
}
