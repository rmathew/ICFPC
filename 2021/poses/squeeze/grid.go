package squeeze

import (
	"encoding/json"
	"fmt"
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
	"io/ioutil"
	"log"
	"math"
	"unsafe"
)

const (
	epsilonScale float64 = 1000000.0
)

type Point struct {
	X, Y int32
}

type Polygon struct {
	Vertices []Point
}

type Line struct {
	StartIdx, EndIdx int
}

type Graph struct {
	Vertices []Point
	Edges    []Line
}

type figVertEdgeInfo struct {
	idx              int
	minDist, maxDist int32
}

type preProcessedInfo struct {
	low, high             Point
	holeLow, holeHigh     Point
	figLow, figHigh       Point
	holeWidth, holeHeight int32
	cellsWithHoles        []bool

	minEdgeScale, maxEdgeScale float64
	figVertEdges               [][]figVertEdgeInfo
}

type Problem struct {
	Hole    Polygon
	Figure  Graph
	Epsilon float64

	preProc preProcessedInfo
}

type Pose struct {
	Vertices []Point
}

type poseViolations struct {
	vertsOutsideHole []int
	numStrayingEdges int
}

type orientation int

const (
	clockwise        orientation = -1
	collinear        orientation = 0
	counterClockwise orientation = +1
)

func (p Point) String() string {
	return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}

func sqDist(p, q Point) int32 {
	return (p.X-q.X)*(p.X-q.X) + (p.Y-q.Y)*(p.Y-q.Y)
}

func getOrientation(p1, p2, p3 Point) orientation {
	// See: http://www.dcs.gla.ac.uk/~pat/52233/slides/Geometry1x1.pdf
	o := (p2.Y-p1.Y)*(p3.X-p2.X) - (p3.Y-p2.Y)*(p2.X-p1.X)

	if o == 0 {
		return collinear
	} else if o < 0 {
		return clockwise
	}
	return counterClockwise
}

func insideSegment(p, q, r Point) bool {
	// NOTE: Assumes r is collinear with the segment (p, q).
	return r.X >= min(p.X, q.X) && r.X <= max(p.X, q.X) &&
		r.Y >= min(p.Y, q.Y) && r.Y <= max(p.Y, q.Y)
}

func segmentsIntersect(p1, q1, p2, q2 Point) bool {
	// See: http://www.dcs.gla.ac.uk/~pat/52233/slides/Geometry1x1.pdf
	o1 := getOrientation(p1, q1, p2)
	o2 := getOrientation(p1, q1, q2)
	o3 := getOrientation(p2, q2, p1)
	o4 := getOrientation(p2, q2, q1)

	// General case.
	if o1 != o2 && o3 != o4 {
		return true
	}

	// Special cases for a segment and a point collinear with it.
	if o1 == collinear && insideSegment(p1, q1, p2) {
		return true
	}
	if o2 == collinear && insideSegment(p1, q1, q2) {
		return true
	}
	if o3 == collinear && insideSegment(p2, q2, p1) {
		return true
	}
	if o4 == collinear && insideSegment(p2, q2, q1) {
		return true
	}

	return false
}

func onTheSegment(p, q, r Point) bool {
	return getOrientation(p, q, r) == collinear && insideSegment(p, q, r)
}

func figEdgeCutsHoleEdge(fs, fe, hs, he Point) bool {
	if !segmentsIntersect(fs, fe, hs, he) {
		return false
	}
	// The problem-definition allows figure-edges to be on the hole-edges.
	if onTheSegment(hs, he, fs) || onTheSegment(hs, he, fe) {
		return false
	}
	return true
}

func minDistToSegment(p, q, r Point) int32 {
	d := sqDist(p, q)
	if d == 0 {
		return sqDist(p, r)
	}
	dot := (r.X-p.X)*(q.X-p.X) + (r.Y-p.Y)*(q.Y-p.Y)
	t := math.Max(0, math.Min(1, float64(dot)/float64(d)))
	pX := int32(float64(p.X)*(1-t) + t*float64(q.X))
	pY := int32(float64(p.Y)*(1-t) + t*float64(q.Y))
	return sqDist(r, Point{pX, pY})
}

func getBounds(pts []Point) (Point, Point) {
	minPt := Point{math.MaxInt32, math.MaxInt32}
	maxPt := Point{math.MinInt32, math.MinInt32}
	for _, p := range pts {
		minPt.X = min(minPt.X, p.X)
		minPt.Y = min(minPt.Y, p.Y)
		maxPt.X = max(maxPt.X, p.X)
		maxPt.Y = max(maxPt.Y, p.Y)
	}
	return minPt, maxPt
}

