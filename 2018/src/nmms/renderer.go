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

type Renderer struct {
	window   *sdl.Window
	renderer *sdl.Renderer
}

func (r *Renderer) Init() error {
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
	return nil
}

func (r *Renderer) Quit() {
	if r.window != nil {
		r.window.Destroy()
	}
	sdl.Quit()
}

func (r *Renderer) waitForEvent() {
	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
				break
			}
		}
	}
}

func (r *Renderer) renderBox(vi, vj, fi, fj []int16) {
	frontFaceColor := sdl.Color{255, 255, 255, 255}
	fi[0], fi[1], fi[2], fi[3] = vi[0], vi[1], vi[6], vi[5]
	fj[0], fj[1], fj[2], fj[3] = vj[0], vj[1], vj[6], vj[5]
	gfx.FilledPolygonColor(r.renderer, fi, fj, frontFaceColor)

	upFaceColor := sdl.Color{208, 208, 208, 255}
	fi[0], fi[1], fi[2], fi[3] = vi[1], vi[2], vi[3], vi[6]
	fj[0], fj[1], fj[2], fj[3] = vj[1], vj[2], vj[3], vj[6]
	gfx.FilledPolygonColor(r.renderer, fi, fj, upFaceColor)

	sideFaceColor := sdl.Color{192, 192, 192, 255}
	fi[0], fi[1], fi[2], fi[3] = vi[5], vi[6], vi[3], vi[4]
	fj[0], fj[1], fj[2], fj[3] = vj[5], vj[6], vj[3], vj[4]

	gfx.FilledPolygonColor(r.renderer, fi, fj, sideFaceColor)
}

func (r *Renderer) renderMatrix(m *Matrix) {
	r.renderer.SetDrawColor(0, 0, 0, 255)
	r.renderer.Clear()
	const gutterSize = 32.0
	maxMatSize := math.Min(float64(winWidth), float64(winHeight)) -
		2.0*gutterSize
	res := m.Resolution()
	tileWidth := int(math.Floor(maxMatSize / (1.0 + sinCos45) / float64(res)))
	tileDelta := int(math.Floor(float64(tileWidth) * sinCos45))
	matRenderSize := res * (tileWidth + tileDelta)
	iOff := (winWidth - matRenderSize) / 2
	jOff := (winHeight - matRenderSize) / 2

	var bi, bj = make([]int16, 4), make([]int16, 4)
	bi[0], bj[0] = int16(iOff), int16(winHeight-jOff)
	bi[1], bj[1] = bi[0]+int16(res*tileDelta), bj[0]-int16(res*tileDelta)
	bi[2], bj[2] = bi[1]+int16(res*tileWidth), bj[1]
	bi[3], bj[3] = bi[2]-int16(res*tileDelta), bj[2]+int16(res*tileDelta)
	gfx.FilledPolygonColor(r.renderer, bi, bj, sdl.Color{64, 64, 64, 255})

	var vi, vj = make([]int16, 7), make([]int16, 7)
	var fi, fj = make([]int16, 4), make([]int16, 4)
	for x := 0; x < res; x++ {
		for y := 0; y < res; y++ {
			prevFull := false
			for z := res - 1; z >= 0; z-- {
				currFull, err := m.IsFull(x, y, z)
				Check(err)
				if !currFull {
					if prevFull {
						r.renderBox(vi, vj, fi, fj)
					}
					prevFull = false
					continue
				}
				vi[0] = int16(iOff + x*tileWidth + z*tileDelta)
				vj[0] = int16(winHeight - (jOff + y*tileWidth + z*tileDelta))
				vi[1] = vi[0]
				vj[1] = vj[0] - int16(tileWidth)
				if !prevFull {
					vi[2] = vi[1] + int16(tileDelta)
					vj[2] = vj[1] - int16(tileDelta)
					vi[3] = vi[2] + int16(tileWidth)
					vj[3] = vj[2]
					vi[4] = vi[3]
					vj[4] = vj[3] + int16(tileWidth)
				}
				vi[5] = vi[0] + int16(tileWidth)
				vj[5] = vj[0]
				vi[6] = vi[5]
				vj[6] = vj[5] - int16(tileWidth)

				prevFull = true
			}
			if prevFull {
				r.renderBox(vi, vj, fi, fj)
			}
		}
	}
	r.renderer.Present()
}

func (r *Renderer) Update(m *Matrix) {
	r.renderMatrix(m)
	r.waitForEvent()
}
