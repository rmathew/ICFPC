package nmms

import (
	"io/ioutil"
)

type Tracer struct {
	data []byte
}

// ReadFromFile populates the Tracer using the given Trace file.
func (t *Tracer) ReadFromFile(path string) error {
	var err error
	t.data, err = ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return nil
}

func (t *Tracer) TakeCommands(n int) ([]Command, error) {
	var cmds []Command
	if len(t.data) == 0 {
		return cmds, nil
	}
	cmds = make([]Command, n)
	for i := 0; i < n; i++ {
		c, offset, err := DecodeNextCommand(t.data)
		if err != nil {
			return nil, err
		}
		cmds[i] = c
		t.data = t.data[offset:]
	}
	return cmds, nil
}
