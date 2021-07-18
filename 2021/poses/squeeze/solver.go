package squeeze

import (
	"log"
	"math"
	"math/rand"
	"time"
)

const (
	maxStepsToTgtSol = 16

	initTemp        = 1.0
	tempDecayFactor = 0.9
	minTemp         = 0.001
	itersPerTemp    = 1000
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

	done       bool
	currSol    *Pose
	currTemp   float64
	kBoltzmann float64

	solCosts []cost
}

func (t *tgtSolSolver) Reset() {
	t.currSol.Vertices = make([]Point, len(t.prob.Figure.Vertices))
	copy(t.currSol.Vertices, t.prob.Figure.Vertices)
	t.currStep = 0
}

func (t *tgtSolSolver) GetNextSolution() *Pose {
	if t.currStep >= maxStepsToTgtSol {
		return t.tgtSol
	}

	t.currStep++
	fracDone := float64(t.currStep) / float64(maxStepsToTgtSol)
	for i, v := range t.currSol.Vertices {
		deltaX := float64(t.tgtSol.Vertices[i].X-v.X) * fracDone
		deltaY := float64(t.tgtSol.Vertices[i].Y-v.Y) * fracDone
		t.currSol.Vertices[i].X += int32(deltaX)
		t.currSol.Vertices[i].Y += int32(deltaY)
	}
	return &t.currSol
}

func (t *tgtSolSolver) WasFinalSolution() bool {
	return t.currStep >= maxStepsToTgtSol
}

func (a *annealer) solCost(sol *Pose) cost {
	c := cost(GetDislikes(sol, a.prob))

	pV := getPoseViolations(sol, a.prob)
	nV := len(a.prob.Hole.Vertices)
	for _, i := range pV.vertsOutsideHole {
		minDist := int32(math.MaxInt32)
		for j, hV := range a.prob.Hole.Vertices {
			next := a.prob.Hole.Vertices[(j+1)%nV]
			minDist = min(minDist, minDistToSegment(hV, next, sol.Vertices[i]))
		}
		c += cost(minDist)
	}

	p := float64(pV.numStrayingEdges) / float64(len(a.prob.Figure.Edges))
	c += cost(float64(c) * p)

	return c
}

func (a *annealer) randomlyTugOnPoints(sol *Pose) {
	nV := len(sol.Vertices)

	// Fiddle with up to 10% of the vertices, but with at least 1 vertex.
	maxVictims := int(0.1 * float64(nV))
	const minVictims = 1
	if maxVictims < minVictims {
		maxVictims = minVictims
	}

	// Up to 1% jitter up/down and right/left, but by at least 1 cell.
	const maxJitterPct float64 = 0.01
	low, high := getBounds(sol.Vertices)
	xW, yH := high.X-low.X, high.Y-low.Y
	maxDispX := math.Max(1.0, maxJitterPct*float64(xW))
	maxDispY := math.Max(1.0, maxJitterPct*float64(yH))

	vQ := rand.Perm(nV)
	for i := 0; i < maxVictims; i++ {
		vIdx := vQ[i]
		ovX, ovY := sol.Vertices[vIdx].X, sol.Vertices[vIdx].Y

		nvX := int32(float64(ovX) + (2.0*rand.Float64()-1.0)*maxDispX)
		nvY := int32(float64(ovY) + (2.0*rand.Float64()-1.0)*maxDispY)
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
	a.randomlyTugOnPoints(&nSol)
	return &nSol
}

func (a *annealer) shouldSwitchSol(c0, c1 cost) bool {
	if c1 <= c0 {
		return true
	}
	// Note that c0 - c1 is negative here.
	return math.Exp(float64(c0-c1)/(a.kBoltzmann*a.currTemp)) > rand.Float64()
}

func (a *annealer) calibrate() {
	oCost := a.solCost(a.currSol)
	wCost := oCost
	const nTrials = 1000
	for i := 0; i < nTrials; i++ {
		nCost := a.solCost(a.getCandidateSol())
		if nCost > wCost {
			wCost = nCost
		}
	}
	dCost := wCost - oCost
	// This should yield a 10% probability at the highest temperature for
	// picking the solution with the worst cost.
	const worsePct = 0.1
	a.kBoltzmann = float64(-dCost) / (math.Log(worsePct) * a.currTemp)
	log.Printf("kB=%0.3f for %0.2f @T=%0.3f [dCost=%d].",
		a.kBoltzmann, worsePct, a.currTemp, dCost)
}

func (a *annealer) shouldGiveUp() bool {
	if a.currTemp < minTemp {
		return true
	}
	nSols := len(a.solCosts)
	const nChk = 3
	if nSols < nChk {
		return false
	}
	for i := nSols - 1; i > nSols-nChk; i-- {
		if a.solCosts[i] != a.solCosts[i-1] {
			return false
		}
	}
	return true
}

func (a *annealer) shouldRestart(cCost, bCost cost, i int) bool {
	if cCost <= bCost {
		return false
	}
	const waitPct = 0.75
	return i > int(waitPct*float64(itersPerTemp))
}

func (a *annealer) Reset() {
	s := Pose{}
	s.Vertices = make([]Point, len(a.prob.Figure.Vertices))
	copy(s.Vertices, a.prob.Figure.Vertices)

	a.done = false
	a.currSol = &s
	a.currTemp = initTemp

	expSols := math.Ceil(math.Log(minTemp/initTemp) / math.Log(tempDecayFactor))
	a.solCosts = make([]cost, 0, int(expSols))

	// Remove this to get a deterministic solution for a problem.
	rand.Seed(time.Now().UnixNano() % math.MaxInt32)

	a.calibrate()
}

func (a *annealer) GetNextSolution() *Pose {
	if a.done {
		return a.currSol
	}
	if a.shouldGiveUp() {
		a.done = true
		log.Printf("Costs: %v", a.solCosts)
		return a.currSol
	}

	currCost := a.solCost(a.currSol)
	bestSol, bestCost := a.currSol, currCost
	for i := 0; i < itersPerTemp; i++ {
		nSol := a.getCandidateSol()
		nCost := a.solCost(nSol)
		if a.shouldSwitchSol(currCost, nCost) {
			a.currSol = nSol
			currCost = nCost
		}
		if nCost < bestCost {
			bestSol, bestCost = nSol, nCost
		}
		if a.shouldRestart(currCost, bestCost, i) {
			a.currSol = bestSol
			currCost = bestCost
		}
	}
	a.solCosts = append(a.solCosts, currCost)
	a.currTemp *= tempDecayFactor

	return a.currSol
}

func (a *annealer) WasFinalSolution() bool {
	return a.done
}

func NewSolver(prob *Problem, tgtSol *Pose) Solver {
	if tgtSol == nil {
		return &annealer{prob: prob}
	}
	return &tgtSolSolver{prob: prob, tgtSol: tgtSol}
}
