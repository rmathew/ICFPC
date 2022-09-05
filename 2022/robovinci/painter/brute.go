package painter

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
	"time"
)

// <<<--- Brute force solver. --->>>

type bruteForceSolver struct{}

func (b *bruteForceSolver) String() string { return "BruteForce" }

func (b *bruteForceSolver) solve(prob *Problem) (*Program, error) {
	// TODO: Implement this.
	tBlocks, err := findTargetBlocks(prob)
	if err != nil {
		return nil, err
	}
	log.Printf("Target-blocks: %s", tBlocks)
	return nil, fmt.Errorf("%q UNIMPLEMENTED.", b)
}

// <<<--- Point-Cuts-based solver. --->>>

func clrCmp(rawVal float64) uint8 {
	rndVal := math.RoundToEven(rawVal)
	if rndVal < 0.0 || rndVal > 256.0 {
		panic(fmt.Errorf("crazy color-component (%.2f)", rndVal))
	}
	return uint8(rndVal)
}

func avgImgClr(img *image.NRGBA) color.NRGBA {
	iDim := img.Bounds()
	var i, avgR, avgG, avgB, avgA float64
	// We use Iterative Mean here to minimize the chances of overflows.
	for y := iDim.Min.Y; y < iDim.Max.Y; y++ {
		for x := iDim.Min.X; x < iDim.Max.X; x++ {
			nxtI := i + 1.0
			clr := img.NRGBAAt(x, y)
			avgR += (float64(clr.R) - avgR) / nxtI
			avgG += (float64(clr.G) - avgG) / nxtI
			avgB += (float64(clr.B) - avgB) / nxtI
			avgA += (float64(clr.A) - avgA) / nxtI
			i = nxtI
		}
	}
	return color.NRGBA{clrCmp(avgR), clrCmp(avgG), clrCmp(avgB), clrCmp(avgA)}
}

type pointCutsSolver struct {
	cSize int
	insns []move
}

type pcsBlockData struct {
	id       string
	blk      image.Rectangle
	img      *image.NRGBA
	clr      color.NRGBA
	simScore int
}

func (p *pointCutsSolver) String() string { return "PointCutter" }

func (p *pointCutsSolver) updBlkClrSimScore(bd *pcsBlockData) {
	bd.clr = avgImgClr(bd.img)

	// TODO: Fix the inefficiency here.
	tmpImg := image.NewNRGBA(bd.img.Bounds())
	clrImg := image.NewUniform(bd.clr)
	draw.Draw(tmpImg, tmpImg.Bounds(), clrImg, image.Point{}, draw.Src)
	bd.simScore = getSimilarityScore(bd.img, tmpImg)
}

func (p *pointCutsSolver) solveBlock(bd *pcsBlockData) {
	log.Printf("Block: %s@%s", bd.id, bd.blk)
	if bd.blk.Dx() < 2 {
		p.insns = append(p.insns, &colorBlock{bd.id, bd.clr})
		return
	}
	var bd1, bd2, bd3, bd4 pcsBlockData

	bd1.id = bd.id + ".0"
	bd2.id = bd.id + ".1"
	bd3.id = bd.id + ".2"
	bd4.id = bd.id + ".3"

	bMidX := (bd.blk.Min.X + bd.blk.Max.X) / 2
	bMidY := (bd.blk.Min.Y + bd.blk.Max.Y) / 2
	bd1.blk, bd2.blk, bd3.blk, bd4.blk = bd.blk, bd.blk, bd.blk, bd.blk
	bd1.blk.Max.X = bMidX
	bd1.blk.Max.Y = bMidY
	bd2.blk.Min.X = bMidX
	bd2.blk.Max.Y = bMidY
	bd3.blk.Min.X = bMidX
	bd3.blk.Min.Y = bMidY
	bd4.blk.Max.X = bMidX
	bd4.blk.Min.Y = bMidY

	iMin, iMax := bd.img.Bounds().Min, bd.img.Bounds().Max
	iMidX, iMidY := (iMin.X+iMax.X)/2, (iMin.Y+iMax.Y)/2
	bdImg1 := bd.img.SubImage(image.Rect(iMin.X, iMin.Y, iMidX, iMidY))
	bdImg2 := bd.img.SubImage(image.Rect(iMidX, iMin.Y, iMax.X, iMidY))
	bdImg3 := bd.img.SubImage(image.Rect(iMidX, iMidY, iMax.X, iMax.Y))
	bdImg4 := bd.img.SubImage(image.Rect(iMin.X, iMidY, iMidX, iMax.Y))
	bd1.img = bdImg1.(*image.NRGBA)
	bd2.img = bdImg2.(*image.NRGBA)
	bd3.img = bdImg3.(*image.NRGBA)
	bd4.img = bdImg4.(*image.NRGBA)

	p.updBlkClrSimScore(&bd1)
	p.updBlkClrSimScore(&bd2)
	p.updBlkClrSimScore(&bd3)
	p.updBlkClrSimScore(&bd4)

	oldCost, _ := moveCost(colorMove, rectSize(&bd.blk), p.cSize)
	oldCost += bd.simScore

	newCost, _ := moveCost(pointCutMove, rectSize(&bd.blk), p.cSize)
	newCost += bd1.simScore + bd2.simScore + bd3.simScore + bd4.simScore
	if newCost < oldCost {
		p.insns = append(p.insns, &pointCut{bd.id, image.Point{bMidX, bMidY}})
		p.solveBlock(&bd1)
		p.solveBlock(&bd2)
		p.solveBlock(&bd3)
		p.solveBlock(&bd4)
	} else {
		p.insns = append(p.insns, &colorBlock{bd.id, bd.clr})
	}
}

func (p *pointCutsSolver) solve(prob *Problem) (*Program, error) {
	tpb := prob.tgtPainting.Bounds()
	b0, ok := prob.initBlocks["0"]
	if !ok || !b0.shape.Eq(tpb) {
		log.Fatalf("%q cannot handle non-default initial-configurations.", p)
	}

	start := time.Now()

	p.cSize = rectSize(&tpb)
	p.insns = make([]move, 0, 1024)

	var bd pcsBlockData
	bd.id = "0"
	bd.blk = tpb
	bd.img = prob.tgtPainting
	p.updBlkClrSimScore(&bd)

	p.solveBlock(&bd)

	timeTakenMs := time.Now().Sub(start).Milliseconds()
	log.Printf("Got %d insns within %d ms.", len(p.insns), timeTakenMs)

	return &Program{p.insns}, nil
}
