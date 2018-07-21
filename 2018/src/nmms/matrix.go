package nmms

import (
	"fmt"
	"io/ioutil"
	"math"
)

const (
	numBitsPerByte = 8
)

// Matrix represents the Matrix that a Nanobot is supposed to fill.
type Matrix struct {
	res  int
	data []byte
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

func (m *Matrix) translateCoord(x, y, z int) (int, byte, error) {
	if x < 0 || x >= m.res || y < 0 || y >= m.res || z < 0 || z >= m.res {
		return 0, 0, fmt.Errorf(
			"Coordinate (%d, %d, %d) is out of bounds for resolution %d", x, y,
			z, m.res)
	}
	bitIdx := x*m.res*m.res + y*m.res + z
	return bitIdx/numBitsPerByte + 1, byte(bitIdx % numBitsPerByte), nil
}

// IsFull checks whether a given Cell in the Matrix is Full.
func (m *Matrix) IsFull(x, y, z int) (bool, error) {
	byteIdx, bitIdx, err := m.translateCoord(x, y, z)
	if err != nil {
		return false, err
	}
	return m.data[byteIdx]&(1<<bitIdx) != 0, nil
}

// SetFull sets a given Cell in the Matrix to be Full.
func (m *Matrix) SetFull(x, y, z int) error {
	byteIdx, bitIdx, err := m.translateCoord(x, y, z)
	if err != nil {
		return err
	}
	m.data[byteIdx] |= (1 << bitIdx)
	return nil
}