func markHoleBorderCells(prob *Problem) {
	pp := &prob.preProc
	xDisp := pp.holeLow.X
	yDisp := pp.holeLow.Y
	hV := prob.Hole.Vertices
	nV := len(hV)
	for i := 0; i < nV; i++ {
		x0, y0 := hV[i].X, hV[i].Y
		x1, y1 := hV[(i+1)%nV].X, hV[(i+1)%nV].Y

		log.Printf("Drawing from %s to %s", Point{x0, y0}, Point{x1, y1})
		if x0 == x1 {
			y00, y11 := min(y0, y1), max(y0, y1)
			idx := (y00-yDisp)*pp.holeWidth + (x0 - xDisp)
			for y := y00; y <= y11; y++ {
				pp.cellsWithHoles[idx] = true
				idx += pp.holeWidth
			}
		} else if y0 == y1 {
			x00, x11 := min(x0, x1), max(x0, x1)
			idx := (y0-yDisp)*pp.holeWidth + (x00 - xDisp)
			for x := x00; x <= x11; x++ {
				pp.cellsWithHoles[idx] = true
				idx++
			}
		} else {
			// Given the line-equation "y = a*x + c", solving for this line-
			// segment gives "a = (y1 - y0)/(x1 - x0)" and
			// "c = (x1*y0 - x0*y1)/(x1 - x0)". Divide-by-zeroes avoided above.
			a := float64(y1-y0) / float64(x1-x0)
			c := float64(x1*y0-x0*y1) / float64(x1-x0)
			// To maximize pixel-coverage, iterate alone the horizontal or the
			// vertical axis depending on whether the line looks wide or not.
			if math.Abs(float64(x1-x0)) > math.Abs(float64(y1-y0)) {
				x00, x11 := min(x0, x1), max(x0, x1)
				for x := x00; x <= x11; x++ {
					y := int32(math.Round(a*float64(x) + c))
					idx := (y-yDisp)*pp.holeWidth + (x - xDisp)
					pp.cellsWithHoles[idx] = true
				}
			} else {
				y00, y11 := min(y0, y1), max(y0, y1)
				for y := y00; y <= y11; y++ {
					x := int32(math.Round((float64(y) - c) / a))
					idx := (y-yDisp)*pp.holeWidth + (x - xDisp)
					pp.cellsWithHoles[idx] = true
				}
			}
		}
	}
}

func markHoleCells(prob *Problem) error {
	pp := &prob.preProc

	// A masochistic option is to manually find the pixels for the polygon
	// corresponding to the hole by simulating its scan-line conversion. A
	// brute-force method is to first find the borders of the polygon, then
	// find an interior-point, and finally flood-fill the polygon. Painful.
	//
	// markHoleBorderCells(prob)

	// Instead of the masochistic option above, just use SDL2_gfx to find the
	// pixels filled for the hole-polygon in a virtual surface.
	s, err := sdl.CreateRGBSurface(
		0, pp.holeWidth, pp.holeHeight, 32, 0, 0, 0, 0)
	if err != nil {
		return err
	}
	var r *sdl.Renderer
	if r, err = sdl.CreateSoftwareRenderer(s); err != nil {
		return err
	}

	if err := r.SetDrawColor(0, 0, 0, 255); err != nil {
		return err
	}
	if err := r.Clear(); err != nil {
		return err
	}
	x := make([]int16, len(prob.Hole.Vertices))
	y := make([]int16, len(prob.Hole.Vertices))
	for i, vv := range prob.Hole.Vertices {
		x[i] = int16(vv.X - pp.holeLow.X)
		y[i] = int16(vv.Y - pp.holeLow.Y)
	}
	if !gfx.FilledPolygonColor(r, x, y, sdl.Color{255, 255, 255, 255}) {
		return fmt.Errorf("FilledPolygonColor() failed")
	}

	pixels := make([]byte, pp.holeWidth*pp.holeHeight*4)
	err = r.ReadPixels(nil, sdl.PIXELFORMAT_ARGB8888,
		unsafe.Pointer(&pixels[0]), int(pp.holeWidth*4))
	if err != nil {
		return fmt.Errorf("ReadPixels() failed")
	}
	pp.cellsWithHoles = make([]bool, pp.holeWidth*pp.holeHeight)
	for i := 0; i < len(pixels); i += 4 {
		pp.cellsWithHoles[i/4] = pixels[i+1] == 255
	}

	if err := r.Destroy(); err != nil {
		return err
	}
	s.Free()
	return nil
}

