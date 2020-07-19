package galaxy

import (
	"bufio"
	"log"
	"os"
)

var functions map[string]expr

func Evaluate(f string) error {
	file, err := os.Open(f)
	if err != nil {
		return err
	}
	defer file.Close()
	log.Printf("Reading the interaction-protocol from %q.", f)

	functions = make(map[string]expr)

	scanner := bufio.NewScanner(file)
	for ln := 1; scanner.Scan(); ln++ {
		if err = parseDef(scanner.Bytes()); err != nil {
			log.Printf("Parse-error at line %d: %v", ln, err)
			return err
		}
	}
	if err = scanner.Err(); err != nil {
		return err
	}
	return nil
}
