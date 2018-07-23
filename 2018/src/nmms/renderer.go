// TODO: Use OpenGL to make this much more efficient and flexible.
package nmms

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
	baseLineColor        sdl.Color
	fullCellColors       []sdl.Color
	botCellColors        []sdl.Color
}

type Renderer struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	params   drawParams
}

func checkGfx(succ bool) {
	if !succ {
		fmt.Printf("GFX_ERROR: Call failed.\n")
	}
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

	r.initDrawParams(n)

	r.renderer.SetDrawColor(0, 0, 0, 255)
	r.renderer.Clear()
	r.renderBase()
	r.renderer.Present()

	return nil
}

func (r *Renderer) initDrawParams(n *NmmSystem) {
	r.params.res = n.Mat.Resolution()

	const sinCos45 = 0.70710678118   // sin/cos of 45 deg, the projection-angle.
	const projScale = sinCos45 / 2.0 // Projected length of a unit length.
	const gutterSize = 32
	maxMatSize := iMin(winWidth, winHeight) - 2*gutterSize
	r.params.tileWidth = int(math.Floor(float64(maxMatSize) /
		(1.0 + projScale) / float64(r.params.res)))
	r.params.tileDelta = int(math.Floor(
		float64(maxMatSize-r.params.tileWidth*r.params.res) /
			float64(r.params.res)))

	matRenderSize := r.params.res * (r.params.tileWidth + r.params.tileDelta)
	r.params.iOff = (winWidth - matRenderSize) / 2
	r.params.jOff = (winHeight - matRenderSize) / 2
	// DEBUG
	fmt.Printf("maxMatSize=%d, matRenderSize=%d\n", maxMatSize, matRenderSize)
	fmt.Printf("iOff=%d, jOff=%d\n", r.params.iOff, r.params.jOff)

	r.params.baseColor = sdl.Color{64, 64, 64, 255}
	r.params.baseLineColor = sdl.Color{0, 0, 0, 255}

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

func (r *Renderer) renderBase() {
	var bi, bj = make([]int16, 4), make([]int16, 4)
	normShift := int16(r.params.res * r.params.tileWidth)
	projShift := int16(r.params.res * r.params.tileDelta)
	bi[0] = int16(r.params.iOff)
	bj[0] = int16(winHeight - r.params.jOff)
	bi[1] = bi[0] + projShift
	bj[1] = bj[0] - projShift
	bi[2] = bi[1] + normShift
	bj[2] = bj[1]
	bi[3] = bi[2] - projShift
	bj[3] = bj[2] + projShift
	checkGfx(gfx.FilledPolygonColor(r.renderer, bi, bj, r.params.baseColor))

	const minGridWidth = 5
	if r.params.tileWidth < minGridWidth {
		return
	}
	for x := 1; x < r.params.res; x++ {
		i0 := int32(r.params.iOff + x*r.params.tileWidth)
		j0 := int32(winHeight - r.params.jOff)
		checkGfx(gfx.LineColor(r.renderer, i0, j0, i0+int32(projShift),
			j0-int32(projShift), r.params.baseLineColor))
	}
	for z := 1; z < r.params.res; z++ {
		i0 := int32(r.params.iOff + z*r.params.tileDelta)
		j0 := int32(winHeight - r.params.jOff - z*r.params.tileDelta)
		checkGfx(gfx.LineColor(r.renderer, i0, j0, i0+int32(normShift),
			j0, r.params.baseLineColor))
	}
}

func (r *Renderer) renderVoxel(vi, vj, fi, fj []int16, colors []sdl.Color) {
	fi[0], fi[1], fi[2], fi[3] = vi[0], vi[1], vi[6], vi[5]
	fj[0], fj[1], fj[2], fj[3] = vj[0], vj[1], vj[6], vj[5]
	checkGfx(gfx.FilledPolygonColor(r.renderer, fi, fj, colors[frontFaceIdx]))

	fi[0], fi[1], fi[2], fi[3] = vi[1], vi[2], vi[3], vi[6]
	fj[0], fj[1], fj[2], fj[3] = vj[1], vj[2], vj[3], vj[6]
	checkGfx(gfx.FilledPolygonColor(r.renderer, fi, fj, colors[upFaceIdx]))

	fi[0], fi[1], fi[2], fi[3] = vi[5], vi[6], vi[3], vi[4]
	fj[0], fj[1], fj[2], fj[3] = vj[5], vj[6], vj[3], vj[4]
	checkGfx(gfx.FilledPolygonColor(r.renderer, fi, fj, colors[rightFaceIdx]))
}

func (r *Renderer) renderMatrix(m *Matrix) {
	var vi, vj = make([]int16, 7), make([]int16, 7)
	var fi, fj = make([]int16, 4), make([]int16, 4)
	// Without bounding-box: x:0->res-1, y:0->res-1, z:res-1->0
	bbMin, bbMax := m.BoundingBox()
	for x := bbMin.X; x <= bbMax.X; x++ {
		for y := bbMin.Y; y <= bbMax.Y; y++ {
			prevFull := false
			for z := bbMax.Z; z >= bbMin.Z; z-- {
				if !m.IsFull(x, y, z) {
					if prevFull {
						r.renderVoxel(vi, vj, fi, fj, r.params.fullCellColors)
					}
					prevFull = false
					continue
				}
				vi[0] = int16(r.params.iOff + x*r.params.tileWidth +
					z*r.params.tileDelta)
				vj[0] = int16(winHeight - (r.params.jOff +
					y*r.params.tileWidth + z*r.params.tileDelta))
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
	r.renderBase()
	r.renderMatrix(&n.Mat)
	r.renderBots(n.Bots)
	r.renderer.Present()
	return r.shouldContinue()
}
