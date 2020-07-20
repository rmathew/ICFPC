package galaxy

import (
	"fmt"
	"math/bits"
	"strconv"
	"strings"
)

func modulate(n int64) string {
	var b strings.Builder
	ne := n
	if n >= 0 {
		b.WriteString("01")
	} else {
		ne = -n
		b.WriteString("10")
	}

	nBits := bits.Len64(uint64(ne))
	numNybbles := nBits / 4
	if nBits%4 > 0 {
		numNybbles++
	}
	for i := 0; i < numNybbles; i++ {
		b.WriteByte('1')
	}
	b.WriteByte('0')

	fullBits := fmt.Sprintf("%064b", ne)
	b.WriteString(fullBits[4*(16-numNybbles):])
	return b.String()
}

func demodulate(r []rune) (int64, error) {
	if len(r) < 3 {
		return 0, fmt.Errorf("too few runes (%d) to demodulate", len(r))
	}
	var mult int64 = +1
	if r[0] == '1' && r[1] == '0' {
		mult = -1
	}
	numNybbles := 0
	for i := 2; r[i] == '1' && i < len(r); i++ {
		numNybbles++
	}

	mSize := 3 + 4*numNybbles
	if len(r) < mSize {
		return 0, fmt.Errorf("too few runes (%d) - need %d", len(r), mSize)
	}

	var n int64 = 0
	if numNybbles > 0 {
		var err error
		n, err = strconv.ParseInt(string(r[3+numNybbles:]), 2, 64)
		if err != nil {
			return 0, err
		}
	}
	return n * mult, nil
}

func modulateList(e expr) (string, error) {
	if e == nil {
		return "", fmt.Errorf("NULL expr")
	}

	var b strings.Builder
	b.WriteString("11")
	if isNil(e) {
		return b.String(), nil
	}

	var e1, e2 expr
	var ok bool
	if ok, e1, e2 = isPair(e); !ok {
		return "", fmt.Errorf("not nil, or pair: %v", e)
	}

	var n int64
	if isNil(e1) {
		b.WriteString("00")
	} else if ok, n = isNumber(e1); ok {
		b.WriteString(modulate(n))
	} else {
		return "", fmt.Errorf("not nil, or number: %v", e1)
	}

	if isNil(e2) {
		b.WriteString("00")
	} else if ok, n = isNumber(e2); ok {
		b.WriteString(modulate(n))
	} else {
		e2m, err := modulateList(e2)
		if err != nil {
			return "", fmt.Errorf("error modulating %v: %v", e2, err)
		}
		b.WriteString(e2m)
	}
	return b.String(), nil
}
