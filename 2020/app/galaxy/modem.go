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

func decodeNumber(r []rune, n *int64) (int, error) {
	if len(r) < 3 {
		return 0, fmt.Errorf("too few runes (%d) to demodulate", len(r))
	}
	var mult int64
	if r[0] == '0' && r[1] == '1' {
		mult = +1
	} else if r[0] == '1' && r[1] == '0' {
		mult = -1
	} else {
		return 0, fmt.Errorf("not encoding a number %q", string(r[:2]))
	}

	numNybbles := 0
	for i := 2; r[i] == '1' && i < len(r); i++ {
		numNybbles++
	}

	// Need space for at least "01"/"10" and "0"-terminated size in unary as
	// well as the actual number encoded in blocks of nybbles (4 bits).
	mSize := 2 + numNybbles + 1 + 4*numNybbles
	if len(r) < mSize {
		return 0, fmt.Errorf("too few runes %d; need %d", len(r), mSize)
	}

	var pn int64 = 0
	if numNybbles > 0 {
		var err error
		idx := 2 + numNybbles + 1
		pn, err = strconv.ParseInt(string(r[idx:idx+4*numNybbles]), 2, 64)
		if err != nil {
			return 0, err
		}
	}
	*n = pn * mult
	return mSize, nil
}

func demodulate(r []rune) (int64, error) {
	var n int64
	if _, err := decodeNumber(r, &n); err != nil {
		return 0, err
	}
	return n, nil
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

func demodulateList(r []rune) (expr, error) {
	if len(r) < 2 {
		return nil, fmt.Errorf("too few runes (%d) to demodulate", len(r))
	}
	if r[0] != '1' || r[1] != '1' {
		return nil, fmt.Errorf("not encoding a list %q", string(r[:2]))
	}
	if len(r) == 2 {
		return mkNil(), nil
	}
	// "11", plus at least two bits each for the pair.
	if len(r) < 6 {
		return 0, fmt.Errorf("too few runes %d; need >= 6", len(r))
	}

	idx := 2
	inc := 0
	var e1, e2 expr
	var n int64
	var err error
	if r[idx] == '0' && r[idx+1] == '0' {
		e1 = mkNil()
		idx += 2
	} else if inc, err = decodeNumber(r[idx:], &n); err == nil {
		e1 = mkNum(n)
		idx += inc
	} else {
		return nil, fmt.Errorf("nil/number not at %d in %q", idx, string(r))
	}

	if r[idx] == '0' && r[idx+1] == '0' {
		e2 = mkNil()
		idx += 2
	} else if inc, err = decodeNumber(r[idx:], &n); err == nil {
		e2 = mkNum(n)
		idx += inc
	} else {
		e2, err = demodulateList(r[idx:])
		if err != nil {
			return nil, fmt.Errorf(
				"error demodulating list %q at %d: %v", string(r), idx, err)
		}
	}
	return mkPair(e1, e2), nil
}
