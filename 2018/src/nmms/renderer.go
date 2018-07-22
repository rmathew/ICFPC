package nmms

import (
	"math"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	winWidth  = 1024
	winHeight = 768
	sinCos45  = 0.70710678118 // sin/cos of 45 degrees, the isometric angle.
)

const (
	frontFaceIdx int = iota
	upFaceIdx
	rightFaceIdx
)

type drawParams struct {
	res                  int
	tileWidth, tileDelta int
	iOff, jOff           int
	baseColor            sdl.Color
	fullCellColors       []sdl.Color
	botCellColors        []sdl.Color
}

type Renderer struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	params   drawParams
}

func (r *Renderer) Init(n *NmmSystem) error {
	var err error
	if err = sdl.Init(sdl.INIT_VIDEO | sdl.INIT_EVENTS); err != nil {
		return err
	}
	r.window, err = sdl.CreateWindow("Matrix Viewer (ICFPC 2018)",
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight,
		sdl.WINDOW_SHOWN)
	if err != nil {
		return err
	}
	r.renderer, err = sdl.CreateRenderer(r.window, -1, sdl.RENDERER_ACCELERATED)

	r.renderer.SetDrawColor(0, 0, 0, 255)
	r.renderer.Clear()
	r.renderer.Present()

	r.initDrawParams(n)

	return nil
}

func (r *Renderer) initDrawParams(n *NmmSystem) {
	r.params.res = n.Mat.Resolution()

	const gutterSize = 32.0
	maxMatSize := math.Min(float64(winWidth), float64(winHeight)) -
		2.0*gutterSize
	r.params.tileWidth = int(math.RoundToEven(maxMatSize / (1.0 + sinCos45) /
		float64(r.params.res)))
	r.params.tileDelta = int(math.RoundToEven(float64(r.params.tileWidth) *
		sinCos45))

	matRenderSize := r.params.res * (r.params.tileWidth + r.params.tileDelta)
	r.params.iOff = (winWidth - matRenderSize) / 2
	r.params.jOff = (winHeight - matRenderSize) / 2

	r.params.baseColor = sdl.Color{64, 64, 64, 255}

	r.params.fullCellColors = make([]sdl.Color, 3)
	r.params.fullCellColors[frontFaceIdx] = sdl.Color{255, 255, 255, 255}
	r.params.fullCellColors[upFaceIdx] = sdl.Color{225, 225, 225, 255}
	r.params.fullCellColors[rightFaceIdx] = sdl.Color{200, 200, 200, 255}

	r.params.botCellColors = make([]sdl.Color, 3)
	r.params.botCellColors[frontFaceIdx] = sdl.Color{255, 255, 0, 255}
	r.params.botCellColors[upFaceIdx] = sdl.Color{225, 225, 0, 255}
	r.params.botCellColors[rightFaceIdx] = sdl.Color{200, 200, 0, 255}
}

func (r *Renderer) Quit() {
	if r.window != nil {
		r.window.Destroy()
	}
	sdl.Quit()
}

func (r *Renderer) shouldContinue() bool {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.QuitEvent:
			return false
		case *sdl.KeyboardEvent:
			if t.Type == sdl.KEYDOWN && t.Keysym.Sym == sdl.K_ESCAPE {
				return false
			}
		}
	}
	return true
}

func (r *Renderer) renderVoxel(vi, vj, fi, fj []int16, colors []sdl.Color) {
	fi[0], fi[1], fi[2], fi[3] = vi[0], vi[1], vi[6], vi[5]
	fj[0], fj[1], fj[2], fj[3] = vj[0], vj[1], vj[6], vj[5]
	gfx.FilledPolygonColor(r.renderer, fi, fj, colors[frontFaceIdx])

	fi[0], fi[1], fi[2], fi[3] = vi[1], vi[2], vi[3], vi[6]
	fj[0], fj[1], fj[2], fj[3] = vj[1], vj[2], vj[3], vj[6]
	gfx.FilledPolygonColor(r.renderer, fi, fj, colors[upFaceIdx])

	fi[0], fi[1], fi[2], fi[3] = vi[5], vi[6], vi[3], vi[4]
	fj[0], fj[1], fj[2], fj[3] = vj[5], vj[6], vj[3], vj[4]

	gfx.FilledPolygonColor(r.renderer, fi, fj, colors[rightFaceIdx])
}

