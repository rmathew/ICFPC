package painter

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
)

// <<<--- Moves on the canvas. --->>>

type moveType int

const (
	lineCutMove moveType = iota
	pointCutMove
	colorMove
	swapMove
	mergeMove
)

type move interface {
	fmt.Stringer
	execute(c *canvas) (int, error)
}

func moveCost(mvType moveType, blockSize, canvasSize int) (int, error) {
	var baseCost float64
	switch mvType {
	case lineCutMove:
		baseCost = 7.0
	case pointCutMove:
		baseCost = 10.0
	case colorMove:
		baseCost = 5.0
	case swapMove:
		baseCost = 3.0
	case mergeMove:
		baseCost = 1.0
	default:
		return 0, fmt.Errorf("unknown move-type")
	}
	cost := baseCost * (float64(canvasSize) / float64(blockSize))
	return int(math.Round(cost)), nil
}

// <<<--- LINE CUT --->>>

type cutDir int

const (
	horizontal cutDir = iota
	vertical
)

type lineCut struct {
	blockId string
	dir     cutDir
	offset  int
}

func (m *lineCut) String() string {
	dirStr := "x"
	if m.dir == horizontal {
		dirStr = "y"
	}
	return fmt.Sprintf("cut [%s] [%s] [%d]", m.blockId, dirStr, m.offset)
}

func (m *lineCut) execute(c *canvas) (int, error) {
	b0, ok := c.blocks[m.blockId]
	if !ok {
		return 0, fmt.Errorf("bad block-id %q", m.blockId)
	}
	b1, b2 := b0, b0
	if m.dir == vertical {
		if m.offset < b0.shape.Min.X || m.offset >= b0.shape.Max.X {
			return 0, fmt.Errorf("X-offset %d out of bounds", m.offset)
		}
		b1.shape.Max.X = m.offset
		b2.shape.Min.X = m.offset
	} else {
		if m.offset < b0.shape.Min.Y || m.offset >= b0.shape.Max.Y {
			return 0, fmt.Errorf("Y-offset %d out of bounds", m.offset)
		}
		b1.shape.Max.Y = m.offset
		b2.shape.Min.Y = m.offset
	}
	delete(c.blocks, m.blockId)
	c.blocks[m.blockId+".0"] = b1
	c.blocks[m.blockId+".1"] = b2
	return moveCost(lineCutMove, rectSize(&b0.shape), c.size)
}

// <<<--- POINT CUT --->>>

type pointCut struct {
	blockId string
	point   image.Point
}

func (m *pointCut) String() string {
	return fmt.Sprintf("cut [%s] [%d, %d]", m.blockId, m.point.X, m.point.Y)
}

func (m *pointCut) execute(c *canvas) (int, error) {
	b0, ok := c.blocks[m.blockId]
	if !ok {
		return 0, fmt.Errorf("bad block-id %q", m.blockId)
	}
	b1, b2, b3, b4 := b0, b0, b0, b0
	if m.point.X < b0.shape.Min.X || m.point.X >= b0.shape.Max.X {
		return 0, fmt.Errorf("X-point %d out of bounds", m.point.X)
	}
	if m.point.Y < b0.shape.Min.Y || m.point.Y >= b0.shape.Max.Y {
		return 0, fmt.Errorf("Y-point %d out of bounds", m.point.Y)
	}
	b1.shape.Max.X = m.point.X
	b1.shape.Max.Y = m.point.Y
	b2.shape.Min.X = m.point.X
	b2.shape.Max.Y = m.point.Y
	b3.shape.Min.X = m.point.X
	b3.shape.Min.Y = m.point.Y
	b4.shape.Max.X = m.point.X
	b4.shape.Min.Y = m.point.Y

	delete(c.blocks, m.blockId)
	c.blocks[m.blockId+".0"] = b1
	c.blocks[m.blockId+".1"] = b2
	c.blocks[m.blockId+".2"] = b3
	c.blocks[m.blockId+".3"] = b4
	return moveCost(pointCutMove, rectSize(&b0.shape), c.size)
}

// <<<--- COLOR BLOCK --->>>

type colorBlock struct {
	blockId    string
	pixelColor color.NRGBA
}

func (m *colorBlock) String() string {
	clr := m.pixelColor
	return fmt.Sprintf("color [%s] [%d, %d, %d, %d]", m.blockId, clr.R, clr.G, clr.B, clr.A)
}

func (m *colorBlock) execute(c *canvas) (int, error) {
	b0, ok := c.blocks[m.blockId]
	if !ok {
		return 0, fmt.Errorf("bad block-id %q", m.blockId)
	}
	b0.pixelColor = m.pixelColor
	c.blocks[m.blockId] = b0
	return moveCost(colorMove, rectSize(&b0.shape), c.size)
}

