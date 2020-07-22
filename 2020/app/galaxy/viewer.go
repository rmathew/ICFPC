package galaxy

import (
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	winWidth  = 1024
	winHeight = 768
)

type GalaxyViewer struct {
	window    *sdl.Window
	renderer  *sdl.Renderer
	colorPool []sdl.Color
	zoomLevel int32
}

func (v *GalaxyViewer) vec2scr(iv *vect, x, y []int16) {
	xx := int64(winWidth/2 + iv.x*int64(v.zoomLevel))
	yy := int64(winHeight/2 - iv.y*int64(v.zoomLevel))
	if xx < 0 || yy < 0 || xx > winWidth || yy > winHeight {
		log.Printf("ERROR: %v maps to invalid coords (%d, %d).", iv, xx, yy)
		return
	}
	x0 := int16(xx)
	y0 := int16(yy)
	x[0] = x0
	y[0] = y0
	x[1] = x0
	y[1] = y0 + int16(v.zoomLevel)
	x[2] = x0 + int16(v.zoomLevel)
	y[2] = y0 + int16(v.zoomLevel)
	x[3] = x0 + int16(v.zoomLevel)
	y[3] = y0
}

func (v *GalaxyViewer) scr2vec(x, y int32) *vect {
	// Note that Go rounds towards zero for integer division.
	zx := x - winWidth/2
	rx := zx / v.zoomLevel
	if zx < 0 && -zx%v.zoomLevel != 0 {
		rx -= 1
	}
	zy := winHeight/2 - y
	ry := zy / v.zoomLevel
	if zy > 0 && zy%v.zoomLevel != 0 {
		ry += 1
	}
	return &vect{int64(rx), int64(ry)}
}

func (v *GalaxyViewer) Init() error {
	var err error
	if err = sdl.Init(sdl.INIT_VIDEO | sdl.INIT_EVENTS); err != nil {
		return err
	}
	v.window, err = sdl.CreateWindow("Galaxy Viewer (ICFPC 2020)",
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight,
		sdl.WINDOW_SHOWN)
	if err != nil {
		return err
	}
	v.renderer, err = sdl.CreateRenderer(v.window, -1, sdl.RENDERER_ACCELERATED)

	v.initColorPool()
	v.update(nil)

	return nil
}

func (v *GalaxyViewer) Quit() {
	if v.window != nil {
		v.window.Destroy()
	}
	sdl.Quit()
}

func (v *GalaxyViewer) initColorPool() {
	v.colorPool = []sdl.Color{
		{255, 255, 255, 255},
		{223, 223, 223, 223},
		{191, 191, 191, 191},
		{159, 159, 159, 159},
		{127, 127, 127, 127},
		{95, 95, 95, 95},
		{63, 63, 63, 63},
		{127, 0, 0, 63},
		{0, 127, 0, 63},
		{0, 0, 127, 63},
		{127, 127, 0, 31},
		{0, 127, 127, 31},
		{127, 0, 127, 31},
	}
}

func (v *GalaxyViewer) update(p [][]*vect) {
	v.renderer.SetDrawColor(0, 0, 0, 255)
	v.renderer.Clear()
	v.drawVectors(p)
	v.renderer.Present()
}

func (v *GalaxyViewer) shouldBreak() bool {
	for evt := sdl.PollEvent(); evt != nil; evt = sdl.PollEvent() {
		switch t := evt.(type) {
		case *sdl.QuitEvent:
			return true
		case *sdl.KeyboardEvent:
			if t.Type == sdl.KEYDOWN && t.Keysym.Sym == sdl.K_ESCAPE {
				return true
			}
		}
	}
	return false
}

func min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func getZoomLevel(p [][]*vect) int32 {
	found := false
	var minX int64 = math.MaxInt64
	var minY int64 = math.MaxInt64
	var maxX int64 = math.MinInt64
	var maxY int64 = math.MinInt64
	for _, img := range p {
		for _, v := range img {
			if v.x < minX {
				minX = v.x
			}
			if v.y < minY {
				minY = v.y
			}
			if v.x > maxX {
				maxX = v.x
			}
			if v.y > maxY {
				maxY = v.y
			}
			found = true
		}
	}

	const minZoomLevel int32 = 1
	const maxZoomLevel int32 = 32
	if !found {
		return maxZoomLevel
	}

	const gutter = 16
	zl := maxZoomLevel
	if minX < 0 {
		zl = min(zl, (winWidth/2-gutter)/int32(-minX))
	}
	if maxX > 0 {
		zl = min(zl, (winWidth/2-gutter)/int32(maxX))
	}
	if minY < 0 {
		zl = min(zl, (winHeight/2-gutter)/int32(-minY))
	}
	if maxY > 0 {
		zl = min(zl, (winHeight/2-gutter)/int32(maxY))
	}
	if zl < minZoomLevel {
		zl = minZoomLevel
	}
	return zl
}

func (v *GalaxyViewer) drawVectors(p [][]*vect) {
	if p == nil {
		return
	}
	v.zoomLevel = getZoomLevel(p)
	x := make([]int16, 4, 4)
	y := make([]int16, 4, 4)
	for i, img := range p {
		idx := i % len(v.colorPool)
		for _, pxs := range img {
			v.vec2scr(pxs, x, y)
			gfx.FilledPolygonColor(v.renderer, x, y, v.colorPool[idx])
		}
	}
}

func (v *GalaxyViewer) requestClick() (bool, *vect) {
	const maxIters = 10000
	for i := 0; i < maxIters; i++ {
		for evt := sdl.PollEvent(); evt != nil; evt = sdl.PollEvent() {
			switch t := evt.(type) {
			case *sdl.QuitEvent:
				return false, nil
			case *sdl.KeyboardEvent:
				if t.Type == sdl.KEYDOWN && t.Keysym.Sym == sdl.K_ESCAPE {
					return false, nil
				}
			case *sdl.MouseButtonEvent:
				if t.Type == sdl.MOUSEBUTTONDOWN {
					return true, v.scr2vec(t.X, t.Y)
				}
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return true, v.scr2vec(rand.Int31n(winWidth), rand.Int31n(winHeight))
}
