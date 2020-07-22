package galaxy

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/bits"
	"net/http"
	"net/url"
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
		return 0, fmt.Errorf("too few runes (%d) to decodeNumber", len(r))
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

func modulateListElt(e expr) (string, error) {
	if isNil(e) {
		return "00", nil
	} else if ok, n := isNumber(e); ok {
		return modulate(n), nil
	} else if s, err := modulateList(e); err == nil {
		return s, nil
	} else {
		return "", fmt.Errorf("error modulating list-element %v: %v", e, err)
	}
	return "", nil
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

	if s, err := modulateListElt(e1); err == nil {
		b.WriteString(s)
	} else {
		return "", nil
	}
	if s, err := modulateListElt(e2); err == nil {
		b.WriteString(s)
	} else {
		return "", nil
	}

	return b.String(), nil
}

func demodulateListElt(r []rune) (expr, int, error) {
	var e expr
	idx := 0
	var n int64
	var err error
	if r[0] == '0' && r[1] == '0' {
		e = mkNil()
		idx = 2
	} else if idx, err = decodeNumber(r, &n); err == nil {
		e = mkNum(n)
	} else if e, idx, err = demodulateList(r); err != nil {
		return nil, 0, fmt.Errorf(
			"error demodulating list %q: %v", string(r), err)
	}
	return e, idx, nil
}

func demodulateList(r []rune) (expr, int, error) {
	if len(r) < 2 {
		return nil, 0, fmt.Errorf("too few runes (%d) to demodulate", len(r))
	}
	if r[0] != '1' || r[1] != '1' {
		return nil, 0, fmt.Errorf("not encoding a list %q", string(r[:2]))
	}
	idx := 2
	if len(r) == 2 {
		return mkNil(), idx, nil
	}

	// "11", plus at least two bits each for the pair.
	if len(r) < 6 {
		return nil, 0, fmt.Errorf("too few runes %d; need >= 6", len(r))
	}

	inc := 0
	var e1, e2 expr
	var err error
	if e1, inc, err = demodulateListElt(r[idx:]); err == nil {
		idx += inc
	} else {
		return nil, 0, err
	}
	if e2, inc, err = demodulateListElt(r[idx:]); err == nil {
		idx += inc
	} else {
		return nil, 0, err
	}

	return mkPair(e1, e2), idx, nil
}

func encodeMsg(e expr) (string, error) {
	if ok, n := isNumber(e); ok {
		return modulate(n), nil
	}
	return modulateList(e)
}

func decodeMsg(r []rune) (expr, error) {
	if r[0] == '1' && r[1] == '1' {
		e, _, err := demodulateList(r)
		return e, err
	}
	n, err := demodulate(r)
	if err != nil {
		return nil, err
	}
	return mkNum(n), nil
}

func sendToAliens(ctx *InterCtx, e expr) (expr, error) {
	var err error
	var msg string
	if msg, err = encodeMsg(e); err != nil {
		return nil, err
	}

	var u *url.URL
	u, err = url.Parse(ctx.BaseUrl)
	if err != nil {
		return nil, err
	}
	u.Path = "/aliens/send"
	if len(ctx.ApiKey) > 0 {
		q := u.Query()
		q.Add("apiKey", ctx.ApiKey)
		u.RawQuery = q.Encode()
	}
	us := u.String()
	log.Printf("Sending message %q to aliens using URL %q.", msg, us)

	var res *http.Response
	res, err = http.Post(us, "text/plain", strings.NewReader(msg))
	if err != nil {
		log.Printf("Server-error: %v", res)
		return nil, err
	}
	log.Printf("Status: %q", res.Status)
	defer res.Body.Close()

	var body []byte
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	bs := string(body)
	log.Printf("Received: %q", bs)
	return decodeMsg([]rune(bs))
}
