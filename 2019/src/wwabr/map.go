package wwabr

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type MapCell int
type Map [][]MapCell

const (
	InvalidCell MapCell = iota
	EmptyCell
	ObstacleCell
	WrappedCell
)

type lineDir int

const (
	leftToRight lineDir = iota
	rightToLeft
	bottomToTop
	topToBottom
)

const (
	pointRegEx = "\\( *[0-9]+ *, *[0-9]+ *\\)"
)

type Point struct {
	X, Y int
}

type rlPolygon struct {
	points                 []Point
	minX, minY, maxX, maxY int
}

func makeMap(s string) (Map, error) {
	rp, err := parseRlPolygon(s)
	if err != nil {
		return nil, err
	}
	m := make([][]MapCell, rp.maxY)
	for i := 0; i < rp.maxY; i++ {
		m[i] = make([]MapCell, rp.maxX)
	}
	fillRlPolygon(m, rp, InvalidCell, EmptyCell)
	return m, nil
}

func populateObstacles(m Map, s string) error {
	if len(s) == 0 {
		return nil
	}
	t := strings.Split(s, ";")
	for _, v := range t {
		rp, err := parseRlPolygon(v)
		if err != nil {
			return err
		}
		fillRlPolygon(m, rp, EmptyCell, ObstacleCell)
	}
	return nil
}

func printMap(m Map) {
	// ANSI escape-sequences via:
	// http://www.lihaoyi.com/post/BuildyourownCommandLinewithANSIescapecodes.html
	fmt.Printf("Map-Size: %dx%d\n", len(m[0]), len(m))
	if len(m) > 75 || len(m[0]) > 75 {
		return
	}
	for y := len(m) - 1; y >= 0; y-- {
		for x := 0; x < len(m[y]); x++ {
			switch m[y][x] {
			case EmptyCell:
				fmt.Print("\u001b[47m \u001b[0m") // BackgroundWhite
			case ObstacleCell:
				fmt.Print("\u001b[41m \u001b[0m") // BackgroundRed
			case WrappedCell:
				fmt.Print("\u001b[44m \u001b[0m") // BackgroundBlue
			default:
				fmt.Print("\u001b[40m \u001b[0m") // BackgroundBlack
			}
		}
		fmt.Print("\n")
	}
}

func fillRlPolygon(m Map, rp rlPolygon, mc0, mc1 MapCell) {
	for i := 1; i <= len(rp.points); i++ {
		drawContour(m, rp.points[i-1], rp.points[i%len(rp.points)], mc1)
	}
	if true {
		fmt.Println("Contoured")
		printMap(m)
	} // DEBUG
	sps := pickSeedPoints(m, rp, mc0, mc1)
	// DEBUG
	if true {
		for _, s := range sps {
			m[s.Y][s.X] = WrappedCell
		}
		fmt.Printf("Seeded %s\n", sps)
		printMap(m)
		for _, s := range sps {
			m[s.Y][s.X] = mc0
		}
	}
	for _, s := range sps {
		floodFill(m, s, mc0, mc1)
		if true {
			fmt.Printf("Seed %s\n", s)
			printMap(m)
		} // DEBUG
	}
	return
}

func floodFill(m Map, sp Point, mc0, mc1 MapCell) {
	// A transcription of the flood-fill algorithm from:
	//   https://en.wikipedia.org/wiki/Flood_fill#Alternative_implementations
	if mc0 == mc1 || m[sp.Y][sp.X] != mc0 {
		return
	}
	m[sp.Y][sp.X] = mc1
	q := make([]Point, 1, 1024)
	q[0] = sp
	mLenX, mLenY := len(m[0]), len(m)
	var n Point
	for len(q) > 0 {
		n, q = q[0], q[1:] // Pop
		x, y := n.X-1, n.Y // West
		if x >= 0 && y >= 0 && x < mLenX && y < mLenY && m[y][x] == mc0 {
			m[y][x] = mc1
			q = append(q, Point{x, y})
		}
		x, y = n.X+1, n.Y // East
		if x >= 0 && y >= 0 && x < mLenX && y < mLenY && m[y][x] == mc0 {
			m[y][x] = mc1
			q = append(q, Point{x, y})
		}
		x, y = n.X, n.Y+1 // North
		if x >= 0 && y >= 0 && x < mLenX && y < mLenY && m[y][x] == mc0 {
			m[y][x] = mc1
			q = append(q, Point{x, y})
		}
		x, y = n.X, n.Y-1 // South
		if x >= 0 && y >= 0 && x < mLenX && y < mLenY && m[y][x] == mc0 {
			m[y][x] = mc1
			q = append(q, Point{x, y})
		}
	}
}

