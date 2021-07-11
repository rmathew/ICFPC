package squeeze

import (
	"math"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	scrWidth  = 1024
	scrHeight = 768
	padding   = 8
)

type UserInput struct {
	Quit bool
}

type Viewer struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	zoom     float64
	input    UserInput

	prob *Problem
	sol  *Pose
}

func (v *Viewer) pt2scr(p Point) (int32, int32) {
	x := int32(padding + float64(p.X)*v.zoom)
	y := int32(padding + float64(p.Y)*v.zoom)
	return x, y
}

func (v *Viewer) maybeDrawGrid() {
	if v.zoom < 6 {
		return
	}
	gridColor := sdl.Color{210, 210, 210, 255}
	for x := int32(0); x <= v.prob.preProc.high.X; x++ {
		xx, y0 := v.pt2scr(Point{x, 0})
		_, y1 := v.pt2scr(Point{x, v.prob.preProc.high.Y})
		gfx.VlineColor(v.renderer, xx, y0, y1, gridColor)
	}
	for y := int32(0); y <= v.prob.preProc.high.Y; y++ {
		x0, yy := v.pt2scr(Point{0, y})
		x1, _ := v.pt2scr(Point{v.prob.preProc.high.X, y})
		gfx.HlineColor(v.renderer, x0, x1, yy, gridColor)
	}
}

func (v *Viewer) drawProblem() {
	if v.prob == nil {
		return
	}
	x := make([]int16, len(v.prob.Hole.Vertices))
	y := make([]int16, len(v.prob.Hole.Vertices))
	for i, vv := range v.prob.Hole.Vertices {
		xx, yy := v.pt2scr(vv)
		x[i] = int16(xx)
		y[i] = int16(yy)
	}
	gfx.FilledPolygonColor(v.renderer, x, y, sdl.Color{0, 0, 0, 255})
}

func (v *Viewer) drawSolution(newSol *Pose) {
	if newSol != nil {
		v.sol = newSol
	}
	var verts []Point
	if v.sol == nil {
		if v.prob == nil {
			return
		}
		verts = v.prob.Figure.Vertices
	} else {
		verts = v.sol.Vertices
	}
	const lineWidth int32 = 3
	lineColor := sdl.Color{255, 0, 0, 255}
	for _, e := range v.prob.Figure.Edges {
		x0, y0 := v.pt2scr(verts[e.StartIdx])
		x1, y1 := v.pt2scr(verts[e.EndIdx])
		gfx.ThickLineColor(v.renderer, x0, y0, x1, y1, lineWidth, lineColor)
	}
}

func (v *Viewer) updateZoom() error {
	maxX := v.prob.preProc.high.X
	maxY := v.prob.preProc.high.Y

	v.zoom = 16
	v.zoom = math.Min(v.zoom, float64(scrWidth-2*padding)/float64(maxX+1))
	v.zoom = math.Min(v.zoom, float64(scrHeight-2*padding)/float64(maxY+1))
	return nil
}

func (v *Viewer) UpdateView(sol *Pose) {
	v.renderer.SetDrawColor(230, 224, 195, 255)
	v.renderer.Clear()
	v.maybeDrawGrid()
	v.drawProblem()
	v.drawSolution(sol)
	v.renderer.Present()
}

func (v *Viewer) Init(prob *Problem) error {
	var err error
	if err = sdl.Init(sdl.INIT_VIDEO | sdl.INIT_EVENTS); err != nil {
		return err
	}
	v.window, err = sdl.CreateWindow("Hole In The Wall Viewer (ICFPC 2021)",
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, scrWidth, scrHeight,
		sdl.WINDOW_SHOWN)
	if err != nil {
		return err
	}
	v.renderer, err = sdl.CreateRenderer(v.window, -1, sdl.RENDERER_ACCELERATED)
	v.prob = prob
	if err = v.updateZoom(); err != nil {
		return err
	}
	v.UpdateView(nil)
	return nil
}

func (v *Viewer) Quit() {
	if v.window != nil {
		v.window.Destroy()
	}
	sdl.Quit()
}

func (v *Viewer) MaybeGetUserInput() (*UserInput, error) {
	v.input.Quit = false
	for evt := sdl.PollEvent(); evt != nil; evt = sdl.PollEvent() {
		switch t := evt.(type) {
		case *sdl.QuitEvent:
			v.input.Quit = true
		case *sdl.WindowEvent:
			if t.Event == sdl.WINDOWEVENT_RESIZED ||
				t.Event == sdl.WINDOWEVENT_EXPOSED {
				if err := v.updateZoom(); err != nil {
					return nil, err
				}
				v.UpdateView(nil)
			}
		case *sdl.KeyboardEvent:
			if t.Type == sdl.KEYDOWN && t.Keysym.Sym == sdl.K_ESCAPE {
				v.input.Quit = true
			}
		}
	}
	return &v.input, nil
}
