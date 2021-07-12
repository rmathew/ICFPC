package squeeze

const (
	maxInterpolationSteps int32 = 16
)

type Solver interface {
	InitSolver()
	GetNextSolution() *Pose
	WasFinalSolution() bool
}

type tgtSolSolver struct {
	prob   *Problem
	tgtSol *Pose

	currSol  Pose
	currStep int32
}

type annealer struct {
	prob *Problem

	sol      Pose
	currTemp float64
}

func (t *tgtSolSolver) InitSolver() {
	t.currSol.Vertices = make([]Point, len(t.prob.Figure.Vertices))
	copy(t.currSol.Vertices, t.prob.Figure.Vertices)
}

func (t *tgtSolSolver) GetNextSolution() *Pose {
	if t.currStep >= maxInterpolationSteps {
		return t.tgtSol
	}

	t.currStep++
	for i, v := range t.currSol.Vertices {
		deltaX := (t.tgtSol.Vertices[i].X - v.X) * t.currStep /
			maxInterpolationSteps
		deltaY := (t.tgtSol.Vertices[i].Y - v.Y) * t.currStep /
			maxInterpolationSteps
		t.currSol.Vertices[i].X += deltaX
		t.currSol.Vertices[i].Y += deltaY
	}
	return &t.currSol
}

func (t *tgtSolSolver) WasFinalSolution() bool {
	return t.currStep >= maxInterpolationSteps
}

func (a *annealer) InitSolver() {
	a.sol.Vertices = make([]Point, len(a.prob.Figure.Vertices))
	copy(a.sol.Vertices, a.prob.Figure.Vertices)

	a.currTemp = 1.0
}

func (a *annealer) GetNextSolution() *Pose {
	// TODO: Flesh this out.
	return &a.sol
}

func (a *annealer) WasFinalSolution() bool {
	// TODO: Flesh this out.
	return true
}

func NewSolver(prob *Problem, tgtSol *Pose) Solver {
	if tgtSol == nil {
		return &annealer{prob: prob}
	}
	return &tgtSolSolver{prob: prob, tgtSol: tgtSol}
}
