package wwabr

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	maxInputLineSize = 256 * 1024
)

type WwabrSystem struct {
}

func NewFromFile(p string) (*WwabrSystem, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, maxInputLineSize), maxInputLineSize)
	d := false
	for s.Scan() {
		if d {
			return nil, errors.New("Extra line in problem-description.")
		}
		t := strings.Split(s.Text(), "#")
		if len(t) != 4 {
			return nil, fmt.Errorf("Got %d tuples instead of 4.", len(t))
		}
		// BEGIN: DEBUG
		var m Map
		if m, err = makeMap(t[0]); err != nil {
			return nil, err
		}
		fmt.Printf("Map dimensions %dx%d\n", len(m[0]), len(m))
		// END: DEBUG
		d = true
	}

	return &WwabrSystem{}, nil
}
