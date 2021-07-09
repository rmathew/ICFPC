package squeeze

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Point struct {
	X, Y int
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
	Epsilon int
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
				prob.Hole.Vertices[i].X = int(vp[0].(float64))
				prob.Hole.Vertices[i].Y = int(vp[1].(float64))
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
						prob.Figure.Vertices[i].X = int(vpp[0].(float64))
						prob.Figure.Vertices[i].Y = int(vpp[1].(float64))
					}
				default:
					return nil, fmt.Errorf(
						"unknown figure-level JSON-key %q", fk)
				}
			}
		case "epsilon":
			prob.Epsilon = int(v.(float64))
		default:
			return nil, fmt.Errorf("unknown top-level JSON-key %q", k)
		}
	}
	return &prob, nil
}
