package main

import (
	"aedificium/lol"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

var idFile = flag.String("id_file", "", "file containing the team identifier.")

// This was 18 during the Lightning round; changed to 6 afterward.
var maxPathFactor = flag.Int("max_path_factor", 18,
	"scaling factor for the allowed path-length based on the problem-size.")

type cmdLineInfo struct {
	teamId        string
	maxPathFactor int
	probName      string
	probSize      int
}

func updateUsage() {
	flag.Usage = func() {
		o := flag.CommandLine.Output()
		p := os.Args[0]
		fmt.Fprintf(o, "Usage of %s: %s [options] <problem> <size>\n", p, p)
		flag.PrintDefaults()
	}
}

func readTeamId(file string) string {
	idBytes, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("ERROR: Failed to read team-id file: %v", err)
	}
	return strings.TrimSpace(string(idBytes))
}

func parseCommandLine(cli *cmdLineInfo) error {
	updateUsage()
	flag.Parse()

	if *idFile == "" {
		flag.Usage()
		log.Fatal("ERROR: Missing team-identifier; use '--id_file'.")
	}
	cli.teamId = readTeamId(*idFile)
	if *maxPathFactor < 1 {
		return fmt.Errorf("maximum path factor must be positive: %d",
			*maxPathFactor)
	}
	cli.maxPathFactor = *maxPathFactor

	if flag.NArg() != 2 {
		return fmt.Errorf("need exactly two arguments (got %d).", flag.NArg())
	}
	cli.probName = flag.Arg(0)
	var err error
	cli.probSize, err = strconv.Atoi(flag.Arg(1))
	if err != nil {
		return fmt.Errorf("bad number of rooms: %w", err)
	}
	if cli.probSize < 1 {
		return fmt.Errorf("number of rooms must be positive: %d", cli.probSize)
	}

	return nil
}

func main() {
	log.SetFlags(0) // We just use `log` as a convenient printer.

	var cli cmdLineInfo
	if err := parseCommandLine(&cli); err != nil {
		flag.Usage()
		log.Fatalf("ERROR: %v", err)
	}

	log.Printf("INFO: Trying to solve the problem '%s'...", cli.probName)
	c := lol.NewClient(cli.teamId)
	m := lol.NewMapper(c, cli.maxPathFactor)
	if err := m.Map(cli.probName, cli.probSize); err != nil {
		log.Fatalf("ERROR: Can't solve the problem '%s': %v", cli.probName, err)
	}
}
