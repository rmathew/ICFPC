package painter

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
)

// IMPORTANT: *All* the internal data-structures use screen-coordinates with its
// origin at the left-top. Since the visualizations are with the usual origin at
// the left-bottom, use `flipImgVertically()` to translate input/output between
// these coordinate_systems.

var initCanvasClr = color.NRGBA{255, 255, 255, 255}

// TODO: Support complex blocks and thus merge moves.
type block struct {
	shape image.Rectangle
	// NOTE: The organizers have clarified that they do not use alpha pre-
	// multiplied RGBA values (RGBA in Go).
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
	c.blocks["0"] = block{p.tgtPainting.Bounds(), initCanvasClr}
	return &c
}

func flipImgVertically(img *image.NRGBA) {
	iDim := img.Bounds()
	iMinX, iMinY, iMaxX, iMaxY := iDim.Min.X, iDim.Min.Y, iDim.Max.X, iDim.Max.Y
	midY := (iMinY + iMaxY) / 2
	for x := iMinX; x < iMaxX; x++ {
		for y := iMinY; y < midY; y++ {
			oldClr := img.NRGBAAt(x, y)
			newY := iMinY + iMaxY - y - 1
			img.SetNRGBA(x, y, img.NRGBAAt(x, newY))
			img.SetNRGBA(x, newY, oldClr)
		}
	}
}

func ReadProblem(pFile string) (*Problem, error) {
	log.Printf("Reading target painting %q...", pFile)
	pF, err := os.Open(pFile)
	if err != nil {
		return nil, err
	}
	defer pF.Close()

	img, err := png.Decode(pF)
	if err != nil {
		return nil, err
	}
	iDim := img.Bounds()
	log.Printf("Size of the target painting: %dx%d", iDim.Dx(), iDim.Dy())
	iCpy := image.NewNRGBA(iDim.Sub(iDim.Min))
	draw.Draw(iCpy, iCpy.Bounds(), img, iDim.Min, draw.Src)
	flipImgVertically(iCpy)

	var prob Problem
	prob.tgtPainting = iCpy
	return &prob, nil
}

func SaveProgram(prog *Program, res *ExecResult, pFile string) error {
	f, err := os.Create(pFile)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintf(f, "# Expected Score: %d\n", res.Score)
	for _, i := range prog.insns {
		fmt.Fprintf(f, "%s\n", i)
	}
	return nil
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
