package painter

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

// TODO: Support complex blocks and thus merge moves.
type block struct {
	shape      image.Rectangle
	pixelColor color.RGBA
}

type canvas struct {
	bounds   image.Rectangle
	size     float64
	block_id int
	blocks   map[string]block
}

type Problem struct {
	tgtPainting *image.RGBA
}

func rectSize(r *image.Rectangle) float64 {
	return float64(r.Dx()) * float64(r.Dy())
}

func newCanvas(p *Problem) *canvas {
	var c canvas
	c.bounds = p.tgtPainting.Bounds()
	c.size = rectSize(&c.bounds)
	c.block_id = 0
	c.blocks = make(map[string]block)
	c.blocks["0"] = block{p.tgtPainting.Bounds(), color.RGBA{255, 255, 255, 255}}
	return &c
}

//
// TODO: Solve the vertical flipping about the X-axis.
//

func ReadProblem(pFile string) (*Problem, error) {
	pF, err := os.Open(pFile)
	if err != nil {
		return nil, err
	}
	defer pF.Close()

	img, err := png.Decode(pF)
	if err != nil {
		return nil, err
	}
	iCpy := image.NewRGBA(img.Bounds().Sub(img.Bounds().Min))
	draw.Draw(iCpy, iCpy.Bounds(), img, img.Bounds().Min, draw.Src)

	var prob Problem
	prob.tgtPainting = iCpy
	return &prob, nil
}

func RenderResult(res *ExecResult, iFile string) error {
	f, err := os.Create(iFile)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := png.Encode(f, res.img); err != nil {
		return err
	}
	return nil
}
