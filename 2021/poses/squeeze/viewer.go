package squeeze

import (
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	winWidth  = 1024
	winHeight = 768
)

type UserInput struct {
	Quit bool
}

type Viewer struct {
	window    *sdl.Window
	renderer  *sdl.Renderer
	zoomLevel int32
	input     UserInput
}

func (v *Viewer) drawProblem(prob *Problem) {
	x := make([]int16, len(prob.Hole.Vertices))
	y := make([]int16, len(prob.Hole.Vertices))
	for i, v := range prob.Hole.Vertices {
		x[i] = int16(v.X)
		y[i] = int16(v.Y)
	}
	gfx.FilledPolygonColor(v.renderer, x, y, sdl.Color{0, 0, 0, 255})
}

func (v *Viewer) drawSolution(prob *Problem, sol *Pose) {
	for _, e := range prob.Figure.Edges {
		x0 := int32(sol.Vertices[e.StartIdx].X)
		y0 := int32(sol.Vertices[e.StartIdx].Y)
		x1 := int32(sol.Vertices[e.EndIdx].X)
		y1 := int32(sol.Vertices[e.EndIdx].Y)
		gfx.LineColor(v.renderer, x0, y0, x1, y1, sdl.Color{255, 0, 0, 255})
	}
}

func (v *Viewer) UpdateView(prob *Problem, sol *Pose) {
	v.renderer.SetDrawColor(230, 224, 195, 255)
	v.renderer.Clear()
	if prob != nil {
		v.drawProblem(prob)
		if sol != nil {
			v.drawSolution(prob, sol)
		}
	}
	v.renderer.Present()
}

func (v *Viewer) Init() error {
	var err error
	if err = sdl.Init(sdl.INIT_VIDEO | sdl.INIT_EVENTS); err != nil {
		return err
	}
	v.window, err = sdl.CreateWindow("Grid Viewer (ICFPC 2021)",
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight,
		sdl.WINDOW_SHOWN)
	if err != nil {
		return err
	}
	v.renderer, err = sdl.CreateRenderer(v.window, -1, sdl.RENDERER_ACCELERATED)
	v.UpdateView(nil, nil)
	return nil
}

func (v *Viewer) Quit() {
	if v.window != nil {
		v.window.Destroy()
	}
	sdl.Quit()
}

func (v *Viewer) MaybeGetUserInput() *UserInput {
	v.input.Quit = false
	for evt := sdl.PollEvent(); evt != nil; evt = sdl.PollEvent() {
		switch t := evt.(type) {
		case *sdl.QuitEvent:
			v.input.Quit = true
		case *sdl.WindowEvent:
			if t.Event == sdl.WINDOWEVENT_RESIZED ||
				t.Event == sdl.WINDOWEVENT_EXPOSED {
				v.UpdateView(nil, nil)
			}
		case *sdl.KeyboardEvent:
			if t.Type == sdl.KEYDOWN && t.Keysym.Sym == sdl.K_ESCAPE {
				v.input.Quit = true
			}
		}
	}
	return &v.input
}
