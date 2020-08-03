package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"
	"strings"

	"app/galaxy"
)

var bUrl = flag.String("base_url", "https://api.pegovka.space/",
	"The base URL for the Alien Proxy Server.")

var aKeyFile = flag.String("api_key_file", "",
	"A file containing an API-key for the Alien Proxy Server.")

var flipY = flag.Bool("flip_y", false,
	"Flip the Y-axis to have the origin at bottom-left instead of top-left.")

var cProf = flag.String("cpu_profile", "",
	"Write CPU-profile to the given file.")

func readInputFile(args []string) *galaxy.FuncDefs {
	if len(args) < 1 {
		log.Fatal("Missing input file-path.")
	}
	fds, err := galaxy.ParseFunctions(args[0])
	if err != nil {
		log.Fatalf("Unable to load & parse %q: %v", args[0], err)
	}
	return fds
}

func readApiKey() string {
	if *aKeyFile == "" {
		return ""
	}
	if akf, err := ioutil.ReadFile(*aKeyFile); err == nil {
		return strings.TrimSpace(string(akf))
	} else {
		log.Fatalf("Unable to read API-key from %q: %v", *aKeyFile, err)
	}
	return ""
}

func maybeCreateGalaxyViewer() *galaxy.GalaxyViewer {
	gv := &galaxy.GalaxyViewer{FlipY: *flipY}
	if err := gv.Init(); err != nil {
		log.Fatalf("Unable to create Galaxy Viewer: %v", err)
	}
	return gv
}

// See https://blog.golang.org/pprof
func maybeCreateCpuProfile() bool {
	if *cProf == "" {
		return false
	}
	if f, err := os.Create(*cProf); err == nil {
		pprof.StartCPUProfile(f)
	} else {
		log.Fatalf("Unable to create CPU-profile in %q: %v", *cProf, err)
	}
	return true
}

func main() {
	flag.Parse()
	log.SetFlags(log.Ltime | log.Lshortfile)

	args := flag.Args()
	fds := readInputFile(args)
	aKey := readApiKey()

	gv := maybeCreateGalaxyViewer()
	if gv != nil {
		defer gv.Quit()
	}

	if maybeCreateCpuProfile() {
		defer pprof.StopCPUProfile()
	}

	ctx := &galaxy.InterCtx{
		BaseUrl:  *bUrl,
		ApiKey:   aKey,
		Protocol: fds,
		Viewer:   gv,
	}
	if err := galaxy.DoInteraction(ctx); err != nil {
		log.Fatalf("Unable to interact using %q: %v", args[0], err)
	}
}
