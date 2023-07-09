package plan

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
)

type Point struct {
	X, Y float64 // (0,0) is at the bottom-left.
}

func (p *Point) String() string {
	return fmt.Sprintf("(%0.2f, %0.2f)", p.X, p.Y)
}

type Rect struct {
	Origin        Point // Position of the bottom-left of the rectangle.
	Width, Height float64
}

func (r *Rect) String() string {
	return fmt.Sprintf("%0.2fx%0.2f@%v", r.Width, r.Height, &r.Origin)
}

type Listener struct {
	Position Point
	Tastes   []float64
}

type Problem struct {
	Room      Rect
	Stage     Rect
	Musicians []int
	Attendees []Listener
}

func ReadProblem(pFile string) (*Problem, error) {
	log.Printf("Reading problem specification %q...", pFile)
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
	prob.Room.Origin = Point{0.0, 0.0}
	m := f.(map[string]interface{})
	for k, v := range m {
		switch k {
		case "room_width":
			prob.Room.Width = v.(float64)
		case "room_height":
			prob.Room.Height = v.(float64)
		case "stage_width":
			prob.Stage.Width = v.(float64)
		case "stage_height":
			prob.Stage.Height = v.(float64)
		case "stage_bottom_left":
			ff := v.([]interface{})
			if len(ff) != 2 {
				return nil, errors.New("len(stage_bottom_left) != 2")
			}
			prob.Stage.Origin.X = ff[0].(float64)
			prob.Stage.Origin.Y = ff[1].(float64)
		case "musicians":
			ff := v.([]interface{})
			prob.Musicians = make([]int, len(ff))
			for ii, vv := range ff {
				prob.Musicians[ii] = int(vv.(float64))
			}
		case "attendees":
			ff := v.([]interface{})
			prob.Attendees = make([]Listener, len(ff))
			for ii, vv := range ff {
				mmm := vv.(map[string]interface{})
				for kkk, vvv := range mmm {
					switch kkk {
					case "x":
						prob.Attendees[ii].Position.X = vvv.(float64)
					case "y":
						prob.Attendees[ii].Position.Y = vvv.(float64)
					case "tastes":
						ffff := vvv.([]interface{})
						prob.Attendees[ii].Tastes = make([]float64, len(ffff))
						for iiiii, vvvvv := range ffff {
							prob.Attendees[ii].Tastes[iiiii] = vvvvv.(float64)
						}
					default:
						return nil, fmt.Errorf("unknown attendees key %q", kkk)
					}
				}
			}
		default:
			return nil, fmt.Errorf("unknown top-level key %q", k)
		}
	}
	return &prob, nil
}