func pickSeedPoints(m Map, rp rlPolygon, mc0, mc1 MapCell) []Point {
	n := len(rp.points)
	p := make([]Point, 0, n)
	if n < 4 || rp.maxX-rp.minX < 3 || rp.maxY-rp.minY < 3 {
		return p
	}
	for i := 2; i <= n; i++ {
		p0, p1, p2 := rp.points[i-2], rp.points[i-1], rp.points[i%n]
		if segLen(p0, p1) < 2 || segLen(p1, p2) < 2 || !isLeftBend(p0, p1, p2) {
			continue
		}
		xOff, yOff := getLeftOff(p0, p1, 1)
		x, y := p1.X + xOff, p1.Y + yOff
		switch ld := getLineDir(p0, p1); ld {
		case leftToRight:
			x -= 2
		case rightToLeft:
			x += 1
		case bottomToTop:
			y -= 2
		case topToBottom:
			y += 1
		}
		if isValidSeedPoint(m, rp, mc0, mc1, x, y) {
			p = append(p, Point{x, y})
		}
	}
	return p
}

func segLen(p0, p1 Point) int {
	if p0.X == p1.X {
		return iAbs(p0.Y - p1.Y)
	}
	return iAbs(p0.X - p1.X)
}

func isLeftBend(p0, p1, p2 Point) bool {
	switch ld01, ld12 := getLineDir(p0, p1), getLineDir(p1, p2); ld01 {
	case leftToRight:
		return ld12 == bottomToTop
	case rightToLeft:
		return ld12 == topToBottom
	case bottomToTop:
		return ld12 == rightToLeft
	default:
		return ld12 == leftToRight
	}
}

func isValidSeedPoint(m Map, rp rlPolygon, mc0, mc1 MapCell, x, y int) bool {
	if x < 0 || x > rp.maxX || y < 0 || y > rp.maxY {
		return false
	}
	if m[y][x] != mc0 {
		return false
	}
	return true
}

func drawContour(m Map, p0, p1 Point, mc MapCell) {
	ld := getLineDir(p0, p1)
	xOff, yOff := getLeftOff(p0, p1, 0)
	if ld == bottomToTop || ld == topToBottom {
		x := p0.X + xOff
		minY, maxY := iMin(p0.Y, p1.Y), iMax(p0.Y, p1.Y)
		for y := minY; y < maxY; y++ {
			m[y+yOff][x] = mc
		}
	} else if ld == leftToRight || ld == rightToLeft {
		minX, maxX := iMin(p0.X, p1.X), iMax(p0.X, p1.X)
		y := p0.Y + yOff
		for x := minX; x < maxX; x++ {
			m[y][x+xOff] = mc
		}
	}
}

func getLeftOff(p0, p1 Point, delta int) (xOff, yOff int) {
	switch ld := getLineDir(p0, p1); ld {
	case leftToRight:
		yOff = delta
	case rightToLeft:
		yOff = -1 - delta
	case bottomToTop:
		xOff = -1 - delta
	case topToBottom:
		xOff = delta
	}
	return
}

func getLineDir(p0, p1 Point) lineDir {
	switch {
	case p0.X == p1.X && p0.Y < p1.Y:
		return bottomToTop
	case p0.X == p1.X && p0.Y > p1.Y:
		return topToBottom
	case p0.Y == p1.Y && p0.X < p1.X:
		return leftToRight
	default:
		return rightToLeft
	}
}

func parseRlPolygon(s string) (rlPolygon, error) {
	var rp rlPolygon
	n := regexp.MustCompile(pointRegEx).FindAllString(s, -1)
	if n == nil || len(n) == 0 {
		return rp, fmt.Errorf("Invalid list of Points '%s'.", n)
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
			if ppt.X == pt.X && ppt.Y == pt.Y {
				return rp, fmt.Errorf("Got point %v instead of a line.", pt)
			}
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

func (p Point) String() string {
	return fmt.Sprintf("(%d,%d)", p.X, p.Y)
}
