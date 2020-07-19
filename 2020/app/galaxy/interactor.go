package galaxy

import (
	"log"
	"fmt"
	"math/bits"
	"strings"
	"strconv"
)

type InterCtx struct {
	BaseUrl string
	ApiKey string
	PlayerKey int64
	Protocol *FuncDefs
}

func DoInteraction(ctx *InterCtx) error {
	var images expr
	var err error
	state := mkNil()
	v := vect{x: 0, y: 0}
	// for {
	click := vec2e(v)
	state, images, err = interact(ctx, state, click)
	if err != nil {
		return err
	}
	// v = requestClick()
	log.Printf("New state: %s", state)
	log.Printf("Images: %s", images)
	// }

	return nil
}

func interact(ctx *InterCtx, state, event expr) (expr, expr, error) {
	fds := ctx.Protocol
	e := mkAp(mkAp(mkName(fds.ip), state), event)
	res, err := eval(fds, e)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("Result: %s", res)

	var resList []expr
	resList, err = extrList(fds, res)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("Result-list has %d elements", len(resList))
	for i, v := range resList {
		log.Printf("#%d: %s", i, v)
	}
	if len(resList) != 3 {
		return nil, nil, fmt.Errorf(
			"result-list has %d elements instead of 3", len(resList))
	}

	var flag int64
	flag, err = evalOneNum(fds, resList[0])
	if err != nil {
		return nil, nil, err
	}
	if flag == 0 {
		return resList[1], resList[2], nil
	}
	// TODO: Implement flag-fetching and sending to aliens.
	return nil, nil, nil
}

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
	if nBits % 4 > 0 {
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
