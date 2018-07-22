package nmms

import (
	"fmt"
	"io/ioutil"
	"math"
)

const (
	numBitsPerByte = 8
)

type Coordinate struct {
	X, Y, Z int
}

// Matrix represents the Matrix that a Nanobot is supposed to fill.
type Matrix struct {
	res  int
	data []byte
}

func (c *Coordinate) String() string {
	return fmt.Sprintf("<%d,%d,%d>", c.X, c.Y, c.Z)
}

func (c *Coordinate) Add(oth *Coordinate) Coordinate {
	return Coordinate{c.X + oth.X, c.Y + oth.Y, c.Z + oth.Z}
}

// ReadFromFile populates the Matrix using the given Model file.
func (m *Matrix) ReadFromFile(path string) error {
	var err error
	m.data, err = ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	m.res = int(m.data[0])
	exp :=
		int(math.Ceil(float64(m.res*m.res*m.res)/float64(numBitsPerByte))) + 1
	if len(m.data) < exp {
		return fmt.Errorf(
			"Bad data in \"%s\" - %d expected vs %d actual bytes", path, exp,
			len(m.data))
	}
	return nil
}

// Resolution returns the current resolution of the Matrix.
func (m *Matrix) Resolution() int {
	return m.res
}

func (m *Matrix) Clear() {
	if len(m.data) < 2 {
		return
	}
	for i := 1; i < len(m.data); i++ {
		m.data[i] = 0
	}
}

func (m *Matrix) translateCoord(x, y, z int) (int, uint) {
	if x < 0 || x >= m.res || y < 0 || y >= m.res || z < 0 || z >= m.res {
		panic(fmt.Sprintf(
			"Coordinate <%d,%d,%d> is out of bounds for resolution %d", x, y,
			z, m.res))
	}
	bitIdx := int(x*m.res*m.res + y*m.res + z)
	return bitIdx/numBitsPerByte + 1, uint(bitIdx) % numBitsPerByte
}

// IsFull checks whether a given Cell in the Matrix is Full.
func (m *Matrix) IsFull(x, y, z int) bool {
	byteIdx, bitIdx := m.translateCoord(x, y, z)
	return m.data[byteIdx]&(1<<bitIdx) != 0
}

// SetFull sets a given Cell in the Matrix to be Full.
func (m *Matrix) SetFull(x, y, z int) {
	byteIdx, bitIdx := m.translateCoord(x, y, z)
	m.data[byteIdx] |= (1 << bitIdx)
}

// SetVoid sets a given Cell in the Matrix to be Void.
func (m *Matrix) SetVoid(x, y, z int) {
	byteIdx, bitIdx := m.translateCoord(x, y, z)
	m.data[byteIdx] &= (1 << bitIdx) ^ 0xFF
}
