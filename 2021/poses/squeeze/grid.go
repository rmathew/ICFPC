package squeeze

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

type Problem struct {
	Hole    Polygon
	Figure  Graph
	Epsilon int32
}

type Pose struct {
	Vertices []Point
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
		default:
			return nil, fmt.Errorf("unknown top-level JSON-key %q", k)
		}
	}
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
