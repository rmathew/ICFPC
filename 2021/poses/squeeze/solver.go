package squeeze

import (
	"math"
	"math/rand"
	"time"
)

const (
	maxInterpolationSteps = 16
)

type cost int32

type Solver interface {
	Reset()
	GetNextSolution() *Pose
	WasFinalSolution() bool
}

type tgtSolSolver struct {
	prob   *Problem
	tgtSol *Pose

	currSol  Pose
	currStep int
}

type annealer struct {
	prob *Problem

	foundSol   bool
	currSol    *Pose
	currTemp   float64
	kBoltzmann float64
}

func (t *tgtSolSolver) Reset() {
	t.currSol.Vertices = make([]Point, len(t.prob.Figure.Vertices))
	copy(t.currSol.Vertices, t.prob.Figure.Vertices)
	t.currStep = 0
}

func (t *tgtSolSolver) GetNextSolution() *Pose {
	if t.currStep >= maxInterpolationSteps {
		return t.tgtSol
	}

	t.currStep++
	fracDone := float64(t.currStep) / float64(maxInterpolationSteps)
	for i, v := range t.currSol.Vertices {
		deltaX := float64(t.tgtSol.Vertices[i].X-v.X) * fracDone
		deltaY := float64(t.tgtSol.Vertices[i].Y-v.Y) * fracDone
		t.currSol.Vertices[i].X += int32(deltaX)
		t.currSol.Vertices[i].Y += int32(deltaY)
	}
	return &t.currSol
}

func (t *tgtSolSolver) WasFinalSolution() bool {
	return t.currStep >= maxInterpolationSteps
}

func (a *annealer) solCost(sol *Pose) cost {
	c := cost(GetDislikes(sol, a.prob))
	if !IsValidSolution(sol, a.prob) {
		return c + cost(math.MaxInt32/2)
	}
	return c
}

func (a *annealer) randomlyTranslate(sol *Pose) {
	low, high := getBounds(sol.Vertices)
	xW, yH := high.X-low.X, high.Y-low.Y

	// Up to 1% jitter up/down and right/left, but by at least 1 cell.
	const maxJitterPct float64 = 0.01
	maxDispX := max(1, int32(maxJitterPct*float64(xW)))
	maxDispY := max(1, int32(maxJitterPct*float64(yH)))
	xBase, yBase := max(0, low.X-maxDispX), max(0, low.Y-maxDispY)
	xDisp := xBase + int32(2.0*rand.Float64()*float64(maxDispX)) - low.X
	yDisp := yBase + int32(2.0*rand.Float64()*float64(maxDispY)) - low.Y

	for i, _ := range sol.Vertices {
		sol.Vertices[i].X += xDisp
		sol.Vertices[i].Y += yDisp
	}
}

func (a *annealer) randomlyRotate(sol *Pose) {
	low, high := getBounds(sol.Vertices)
	cX, cY := (low.X+high.X)/2, (low.Y+high.Y)/2

	// TODO: Find a sure-fire way of rotating within bounds instead of doing
	// trial-and-error.
	const maxAttempts = 3
	const maxJitterPct float64 = 0.01
	for i := 0; i < maxAttempts; i++ {
		v := make([]Point, len(sol.Vertices))
		copy(v, sol.Vertices)

		// Rotate up to `maxJitterPct`-% of 2xPi radians in either direction.
		theta := 2.0 * maxJitterPct * 2.0 * math.Pi * (rand.Float64() - 0.5)
		cosT, sinT := math.Cos(theta), math.Sin(theta)
		for j, _ := range v {
			x, y := v[j].X, v[j].Y
			// Translate so that the center is at the origin.
			x -= cX
			y -= cY
			// Rotate around the new origin.
			nX := int32(float64(x)*cosT - float64(y)*sinT)
			nY := int32(float64(x)*sinT + float64(y)*cosT)
			// Translate back into the original frame of reference.
			v[j].X, v[j].Y = nX+cX, nY+cY
		}

		// Use the output of this attempt only if it is within bounds.
		nL, _ := getBounds(v)
		if nL.X < 0 || nL.Y < 0 {
			continue
		}
		copy(sol.Vertices, v)
		return
	}
}

