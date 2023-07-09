package plan

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math"
)

type Solution struct {
	MusicianPosns []Point
	Score         float64
}

type Solver interface {
	Reset()
	GetNextSolution() *Solution
	WasFinalSolution() bool
}

type annealer struct {
	prob *Problem

	done bool
}

func isBlocked(l *Point, mi int, mPos []Point) bool {
	// A musician is blocked by another one if the line-segment from the
	// listener to the former musician intersects the circle of radius 5.0
	// around the latter musician at two points.
	//
	// A line-segment has the equations:
	//   x = t*x0 + (1 - t)*x1
	//   y = t*y0 + (1 - t)*y1
	// for its coordinates, where 0 <= t <= 1.0. A circle has the equation:
	//   (x - xR)^2 + (y - yR)^2 = R^2
	// for its coordinates. Substituting the first set into the second and
	// simplifying a bit, we just need to solve the quadratic equation:
	//   ((x0 - x1)^2 + (y0 - y1)^2)*t^2 + \
	//     2*(x0*x1 - x0*xR - x1^2 + x1*xR)*t + \
	//     2*(y0*y1 - y0*yR - y1^2 + y1*yR)*t + \
	//     (x1 - xR)^2 + (y1 - yR)^2 - R^2
	// using the good old formula:
	//   (-b +/- sqrt(b^2 - 4*a*c)) / 2*a
	// to determine the X-axis values of the intersection-points, where:
	//   a = (x0 - x1)^2 + (y0 - y1)^2
	//   b = 2*(x0*x1 - x0*xR - x1^2 + x1*xR + y0*y1 - y0*yR - y1^2 + y1*yR)
	//   c = (x1 - xR)^2 + (y1 - yR)^2 - R^2
	// Since we're only interested in whether there is an intersection:
	//   1. No intersection (no real solutions), if b^2 - 4*a*c < 0.
	//   2. Touching (one solution), if b^2 - 4*a*c = 0.
	//   3. Intersection (two solutions), if b^2 - 4*a*c > 0.
	// we only need to determine #3.
	x0, y0 := l.X, l.Y
	x1, y1 := mPos[mi].X, mPos[mi].Y
	a := (x0-x1)*(x0-x1) + (y0-y1)*(y0-y1)
	for i, m := range mPos {
		if i == mi {
			continue
		}
		xR, yR := m.X, m.Y
		b := 2 * (x0*x1 - x0*xR - x1*x1 + x1*xR + y0*y1 - y0*yR - y1*y1 + y1*yR)
		c := (x1-xR)*(x1-xR) + (y1-yR)*(y1-yR) - 25.0
		if b*b > 4*a*c {
			return true
		}
	}
	return false
}

func (a *annealer) computeScore(mPos []Point) float64 {
	var score float64
	for _, l := range a.prob.Attendees {
		x, y := l.Position.X, l.Position.Y
		for i, m := range mPos {
			if isBlocked(&l.Position, i, mPos) {
				continue
			}
			dSq := (x-m.X)*(x-m.X) + (y-m.Y)*(y-m.Y)
			taste := l.Tastes[a.prob.Musicians[i]]
			score += math.Ceil(1000000.0 * taste / dSq)
		}
	}
	return score
}

func (a *annealer) Reset() {
	a.done = false
}

func (a *annealer) GetNextSolution() *Solution {
	// FIXME: This is a dummy solution from the example to check scoring.
	soln := &Solution{MusicianPosns: []Point{Point{590, 10}, Point{1100, 100}, Point{1100, 150}}}
	soln.Score = a.computeScore(soln.MusicianPosns)
	return soln
	// return nil // TODO: Implement.
}

func (a *annealer) WasFinalSolution() bool {
	// FIXME: return a.done
	return true
}

func NewSolver(prob *Problem) Solver {
	return &annealer{prob: prob}
}

func WriteSolution(soln *Solution, sFile string) error {
	if soln == nil {
		return errors.New("No solution to write out.")
	}
	if len(soln.MusicianPosns) == 0 {
		return errors.New("No musicians in the solution.")
	}
	if len(sFile) == 0 {
		return errors.New("Output file not specified.")
	}

	v := make([]map[string]float64, len(soln.MusicianPosns))
	for i, p := range soln.MusicianPosns {
		v[i] = make(map[string]float64)
		v[i]["x"] = p.X
		v[i]["y"] = p.Y
	}
	m := make(map[string][]map[string]float64)
	m["placements"] = v

	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(sFile, b, 0644)
}
