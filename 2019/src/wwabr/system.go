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
	MineMap   Map
	WorkerLoc Point
}

func NewFromFile(p string) (*WwabrSystem, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, maxInputLineSize), maxInputLineSize)

	var wSys WwabrSystem
	done := false
	for s.Scan() {
		if done {
			return nil, errors.New("Extra line in problem-description.")
		}
		t := strings.Split(s.Text(), "#")
		if len(t) != 4 {
			return nil, fmt.Errorf("Got %d tuples instead of 4.", len(t))
		}
		if wSys.MineMap, err = makeMap(t[0]); err != nil {
			return nil, err
		}
		if wSys.WorkerLoc, err = parsePoint(t[1]); err != nil {
			return nil, err
		}
		if err = populateObstacles(wSys.MineMap, t[2]); err != nil {
			return nil, err
		}
		printMap(wSys.MineMap)
		done = true
	}

	return &wSys, nil
}
