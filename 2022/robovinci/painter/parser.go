package painter

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	// RegExes copied from the Playground Parser.ts (with minor modifications).
	numRE   = `(0|[1-9]+[0-9]*)`
	byteRE  = `(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])`
	rgbaRE  = `\[(` + byteRE + `,` + byteRE + `,` + byteRE + `,` + byteRE + `)\]`
	blkId   = `\[(` + numRE + `(\.` + numRE + `)*)\]`
	pointRE = `\[(` + numRE + `,` + numRE + `)\]`
)

var colorInsnRE = regexp.MustCompile(`color` + blkId + rgbaRE)
var lineCutInsnRE = regexp.MustCompile(`cut` + blkId + `\[(x|X|y|Y)\]\[(` + numRE + `)\]`)
var pointCutInsnRE = regexp.MustCompile(`cut` + blkId + pointRE)
var mergeInsnRE = regexp.MustCompile(`merge` + blkId + blkId)
var swapInsnRE = regexp.MustCompile(`swap` + blkId + blkId)

func rgbaFromStr(sA []string) color.RGBA {
	rgba := make([]uint8, 0, len(sA))
	for _, s := range sA {
		n, err := strconv.Atoi(s)
		if err != nil {
			panic(err)
		}
		rgba = append(rgba, uint8(n))
	}
	return color.RGBA{rgba[0], rgba[1], rgba[2], rgba[3]}
}

func parseProgramLine(line string) (move, error) {
	// Strip away all the whitespace within the line.
	ln := strings.Join(strings.Fields(line), "")
	if len(ln) == 0 || strings.HasPrefix(ln, "#") {
		return nil, nil
	}
	var cGroups []string
	if cGroups = colorInsnRE.FindStringSubmatch(ln); cGroups != nil {
		blockId := cGroups[1]
		rgba := cGroups[len(cGroups)-5]
		return &colorBlock{blockId, rgbaFromStr(strings.Split(rgba, ","))}, nil
	} else if cGroups = lineCutInsnRE.FindStringSubmatch(ln); cGroups != nil {
		blockId := cGroups[1]
		offset, err := strconv.Atoi(cGroups[len(cGroups)-1])
		if err != nil {
			panic(err)
		}
		var dir cutDir
		dirChar := cGroups[len(cGroups)-3]
		if dirChar == "x" || dirChar == "X" {
			dir = vertical
		} else if dirChar == "y" || dirChar == "Y" {
			dir = horizontal
		} else {
			panic(fmt.Errorf("unexpected direction-character %q", dirChar))
		}
		return &lineCut{blockId, dir, offset}, nil
	} else if cGroups = pointCutInsnRE.FindStringSubmatch(ln); cGroups != nil {
		blockId := cGroups[1]
		x, err := strconv.Atoi(cGroups[len(cGroups)-2])
		if err != nil {
			panic(err)
		}
		y, err := strconv.Atoi(cGroups[len(cGroups)-1])
		if err != nil {
			panic(err)
		}
		return &pointCut{blockId, image.Point{x, y}}, nil
	} else if cGroups = mergeInsnRE.FindStringSubmatch(ln); cGroups != nil {
		blockId1 := cGroups[1]
		blockId2 := cGroups[5]
		return &mergeBlocks{blockId1, blockId2}, nil
	} else if cGroups = swapInsnRE.FindStringSubmatch(ln); cGroups != nil {
		blockId1 := cGroups[1]
		blockId2 := cGroups[5]
		return &swapBlocks{blockId1, blockId2}, nil
	} else {
		return nil, fmt.Errorf("unknown instruction %q", ln)
	}
	return nil, nil
}

func parseProgram(file string) ([]move, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	progInsns := make([]move, 0, 32)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		m, err := parseProgramLine(scanner.Text())
		if err != nil {
			return nil, err
		}
		if m != nil {
			progInsns = append(progInsns, m)
		}
	}
	return progInsns, nil
}

func ReadProgram(sFile string) (*Program, error) {
	insns, err := parseProgram(sFile)
	if err != nil {
		return nil, err
	}
	var prog Program
	prog.insns = insns
	return &prog, nil
}