// <<<--- SWAP BLOCKS --->>>

type swapBlocks struct {
	blockId1 string
	blockId2 string
}

func (m *swapBlocks) String() string {
	return fmt.Sprintf("swap [%s] [%s]", m.blockId1, m.blockId2)
}

func (m *swapBlocks) execute(c *canvas) (int, error) {
	b1, ok := c.blocks[m.blockId1]
	if !ok {
		return 0, fmt.Errorf("bad first block-id %q", m.blockId1)
	}
	b2, ok := c.blocks[m.blockId2]
	if !ok {
		return 0, fmt.Errorf("bad second block-id %q", m.blockId2)
	}
	if b1.shape.Dx() != b2.shape.Dx() || b1.shape.Dy() != b2.shape.Dy() {
		return 0, fmt.Errorf("cannot swap blocks of different shape")
	}
	b1, b2 = b2, b1
	c.blocks[m.blockId1] = b1
	c.blocks[m.blockId2] = b2
	return moveCost(swapMove, rectSize(&b1.shape), c.size)
}

// <<<--- MERGE BLOCKS --->>>

type mergeBlocks struct {
	blockId1 string
	blockId2 string
}

func (m *mergeBlocks) String() string {
	return fmt.Sprintf("merge [%s] [%s]", m.blockId1, m.blockId2)
}

func compatibleRectsForMerge(r1, r2 *image.Rectangle) bool {
	// TODO: Implement this.
	return false
}

func (m *mergeBlocks) execute(c *canvas) (int, error) {
	b1, ok := c.blocks[m.blockId1]
	if !ok {
		return 0, fmt.Errorf("bad first block-id %q", m.blockId1)
	}
	b2, ok := c.blocks[m.blockId2]
	if !ok {
		return 0, fmt.Errorf("bad second block-id %q", m.blockId2)
	}
	if !compatibleRectsForMerge(&b1.shape, &b2.shape) {
		return 0, fmt.Errorf("blocks to be merged are not compatible")
	}
	// TODO: Implement correct cost-calculation.
	// Clarification from the organizers: "When two blocks are merged, the cost
	// is calculated by picking the larger block for computation."
	// TODO: Implement complex blocks.
	return 0, fmt.Errorf("unimplemented")
	// return moveCost(mergeMove, &b0, c)
}

// <<<--- Program-interpretation. --->>>

type Program struct {
	insns []move
}

type ExecResult struct {
	img   *image.NRGBA
	Score int
}

func getSimilarityScore(img1, img2 *image.NRGBA) int {
	iDim1, iDim2 := img1.Bounds(), img2.Bounds()
	if !iDim1.Eq(iDim2) {
		panic(fmt.Errorf("painting and canvas images of unequal bounds"))
	}

	diff := 0.0
	const alpha = 0.005
	for y := iDim1.Min.Y; y < iDim1.Max.Y; y++ {
		for x := iDim1.Min.X; x < iDim1.Max.X; x++ {
			clr1, clr2 := img1.NRGBAAt(x, y), img2.NRGBAAt(x, y)
			rDist := int(clr1.R) - int(clr2.R)
			gDist := int(clr1.G) - int(clr2.G)
			bDist := int(clr1.B) - int(clr2.B)
			aDist := int(clr1.A) - int(clr2.A)
			sumSqrDist := rDist*rDist + gDist*gDist + bDist*bDist + aDist*aDist
			diff += math.Sqrt(float64(sumSqrDist))
		}
	}
	return int(math.Round(diff * alpha))
}

func InterpretProgram(prob *Problem, prog *Program) (*ExecResult, error) {
	c := newCanvas(prob)
	var totalCost int
	for _, i := range prog.insns {
		insnCost, err := i.execute(c)
		if err != nil {
			return nil, fmt.Errorf("error executing instruction %w", err)
		}
		totalCost += insnCost
	}

	img := image.NewNRGBA(c.bounds)
	blk0, ok := c.blocks["0"]
	if !ok || !blk0.shape.Eq(img.Bounds()) {
		blankImg := image.NewUniform(initCanvasClr)
		draw.Draw(img, img.Bounds(), blankImg, image.Point{}, draw.Src)
	}
	for _, block := range c.blocks {
		srcImg := image.NewUniform(block.pixelColor)
		// TODO: Handle complex blocks as well.
		draw.Draw(img, block.shape, srcImg, image.Point{}, draw.Src)
	}

	var res ExecResult
	res.img = img
	res.Score = totalCost + getSimilarityScore(prob.tgtPainting, res.img)
	return &res, nil
}
