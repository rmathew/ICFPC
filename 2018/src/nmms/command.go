package nmms

import (
	"fmt"
)

type Command interface {
	Execute(n *NmmSystem, bIdx int) error
	decode(encCmds []byte) (int, error)
	fmt.Stringer
}

func DecodeNextCommand(encCmds []byte) (Command, int, error) {
	if len(encCmds) <= 0 {
		return nil, 0, fmt.Errorf("premature end of Command-stream")
	}
	var cmd Command
	firstByte := encCmds[0]
	switch {
	case firstByte == 0xFF: // i.e. is equal to 11111111
		cmd = new(HaltCmd)
	case firstByte == 0xFE: // i.e. is equal to 11111110
		cmd = new(WaitCmd)
	case firstByte == 0xFD: // i.e. is equal to 11111101
		cmd = new(FlipCmd)
	case firstByte&0xCF == 0x04: // i.e. has the form 00xx0100
		cmd = new(SMoveCmd)
	case firstByte&0x0F == 0x0C: // i.e. has the form xxyy1100
		cmd = new(LMoveCmd)
	case firstByte&0x07 == 0x07: // i.e. has the form xxxxx111
		cmd = new(FusionPCmd)
	case firstByte&0x07 == 0x06: // i.e. has the form xxxxx110
		cmd = new(FusionSCmd)
	case firstByte&0x07 == 0x05: // i.e. has the form xxxxx101
		cmd = new(FissionCmd)
	case firstByte&0x07 == 0x03: // i.e. has the form xxxxx011
		cmd = new(FillCmd)
	case firstByte&0x07 == 0x02: // i.e. has the form xxxxx010
		cmd = new(VoidCmd)
	case firstByte&0x07 == 0x01: // i.e. has the form xxxxx001
		cmd = new(GFillCmd)
	case firstByte&0x07 == 0x00: // i.e. has the form xxxxx000
		cmd = new(GVoidCmd)
	}
	if cmd != nil {
		offset, err := cmd.decode(encCmds)
		return cmd, offset, err
	}
	return nil, 0, fmt.Errorf("unknown Command-encoding %08b in stream",
		firstByte)
}

type HaltCmd bool

func (h *HaltCmd) Execute(n *NmmSystem, bIdx int) error {
	if len(n.Bots) != 1 || bIdx != 0 {
		return fmt.Errorf("halt requires a single Nanobot to be left")
	}
	n.Bots = []Nanobot{}
	return nil
}

func (h *HaltCmd) decode(encCmds []byte) (int, error) {
	return 1, nil
}

func (h *HaltCmd) String() string {
	return "Halt"
}

type WaitCmd bool

func (w *WaitCmd) Execute(n *NmmSystem, bIdx int) error {
	return nil
}

func (w *WaitCmd) decode(encCmds []byte) (int, error) {
	return 1, nil
}

func (w *WaitCmd) String() string {
	return "Wait"
}

type FlipCmd bool

func (f *FlipCmd) Execute(n *NmmSystem, bIdx int) error {
	if n.HighHarmonics {
		n.HighHarmonics = false
	} else {
		n.HighHarmonics = true
	}
	return nil
}

func (f *FlipCmd) decode(encCmds []byte) (int, error) {
	return 1, nil
}

func (f *FlipCmd) String() string {
	return "Flip"
}

type SMoveCmd struct {
	lld Coordinate
}

func (s *SMoveCmd) Execute(n *NmmSystem, bIdx int) error {
	n.Bots[bIdx].Pos = n.Bots[bIdx].Pos.Add(&s.lld)
	return nil
}

func toLld(axis byte, integer int) (Coordinate, error) {
	coord := Coordinate{}
	switch axis {
	case 0x1:
		coord.X = integer - 15
	case 0x2:
		coord.Y = integer - 15
	case 0x3:
		coord.Z = integer - 15
	default:
		return coord, fmt.Errorf("corrupt axis-encoding for SMove")
	}
	return coord, nil
}

func (s *SMoveCmd) decode(encCmds []byte) (int, error) {
	if len(encCmds) < 2 {
		return 0, fmt.Errorf("premature end of Command-stream for SMove")
	}
	axis := byte((encCmds[0] & 0x30) >> 4)
	integer := int(encCmds[1] & 0x1F)
	var err error
	s.lld, err = toLld(axis, integer)
	if err != nil {
		return 0, err
	}
	return 2, nil
}

func (s *SMoveCmd) String() string {
	return fmt.Sprintf("SMove %v", &s.lld)
}

func toSld(axis byte, integer int) (Coordinate, error) {
	coord := Coordinate{}
	switch axis {
	case 0x1:
		coord.X = integer - 5
	case 0x2:
		coord.Y = integer - 5
	case 0x3:
		coord.Z = integer - 5
	default:
		return coord, fmt.Errorf("corrupt axis-encoding for LMove")
	}
	return coord, nil
}

type LMoveCmd struct {
	sld1 Coordinate
	sld2 Coordinate
}

func (l *LMoveCmd) Execute(n *NmmSystem, bIdx int) error {
	// TODO: Implement this.
	return fmt.Errorf("unimplemented %v", l)
}

func (l *LMoveCmd) decode(encCmds []byte) (int, error) {
	if len(encCmds) < 2 {
		return 0, fmt.Errorf("premature end of Command-stream for LMove")
	}
	axis := byte((encCmds[0] & 0x30) >> 4)
	integer := int(encCmds[1] & 0x0F)
	var err error
	l.sld1, err = toSld(axis, integer)
	if err != nil {
		return 0, err
	}
	axis = byte((encCmds[0] & 0xD0) >> 6)
	integer = int((encCmds[1] & 0xF0) >> 4)
	l.sld2, err = toSld(axis, integer)
	if err != nil {
		return 0, err
	}
	return 2, nil
}

