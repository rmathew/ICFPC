package galaxy

import (
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	winWidth  = 1024
	winHeight = 768
	zoomLevel = 4
)

type GalaxyViewer struct {
	window    *sdl.Window
	renderer  *sdl.Renderer
	colorPool []sdl.Color
}

func vec2scr(v *vect, x, y []int16) {
	x0 := int16(v.x*zoomLevel + winWidth/2)
	y0 := int16(v.y*zoomLevel + winHeight/2)
	x[0] = x0
	y[0] = y0
	x[1] = x0
	y[1] = y0 + zoomLevel
	x[2] = x0 + zoomLevel
	y[2] = y0 + zoomLevel
	x[3] = x0 + zoomLevel
	y[3] = y0
}

func scr2vec(x, y int32) *vect {
	return &vect{
		int64((x - winWidth/2) / zoomLevel),
		int64((y - winHeight/2) / zoomLevel),
	}
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
		{255, 0, 0, 255},
		{0, 255, 0, 255},
		{0, 0, 255, 255},
		{128, 128, 0, 255},
		{0, 128, 128, 255},
		{128, 0, 128, 255},
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

func (v *GalaxyViewer) drawVectors(p [][]*vect) {
	if p == nil {
		return
	}
	x := make([]int16, 4, 4)
	y := make([]int16, 4, 4)
	for i, img := range p {
		idx := i % len(v.colorPool)
		for _, pxs := range img {
			vec2scr(pxs, x, y)
			gfx.FilledPolygonColor(v.renderer, x, y, v.colorPool[idx])
		}
	}
}

func (v *GalaxyViewer) requestClick() (bool, *vect) {
	for {
		evt := sdl.WaitEvent()
		switch t := evt.(type) {
		case *sdl.QuitEvent:
			return false, nil
		case *sdl.KeyboardEvent:
			if t.Type == sdl.KEYDOWN && t.Keysym.Sym == sdl.K_ESCAPE {
				return false, nil
			}
		case *sdl.MouseButtonEvent:
			if t.Type == sdl.MOUSEBUTTONDOWN {
				return true, scr2vec(t.X, t.Y)
			}
		}
	}
}
