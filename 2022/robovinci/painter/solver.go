package painter

import (
	"fmt"
	"image"
	"image/color"
	"log"
)

type solver interface {
	fmt.Stringer
	solve(prob *Problem) (*Program, error)
}

func SolveProblem(prob *Problem) (*Program, error) {
	b := bruteForceSolver{}
	log.Printf("Using the solver %q for the problem", b)
	return b.solve(prob)
}

type tgtBlock struct {
	id     int
	bounds image.Rectangle
	clr    color.NRGBA
}

func clr2Str(c color.NRGBA) string {
	return fmt.Sprintf("(%d,%d,%d,%d)", c.R, c.G, c.B, c.A)
}

func (t tgtBlock) String() string {
	return fmt.Sprintf("%d@[%s:%s]", t.id, t.bounds, clr2Str(t.clr))
}

func maybeGrowTargetBlock(blk *tgtBlock, img *image.NRGBA, pBlks []int) {
	iDim := img.Bounds()
	iMaxX, iMaxY := iDim.Max.X, iDim.Max.Y
	iW := iDim.Dx()
	bDim := blk.bounds
	bMinX, bMinY, bMaxX, bMaxY := bDim.Min.X, bDim.Min.Y, bDim.Max.X, bDim.Max.Y
	clr := img.NRGBAAt(bMinX, bMinY)

	canGrowRight := bMaxX < iMaxX
	if canGrowRight {
		x := bMaxX
		for y := bMinY; y < bMaxY; y++ {
			if pBlks[y*iW+x] != 0 || img.NRGBAAt(x, y) != clr {
				canGrowRight = false
				break
			}
		}
	}
	canGrowDown := bMaxY < iMaxY
	if canGrowDown {
		y := bMaxY
		for x := bMinX; x < bMaxX; x++ {
			if pBlks[y*iW+x] != 0 || img.NRGBAAt(x, y) != clr {
				canGrowDown = false
				break
			}
		}
	}
	canGrowDiag := canGrowRight && canGrowDown
	if canGrowDiag {
		x, y := bMaxX, bMaxY
		canGrowDiag = pBlks[y*iW+x] == 0 && img.NRGBAAt(x, y) == clr
	}

	x, y := bMaxX, bMaxY
	if canGrowDiag {
		bMaxX, bMaxY = bMaxX+1, bMaxY+1
		blk.bounds = image.Rect(bMinX, bMinY, bMaxX, bMaxY)
		for i := bMinX; i < bMaxX; i++ {
			pBlks[y*iW+i] = blk.id
		}
		for i := bMinY; i < bMaxY; i++ {
			pBlks[i*iW+x] = blk.id
		}
	} else if canGrowRight {
		bMaxX++
		blk.bounds = image.Rect(bMinX, bMinY, bMaxX, bMaxY)
		for i := bMinY; i < bMaxY; i++ {
			pBlks[i*iW+x] = blk.id
		}
	} else if canGrowDown {
		bMaxY++
		blk.bounds = image.Rect(bMinX, bMinY, bMaxX, bMaxY)
		for i := bMinX; i < bMaxX; i++ {
			pBlks[y*iW+i] = blk.id
		}
	}
	didGrow := canGrowRight || canGrowDown
	if didGrow && (x < iMaxX || y < iMaxY) {
		maybeGrowTargetBlock(blk, img, pBlks)
	}
}

func findTargetBlocks(prob *Problem) ([]tgtBlock, error) {
	img := prob.tgtPainting
	iDim := img.Bounds()
	iW, iH := iDim.Dx(), iDim.Dy()
	if iDim.Min.X != 0 || iDim.Min.Y != 0 {
		return nil, fmt.Errorf("non-origin-anchored target painting")
	}

	currBlkId := 0
	tBlocks := make([]tgtBlock, 0, 16)
	pixelBlocks := make([]int, iW*iH)
	for i := 0; i < len(pixelBlocks); i++ {
		blkId := pixelBlocks[i]
		if blkId != 0 {
			continue
		}

		currBlkId++
		pixelBlocks[i] = currBlkId
		x, y := i%iW, i/iW
		clr := img.NRGBAAt(x, y)
		blk := tgtBlock{currBlkId, image.Rect(x, y, x+1, y+1), clr}
		maybeGrowTargetBlock(&blk, img, pixelBlocks)
		tBlocks = append(tBlocks, blk)
	}
	return tBlocks, nil
}
