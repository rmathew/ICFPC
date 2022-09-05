package painter

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
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
	initBlocks  map[string]block
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
	for k, v := range p.initBlocks {
		c.blocks[k] = v
		blkIdParts := strings.Split(k, ".")
		if bid, err := strconv.Atoi(blkIdParts[0]); err == nil {
			if bid > c.block_id {
				c.block_id = bid
			}
		}
	}
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

func maybeReadInitCfg(prob *Problem, cFile string) error {
	if len(cFile) == 0 {
		return nil
	}
	log.Printf("Reading initial configuration %q...", cFile)
	b, err := ioutil.ReadFile(cFile)
	if err != nil {
		return err
	}
	var f interface{}
	err = json.Unmarshal(b, &f)
	if err != nil {
		return err
	}
	m := f.(map[string]interface{})
	for k, v := range m {
		switch k {
		case "width": // Ignore
		case "height": // Ignore
		case "blocks":
			ff := v.([]interface{})
			for _, vv := range ff {
				var blkId string
				var blk block
				fff := vv.(map[string]interface{})
				for kkk, vvv := range fff {
					switch kkk {
					case "blockId":
						blkId = vvv.(string)
					case "bottomLeft":
						bl := vvv.([]interface{})
						blk.shape.Min.X = int(bl[0].(float64))
						blk.shape.Min.Y = int(bl[1].(float64))
					case "topRight":
						tr := vvv.([]interface{})
						blk.shape.Max.X = int(tr[0].(float64))
						blk.shape.Max.Y = int(tr[1].(float64))
					case "color":
						fclr := vvv.([]interface{})
						var clr color.NRGBA
						clr.R = uint8(fclr[0].(float64))
						clr.G = uint8(fclr[1].(float64))
						clr.B = uint8(fclr[2].(float64))
						clr.A = uint8(fclr[3].(float64))
						blk.pixelColor = clr
					}
				}
				// TODO: Check the sanity of the input first.
				log.Printf("InitBlock: %s@[%s:%s]", blkId, blk.shape, clr2Str(blk.pixelColor))
				prob.initBlocks[blkId] = blk
			}
		}
	}
	return nil
}

func ReadProblem(pFile, cFile string) (*Problem, error) {
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
	prob.initBlocks = make(map[string]block)
	if err := maybeReadInitCfg(&prob, cFile); err != nil {
		return nil, err
	}
	if len(prob.initBlocks) == 0 {
		prob.initBlocks["0"] = block{iCpy.Bounds(), initCanvasClr}
	}
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