func isHoleCell(pt Point, prob *Problem) bool {
	pp := &prob.preProc
	if pt.X < pp.holeLow.X || pt.X > pp.holeHigh.X {
		return false
	}
	if pt.Y < pp.holeLow.Y || pt.Y > pp.holeHigh.Y {
		return false
	}
	idx := (pt.Y-pp.holeLow.Y)*pp.holeWidth + (pt.X - pp.holeLow.X)
	return pp.cellsWithHoles[idx]
}

func preProcessProblem(prob *Problem) error {
	fv := prob.Figure.Vertices
	pp := &prob.preProc
	pp.holeLow, pp.holeHigh = getBounds(prob.Hole.Vertices)
	pp.holeWidth = pp.holeHigh.X - pp.holeLow.X + 1
	pp.holeHeight = pp.holeHigh.Y - pp.holeLow.Y + 1
	pp.figLow, pp.figHigh = getBounds(fv)
	pp.low.X = min(pp.holeLow.X, pp.figLow.X)
	pp.low.Y = min(pp.holeLow.Y, pp.figLow.Y)
	pp.high.X = max(pp.holeHigh.X, pp.figHigh.X)
	pp.high.Y = max(pp.holeHigh.Y, pp.figHigh.Y)

	if err := markHoleCells(prob); err != nil {
		return err
	}

	pp.minEdgeScale = (epsilonScale - prob.Epsilon) / epsilonScale
	pp.maxEdgeScale = (epsilonScale + prob.Epsilon) / epsilonScale

	pp.figVertEdges = make([][]figVertEdgeInfo, len(fv))
	for i := 0; i < len(fv); i++ {
		pp.figVertEdges[i] = make([]figVertEdgeInfo, 0, len(fv)-1)
	}
	for _, e := range prob.Figure.Edges {
		pi, qi := e.StartIdx, e.EndIdx
		dist := float64(sqDist(fv[pi], fv[qi]))
		minDist := int32(pp.minEdgeScale * dist)
		maxDist := int32(pp.maxEdgeScale * dist)

		pei := figVertEdgeInfo{qi, minDist, maxDist}
		pp.figVertEdges[pi] = append(pp.figVertEdges[pi], pei)

		qei := figVertEdgeInfo{pi, minDist, maxDist}
		pp.figVertEdges[qi] = append(pp.figVertEdges[qi], qei)
	}

	log.Printf("Overall bounds: min=%s max=%s", pp.low, pp.high)
	log.Printf("Hole bounds: min=%s max=%s", pp.holeLow, pp.holeHigh)
	log.Printf("Figure bounds: min=%s max=%s", pp.figLow, pp.figHigh)
	log.Printf("Stretchability: %f -> %f", pp.minEdgeScale, pp.maxEdgeScale)

	return nil
}

func isVertexAllowed(sol *Pose, idx int, prob *Problem) bool {
	p := sol.Vertices[idx]
	for _, q := range prob.preProc.figVertEdges[idx] {
		dist := sqDist(p, sol.Vertices[q.idx])
		if dist < q.minDist || dist > q.maxDist {
			return false
		}
	}
	return true
}

func getPoseViolations(sol *Pose, prob *Problem) *poseViolations {
	pV := poseViolations{}

	pV.vertsOutsideHole = make([]int, 0, len(sol.Vertices))
	for i, sv := range sol.Vertices {
		if isHoleCell(sv, prob) {
			continue
		}
		pV.vertsOutsideHole = append(pV.vertsOutsideHole, i)
	}

	hV := prob.Hole.Vertices
	nhV := len(hV)
FigEdgesLoop:
	for _, e := range prob.Figure.Edges {
		p1, q1 := sol.Vertices[e.StartIdx], sol.Vertices[e.EndIdx]
		for i, p2 := range hV {
			q2 := hV[(i+1)%nhV]
			if figEdgeCutsHoleEdge(p1, q1, p2, q2) {
				pV.numStrayingEdges++
				continue FigEdgesLoop
			}
		}
	}

	return &pV
}

