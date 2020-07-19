package galaxy

import (
	"log"
)

func Evaluate(f string) error {
	var funcDefs map[string]expr
	var err error
	if funcDefs, err = parse(f); err != nil {
		return err
	}
	log.Printf("Found %d func-def(s) in %q.", len(funcDefs), f)

	return nil
}