func (r *Renderer) renderMatrix(m *Matrix) {
	var bi, bj = make([]int16, 4), make([]int16, 4)
	bi[0], bj[0] = int16(r.params.iOff), int16(winHeight-r.params.jOff)
	bi[1], bj[1] = bi[0]+int16(r.params.res*r.params.tileDelta),
		bj[0]-int16(r.params.res*r.params.tileDelta)
	bi[2], bj[2] = bi[1]+int16(r.params.res*r.params.tileWidth), bj[1]
	bi[3], bj[3] = bi[2]-int16(r.params.res*r.params.tileDelta),
		bj[2]+int16(r.params.res*r.params.tileDelta)
	gfx.FilledPolygonColor(r.renderer, bi, bj, r.params.baseColor)

	var vi, vj = make([]int16, 7), make([]int16, 7)
	var fi, fj = make([]int16, 4), make([]int16, 4)
	for x := 0; x < r.params.res; x++ {
		for y := 0; y < r.params.res; y++ {
			prevFull := false
			for z := r.params.res - 1; z >= 0; z-- {
				if !m.IsFull(x, y, z) {
					if prevFull {
						r.renderVoxel(vi, vj, fi, fj, r.params.fullCellColors)
					}
					prevFull = false
					continue
				}
				vi[0] = int16(r.params.iOff + int(x)*r.params.tileWidth +
					int(z)*r.params.tileDelta)
				vj[0] = int16(winHeight - (r.params.jOff +
					int(y)*r.params.tileWidth + int(z)*r.params.tileDelta))
				vi[1] = vi[0]
				vj[1] = vj[0] - int16(r.params.tileWidth)
				if !prevFull {
					vi[2] = vi[1] + int16(r.params.tileDelta)
					vj[2] = vj[1] - int16(r.params.tileDelta)
					vi[3] = vi[2] + int16(r.params.tileWidth)
					vj[3] = vj[2]
					vi[4] = vi[3]
					vj[4] = vj[3] + int16(r.params.tileWidth)
				}
				vi[5] = vi[0] + int16(r.params.tileWidth)
				vj[5] = vj[0]
				vi[6] = vi[5]
				vj[6] = vj[5] - int16(r.params.tileWidth)

				prevFull = true
			}
			if prevFull {
				r.renderVoxel(vi, vj, fi, fj, r.params.fullCellColors)
			}
		}
	}
}

func (r *Renderer) renderBots(bots []Nanobot) {
	var vi, vj = make([]int16, 7), make([]int16, 7)
	var fi, fj = make([]int16, 4), make([]int16, 4)
	for _, b := range bots {
		vi[0] = int16(r.params.iOff + int(b.Pos.X)*r.params.tileWidth +
			int(b.Pos.Z)*r.params.tileDelta)
		vj[0] = int16(winHeight - (r.params.jOff +
			int(b.Pos.Y)*r.params.tileWidth + int(b.Pos.Z)*r.params.tileDelta))
		vi[1] = vi[0]
		vj[1] = vj[0] - int16(r.params.tileWidth)
		vi[2] = vi[1] + int16(r.params.tileDelta)
		vj[2] = vj[1] - int16(r.params.tileDelta)
		vi[3] = vi[2] + int16(r.params.tileWidth)
		vj[3] = vj[2]
		vi[4] = vi[3]
		vj[4] = vj[3] + int16(r.params.tileWidth)
		vi[5] = vi[0] + int16(r.params.tileWidth)
		vj[5] = vj[0]
		vi[6] = vi[5]
		vj[6] = vj[5] - int16(r.params.tileWidth)

		r.renderVoxel(vi, vj, fi, fj, r.params.botCellColors)
	}
}

func (r *Renderer) Update(n *NmmSystem) bool {
	r.renderer.SetDrawColor(0, 0, 0, 255)
	r.renderer.Clear()
	r.renderMatrix(&n.Mat)
	r.renderBots(n.Bots)
	r.renderer.Present()
	return r.shouldContinue()
}
