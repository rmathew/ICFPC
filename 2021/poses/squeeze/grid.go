package squeeze

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
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

type preProcessedInfo struct {
	low, high         Point
	holeLow, holeHigh Point
	figLow, figHigh   Point
}

type Problem struct {
	Hole    Polygon
	Figure  Graph
	Epsilon int32

	preProc preProcessedInfo
}

type Pose struct {
	Vertices []Point
}

func (p Point) String() string {
	return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}

func getBounds(points []Point) (Point, Point) {
	minPt := Point{math.MaxInt32, math.MaxInt32}
	maxPt := Point{math.MinInt32, math.MinInt32}
	for _, p := range points {
		minPt.X = min(minPt.X, p.X)
		minPt.Y = min(minPt.Y, p.Y)
		maxPt.X = max(maxPt.X, p.X)
		maxPt.Y = max(maxPt.Y, p.Y)
	}
	return minPt, maxPt
}

func preProcessProblem(prob *Problem) {
	pp := &prob.preProc
	pp.holeLow, pp.holeHigh = getBounds(prob.Hole.Vertices)
	pp.figLow, pp.figHigh = getBounds(prob.Figure.Vertices)
	pp.low.X = min(pp.holeLow.X, pp.figLow.X)
	pp.low.Y = min(pp.holeLow.Y, pp.figLow.Y)
	pp.high.X = max(pp.holeHigh.X, pp.figHigh.X)
	pp.high.Y = max(pp.holeHigh.Y, pp.figHigh.Y)
	// DEBUG
	fmt.Printf("BOUNDS: min=%s max=%s\n", pp.low, pp.high)
	fmt.Printf("HOLE_BOUNDS: min=%s max=%s\n", pp.holeLow, pp.holeHigh)
	fmt.Printf("FIGURE_BOUNDS: min=%s max=%s\n", pp.figLow, pp.figHigh)
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
			prob.Epsilon = int32(v.(float64))
		case "bonuses":
			// TODO: Parse and use bonuses.
		default:
			return nil, fmt.Errorf("unknown top-level JSON-key %q", k)
		}
	}
	preProcessProblem(&prob)
	return &prob, nil
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
	if !IsValidPose(&sol, prob) {
		return nil, fmt.Errorf("invalid pose for problem")
	}
	return &sol, nil
}

func IsValidPose(sol *Pose, prob *Problem) bool {
	if prob == nil || sol == nil {
		return true
	}
	// TODO: Check epsilon-based constraints.
	return len(sol.Vertices) == len(prob.Figure.Vertices)
}