func ReadProblem(pFile string) (*Problem, error) {
	b, err := ioutil.ReadFile(pFile)
	if err != nil {
		return nil, err
	}
	var f interface{}
	err = json.Unmarshal(b, &f)
	if err != nil {
		return nil, err
	}
	var prob Problem
	m := f.(map[string]interface{})
	for k, v := range m {
		switch k {
		case "hole":
			h := v.([]interface{})
			prob.Hole.Vertices = make([]Point, len(h))
			for i, vv := range h {
				vp := vv.([]interface{})
				prob.Hole.Vertices[i].X = int32(vp[0].(float64))
				prob.Hole.Vertices[i].Y = int32(vp[1].(float64))
			}
		case "figure":
			ff := v.(map[string]interface{})
			for fk, fv := range ff {
				switch fk {
				case "edges":
					fe := fv.([]interface{})
					prob.Figure.Edges = make([]Line, len(fe))
					for i, vv := range fe {
						vl := vv.([]interface{})
						prob.Figure.Edges[i].StartIdx = int(vl[0].(float64))
						prob.Figure.Edges[i].EndIdx = int(vl[1].(float64))
					}
				case "vertices":
					fvv := fv.([]interface{})
					prob.Figure.Vertices = make([]Point, len(fvv))
					for i, vp := range fvv {
						vpp := vp.([]interface{})
						prob.Figure.Vertices[i].X = int32(vpp[0].(float64))
						prob.Figure.Vertices[i].Y = int32(vpp[1].(float64))
					}
				default:
					return nil, fmt.Errorf(
						"unknown figure-level JSON-key %q", fk)
				}
			}
		case "epsilon":
			prob.Epsilon = v.(float64)
		case "bonuses":
			// TODO: Parse and use bonuses.
		default:
			return nil, fmt.Errorf("unknown top-level JSON-key %q", k)
		}
	}
	if err := preProcessProblem(&prob); err != nil {
		return nil, err
	}
	return &prob, nil
}

func ValidateSolution(sol *Pose, prob *Problem) error {
	if len(sol.Vertices) != len(prob.Figure.Vertices) {
		return fmt.Errorf("wrong number of vertices (%d)", len(sol.Vertices))
	}
	pv := getPoseViolations(sol, prob)
	if len(pv.vertsOutsideHole) > 0 {
		return fmt.Errorf("%d vertices outside", len(pv.vertsOutsideHole))
	}
	if pv.numStrayingEdges > 0 {
		return fmt.Errorf("%d stray edges", pv.numStrayingEdges)
	}
	for i, _ := range sol.Vertices {
		if !isVertexAllowed(sol, i, prob) {
			return fmt.Errorf("improperly stretched vertex @%d", i)
		}
	}
	return nil
}

func ReadSolution(sFile string, prob *Problem) (*Pose, error) {
	b, err := ioutil.ReadFile(sFile)
	if err != nil {
		return nil, err
	}
	var f interface{}
	err = json.Unmarshal(b, &f)
	if err != nil {
		return nil, err
	}
	var sol Pose
	m := f.(map[string]interface{})
	for k, v := range m {
		switch k {
		case "vertices":
			vv := v.([]interface{})
			sol.Vertices = make([]Point, len(vv))
			for i, vp := range vv {
				vpp := vp.([]interface{})
				sol.Vertices[i].X = int32(vpp[0].(float64))
				sol.Vertices[i].Y = int32(vpp[1].(float64))
			}
		default:
			return nil, fmt.Errorf("unknown pose-level JSON-key %q", k)
		}
	}
	if err := ValidateSolution(&sol, prob); err != nil {
		return nil, fmt.Errorf("invalid solution: %w", err)
	}
	return &sol, nil
}

func WriteSolution(sol *Pose, sFile string) error {
	v := make([][]int32, len(sol.Vertices))
	for i, p := range sol.Vertices {
		v[i] = make([]int32, 2, 2)
		v[i][0] = p.X
		v[i][1] = p.Y
	}
	m := make(map[string][][]int32)
	m["vertices"] = v

	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(sFile, b, 0644)
}

func GetDislikes(sol *Pose, prob *Problem) int32 {
	var dislikes int32 = 0
	for _, hv := range prob.Hole.Vertices {
		var minDist int32 = math.MaxInt32
		for _, fv := range sol.Vertices {
			minDist = min(minDist, sqDist(hv, fv))
		}
		dislikes += minDist
	}
	return dislikes
}
