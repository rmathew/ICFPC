package main

import (
	"aedificium/lol"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var idFile = flag.String("id_file", "", "file containing the team identifier.")

type cmdLineInfo struct {
	teamId   string
	probName string
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

	if flag.NArg() != 1 {
		return fmt.Errorf("need exactly one argument (got %d).", flag.NArg())
	}
	cli.probName = flag.Arg(0)

	return nil
}

func main() {
	log.SetFlags(0) // We just use `log` as a convenient printer.

	var cli cmdLineInfo
	if err := parseCommandLine(&cli); err != nil {
		flag.Usage()
		log.Fatalf("ERROR: %v", err)
	}

	c := lol.NewClient(cli.teamId)
	m := lol.NewMapper(c)
	if err := m.Map(cli.probName); err != nil {
		log.Fatalf("ERROR: Can't solve the problem '%s': %v", cli.probName, err)
	}
}