func (a *annealer) randomlyTugOnPoints(sol *Pose) {
	low, high := getBounds(sol.Vertices)
	xW, yH := high.X-low.X, high.Y-low.Y

	// Up to 5% jitter up/down and right/left, but by at least 1 cell.
	const maxJitterPct float64 = 0.05
	maxDispX := max(1, int32(maxJitterPct*float64(xW)))
	maxDispY := max(1, int32(maxJitterPct*float64(yH)))
	const maxAttempts = 10
	for i := 0; i < maxAttempts; i++ {
		vIdx := rand.Intn(len(sol.Vertices))
		ovX, ovY := sol.Vertices[vIdx].X, sol.Vertices[vIdx].Y

		nvX := ovX + int32((2.0*rand.Float64()-1.0)*float64(maxDispX))
		nvY := ovY + int32((2.0*rand.Float64()-1.0)*float64(maxDispY))
		if nvX < 0 || nvY < 0 {
			continue
		}
		sol.Vertices[vIdx].X, sol.Vertices[vIdx].Y = nvX, nvY
		if !isVertexAllowed(sol, vIdx, a.prob) {
			sol.Vertices[vIdx].X, sol.Vertices[vIdx].Y = ovX, ovY
		}
	}
}

func (a *annealer) getCandidateSol() *Pose {
	var nSol Pose
	nSol.Vertices = make([]Point, len(a.currSol.Vertices))
	copy(nSol.Vertices, a.currSol.Vertices)

	switch rand.Intn(3) {
	case 0:
		a.randomlyTranslate(&nSol)
	case 1:
		a.randomlyRotate(&nSol)
	case 2:
		a.randomlyTugOnPoints(&nSol)
	}
	return &nSol
}

func (a *annealer) shouldSwitchSol(c0, c1 cost) bool {
	if c1 < c0 {
		return true
	} else if c1 == c0 {
		return false
	}
	// Note that c0 - c1 is negative here.
	return math.Exp(float64(c0-c1)/(a.kBoltzmann*a.currTemp)) > rand.Float64()
}

func (a *annealer) Reset() {
	s := Pose{}
	s.Vertices = make([]Point, len(a.prob.Figure.Vertices))
	copy(s.Vertices, a.prob.Figure.Vertices)

	a.foundSol = false
	a.currSol = &s
	a.currTemp = 1.0
	// TODO: Determine a suitable value for this normalization constant.
	// Ideally it should give a 50% probability at the highest temperature.
	a.kBoltzmann = 1.0

	// Remove this to get a deterministic solution for a problem.
	rand.Seed(time.Now().UnixNano() % math.MaxInt32)
}

func (a *annealer) GetNextSolution() *Pose {
	if a.foundSol {
		return a.currSol
	}

	initCost := a.solCost(a.currSol)
	currCost := initCost
	const itersPerTemp = 10000
	for i := 0; i < itersPerTemp; i++ {
		nSol := a.getCandidateSol()
		nCost := a.solCost(nSol)
		if a.shouldSwitchSol(currCost, nCost) {
			a.currSol = nSol
			currCost = nCost
		}
	}
	// TODO: Determine a suitable value for this decay factor.
	const tempDecayFactor = 0.98
	a.currTemp *= tempDecayFactor

	if currCost == initCost {
		a.foundSol = true
	}
	return a.currSol
}

func (a *annealer) WasFinalSolution() bool {
	return a.foundSol
}

func NewSolver(prob *Problem, tgtSol *Pose) Solver {
	if tgtSol == nil {
		return &annealer{prob: prob}
	}
	return &tgtSolSolver{prob: prob, tgtSol: tgtSol}
}
