package squeeze

import (
	"fmt"
	"math"

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

	prob *Problem
	sol  *Pose
}

func (v *Viewer) drawProblem() {
	if v.prob == nil {
		return
	}
	x := make([]int16, len(v.prob.Hole.Vertices))
	y := make([]int16, len(v.prob.Hole.Vertices))
	for i, vv := range v.prob.Hole.Vertices {
		x[i] = int16(vv.X * v.zoomLevel)
		y[i] = int16(vv.Y * v.zoomLevel)
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
	for _, e := range v.prob.Figure.Edges {
		x0 := int32(verts[e.StartIdx].X * v.zoomLevel)
		y0 := int32(verts[e.StartIdx].Y * v.zoomLevel)
		x1 := int32(verts[e.EndIdx].X * v.zoomLevel)
		y1 := int32(verts[e.EndIdx].Y * v.zoomLevel)
		gfx.ThickLineColor(v.renderer, x0, y0, x1, y1, lineWidth,
			sdl.Color{255, 0, 0, 255})
	}
}

func min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func (v *Viewer) updateZoomLevel() error {
	var maxX int32 = math.MinInt32
	var maxY int32 = math.MinInt32
	for _, vv := range v.prob.Hole.Vertices {
		maxX = max(maxX, vv.X)
		maxY = max(maxY, vv.Y)
	}
	for _, fv := range v.prob.Figure.Vertices {
		maxX = max(maxX, fv.X)
		maxY = max(maxY, fv.Y)
	}
	if maxX > winWidth {
		return fmt.Errorf(
			"max X (%d) more than screen-width (%d)", maxX, winWidth)
	}
	if maxY > winHeight {
		return fmt.Errorf(
			"max Y (%d) more than screen-height (%d)", maxY, winHeight)
	}

	const minZoomLevel int32 = 1
	const maxZoomLevel int32 = 16
	const gutter int32 = 8
	zl := maxZoomLevel
	zl = min(zl, (winWidth-2*gutter)/(maxX+1))
	zl = min(zl, (winHeight-2*gutter)/(maxY+1))
	v.zoomLevel = max(zl, minZoomLevel)
	return nil
}

func (v *Viewer) UpdateView(sol *Pose) {
	v.renderer.SetDrawColor(230, 224, 195, 255)
	v.renderer.Clear()
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
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight,
		sdl.WINDOW_SHOWN)
	if err != nil {
		return err
	}
	v.renderer, err = sdl.CreateRenderer(v.window, -1, sdl.RENDERER_ACCELERATED)
	v.prob = prob
	if err = v.updateZoomLevel(); err != nil {
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
				if err := v.updateZoomLevel(); err != nil {
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
