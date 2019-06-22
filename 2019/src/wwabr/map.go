package wwabr

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
)

type MapCell int
type Map [][]MapCell

const (
	INVALID_CELL MapCell = iota
	EMPTY_CELL
	OBSTACLE_CELL
	WRAPPED_CELL
)

const (
	pointRegEx = "\\( *[0-9]+ *, *[0-9]+ *\\)"
)

type Point struct {
	X, Y int
}

type rectilinearPolygon struct {
	points                 []Point
	minX, minY, maxX, maxY int
}

func makeMap(s string) (Map, error) {
	rp, err := parseRectilinearPolygon(s)
	if err != nil {
		return nil, err
	}
	m := make([][]MapCell, rp.maxY)
	for i := 0; i < rp.maxY; i++ {
		m[i] = make([]MapCell, rp.maxX)
	}
	fillRectilinearPolygon(m, rp, INVALID_CELL, EMPTY_CELL)
	printMap(m)
	return m, nil
}

func printMap(m Map) {
	if len(m) > 75 || len(m[0]) > 75 {
		return
	}
	for y := len(m) - 1; y >= 0; y-- {
		for x := 0; x < len(m[y]); x++ {
			switch m[y][x] {
			case EMPTY_CELL:
				fmt.Print(".")
			case OBSTACLE_CELL:
				fmt.Print("#")
			case WRAPPED_CELL:
				fmt.Print("*")
			default:
				fmt.Print("#")
			}
		}
		fmt.Print("\n")
	}
}

func fillRectilinearPolygon(m Map, rp rectilinearPolygon, mc0, mc1 MapCell) {
	for i := 1; i <= len(rp.points); i++ {
		drawContour(m, rp.points[i-1], rp.points[i%len(rp.points)], mc1)
	}
	for y := rp.minY; y < rp.maxY; y++ {
		prvMc := mc0
		for x := rp.minX; x < rp.maxX; x++ {
			curMc := m[y][x]
			// TODO: Flood-fill the polygon.
			prvMc = curMc
		}
	}
}

func drawContour(m Map, p0, p1 Point, mc MapCell) {
	var xOff, yOff = 0, 0
	if p0.X == p1.X {
		if p0.Y < p1.Y {
			xOff = -1
		}
		for y := iMin(p0.Y, p1.Y); y < iMax(p0.Y, p1.Y); y++ {
			m[y][p0.X+xOff] = mc
		}
	} else if p0.Y == p1.Y {
		if p0.X > p1.X {
			yOff = -1
		}
		for x := iMin(p0.X, p1.X); x < iMax(p0.X, p1.X); x++ {
			m[p0.Y+yOff][x] = mc
		}
	}
}

func parseRectilinearPolygon(s string) (rectilinearPolygon, error) {
	var rp rectilinearPolygon
	n := regexp.MustCompile(pointRegEx).FindAllString(s, -1)
	if n == nil || len(n) == 0 {
		return rp, fmt.Errorf("Invalid list of Points.")
	}
	rp.points = make([]Point, len(n))
	rp.minX, rp.minY = math.MaxInt32, math.MaxInt32
	rp.maxX, rp.maxY = math.MinInt32, math.MinInt32
	var pt Point
	var err error
	for i, v := range n {
		if pt, err = parsePoint(v); err != nil {
			return rp, err
		}
		rp.minX = iMin(rp.minX, pt.X)
		rp.minY = iMin(rp.minY, pt.Y)
		rp.maxX = iMax(rp.maxX, pt.X)
		rp.maxY = iMax(rp.maxY, pt.Y)
		if i > 0 {
			ppt := rp.points[i-1]
			if ppt.X != pt.X && ppt.Y != pt.Y {
				return rp, fmt.Errorf("Non-rectilinear line from %v to %v.",
					ppt, pt)
			}
		}
		rp.points[i] = pt
	}
	return rp, nil
}

func parsePoint(s string) (Point, error) {
	var p Point
	m, err := regexp.MatchString(pointRegEx, s)
	if err != nil {
		return p, err
	}
	if !m {
		return p, fmt.Errorf("Invalid Point \"%s\".", s)
	}
	n := regexp.MustCompile("[0-9]+").FindAllString(s, -1)
	if len(n) != 2 {
		return p, fmt.Errorf("Invalid Point \"%s\".", s)
	}
	if p.X, err = strconv.Atoi(n[0]); err != nil {
		return p, err
	}
	if p.Y, err = strconv.Atoi(n[1]); err != nil {
		return p, err
	}
	return p, nil
}

func (p *Point) String() string {
	return fmt.Sprintf("(%d,%d)", p.X, p.Y)
}