func (l *LMoveCmd) String() string {
	return fmt.Sprintf("LMove %v %v", &l.sld1, &l.sld2)
}

func isValidNd(c Coordinate) bool {
	numSet := 0
	if c.X == -1 || c.X == +1 {
		numSet++
	}
	if c.Y == -1 || c.Y == +1 {
		numSet++
	}
	if c.Z == -1 || c.Z == +1 {
		numSet++
	}
	return numSet > 0 && numSet < 3
}

func toNd(integer int) (Coordinate, error) {
	coord := Coordinate{}
	coord.X = integer/9 - 1
	integer %= 9
	coord.Y = integer/3 - 1
	integer %= 3
	coord.Z = integer - 1
	if !isValidNd(coord) {
		return coord, fmt.Errorf("malformed near coordinate difference %v",
			coord)
	}
	return coord, nil
}

type FusionPCmd struct {
	nd Coordinate
}

func (f *FusionPCmd) Execute(n *NmmSystem, bIdx int) error {
	// TODO: Implement this.
	return fmt.Errorf("unimplemented %v", f)
}

func (f *FusionPCmd) decode(encCmds []byte) (int, error) {
	integer := int((encCmds[0] & 0xF8) >> 3)
	var err error
	f.nd, err = toNd(integer)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func (f *FusionPCmd) String() string {
	return fmt.Sprintf("FusionP %v", &f.nd)
}

type FusionSCmd struct {
	nd Coordinate
}

func (f *FusionSCmd) Execute(n *NmmSystem, bIdx int) error {
	// TODO: Implement this.
	return fmt.Errorf("unimplemented %v", f)
}

func (f *FusionSCmd) decode(encCmds []byte) (int, error) {
	integer := int((encCmds[0] & 0xF8) >> 3)
	var err error
	f.nd, err = toNd(integer)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func (f *FusionSCmd) String() string {
	return fmt.Sprintf("FusionS %v", &f.nd)
}

type FissionCmd struct {
	nd Coordinate
	m  int
}

func (f *FissionCmd) Execute(n *NmmSystem, bIdx int) error {
	// TODO: Implement this.
	return fmt.Errorf("unimplemented %v", f)
}

func (f *FissionCmd) decode(encCmds []byte) (int, error) {
	if len(encCmds) < 2 {
		return 0, fmt.Errorf("premature end of Command-stream for Fission")
	}
	integer := int((encCmds[0] & 0xF8) >> 3)
	var err error
	f.nd, err = toNd(integer)
	if err != nil {
		return 0, err
	}
	f.m = int(encCmds[1])
	return 2, nil
}

func (f *FissionCmd) String() string {
	return fmt.Sprintf("Fission %v %d", &f.nd, f.m)
}

type FillCmd struct {
	nd Coordinate
}

func (f *FillCmd) Execute(n *NmmSystem, bIdx int) error {
	c := n.Bots[bIdx].Pos.Add(&f.nd)
	n.Mat.SetFull(c.X, c.Y, c.Z)
	return nil
}

func (f *FillCmd) decode(encCmds []byte) (int, error) {
	integer := int((encCmds[0] & 0xF8) >> 3)
	var err error
	f.nd, err = toNd(integer)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func (f *FillCmd) String() string {
	return fmt.Sprintf("Fill %v", &f.nd)
}

type VoidCmd struct {
	nd Coordinate
}

func (v *VoidCmd) Execute(n *NmmSystem, bIdx int) error {
	// TODO: Implement this.
	return fmt.Errorf("unimplemented %v", v)
}

func (v *VoidCmd) decode(encCmds []byte) (int, error) {
	integer := int((encCmds[0] & 0xF8) >> 3)
	var err error
	v.nd, err = toNd(integer)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func (v *VoidCmd) String() string {
	return fmt.Sprintf("Void %v", &v.nd)
}

type GFillCmd struct {
	nd Coordinate
	fd Coordinate
}

func (g *GFillCmd) Execute(n *NmmSystem, bIdx int) error {
	// TODO: Implement this.
	return fmt.Errorf("unimplemented %v", g)
}

func (g *GFillCmd) decode(encCmds []byte) (int, error) {
	if len(encCmds) < 4 {
		return 0, fmt.Errorf("premature end of Command-stream for GFill")
	}
	integer := int((encCmds[0] & 0xF8) >> 3)
	var err error
	g.nd, err = toNd(integer)
	if err != nil {
		return 0, err
	}
	g.fd.X = int(encCmds[1])
	g.fd.Y = int(encCmds[2])
	g.fd.Z = int(encCmds[3])
	return 4, nil
}

func (g *GFillCmd) String() string {
	return fmt.Sprintf("GFill %v %v", &g.nd, &g.fd)
}

type GVoidCmd struct {
	nd Coordinate
	fd Coordinate
}

func (g *GVoidCmd) Execute(n *NmmSystem, bIdx int) error {
	// TODO: Implement this.
	return fmt.Errorf("unimplemented %v", g)
}

func (g *GVoidCmd) decode(encCmds []byte) (int, error) {
	if len(encCmds) < 4 {
		return 0, fmt.Errorf("premature end of Command-stream for GVoid")
	}
	integer := int((encCmds[0] & 0xF8) >> 3)
	var err error
	g.nd, err = toNd(integer)
	if err != nil {
		return 0, err
	}
	g.fd.X = int(encCmds[1])
	g.fd.Y = int(encCmds[2])
	g.fd.Z = int(encCmds[3])
	return 4, nil
}

func (g *GVoidCmd) String() string {
	return fmt.Sprintf("GVoid %v %v", &g.nd, &g.fd)
}
