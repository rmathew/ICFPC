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
	shape image.Rectangle
	// NOTE: The specification is unclear on whether the pixel-values are
	// alpha pre-multiplied (RGBA) or not (NRGBA). Assume the latter for now.
	pixelColor color.NRGBA
}

type canvas struct {
	bounds   image.Rectangle
	size     int
	block_id int
	blocks   map[string]block
}

type Problem struct {
	tgtPainting *image.NRGBA
}

func rectSize(r *image.Rectangle) int {
	return r.Dx() * r.Dy()
}

func newCanvas(p *Problem) *canvas {
	var c canvas
	c.bounds = p.tgtPainting.Bounds()
	c.size = rectSize(&c.bounds)
	c.block_id = 0
	c.blocks = make(map[string]block)
	c.blocks["0"] = block{p.tgtPainting.Bounds(), color.NRGBA{255, 255, 255, 255}}
	return &c
}

// Since the normal coordinates are with the origin at the left-bottom and the
// screen coordinates are with the origin at the left top, this function
// translates an image back and forth between the two systems.
func flipImgVertically(img *image.NRGBA) {
	bounds := img.Bounds()
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		maxY := (bounds.Min.Y + bounds.Max.Y) / 2
		for y := bounds.Min.Y; y < maxY; y++ {
			oldClr := img.At(x, y)
			newY := bounds.Min.Y + bounds.Max.Y - y
			img.Set(x, y, img.At(x, newY))
			img.Set(x, newY, oldClr)
		}
	}
}

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
	iCpy := image.NewNRGBA(img.Bounds().Sub(img.Bounds().Min))
	draw.Draw(iCpy, iCpy.Bounds(), img, img.Bounds().Min, draw.Src)
	flipImgVertically(iCpy)

	var prob Problem
	prob.tgtPainting = iCpy
	return &prob, nil
}

func RenderResult(res *ExecResult, iFile string) error {
	iCpy := image.NewNRGBA(res.img.Bounds().Sub(res.img.Bounds().Min))
	draw.Draw(iCpy, iCpy.Bounds(), res.img, res.img.Bounds().Min, draw.Src)
	flipImgVertically(iCpy)

	f, err := os.Create(iFile)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := png.Encode(f, iCpy); err != nil {
		return err
	}
	return nil
}
