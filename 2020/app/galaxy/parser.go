package galaxy

import (
	"bytes"
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf8"
)

type tokenType int

const (
	tkUnknown tokenType = iota

	tkTrue
	tkFalse
	tkNil
	tkAp
	tkCons
	tkEquals
	tkFunc
	tkNumber
)

type token struct {
	tType tokenType
	tStr  string
	tNum  int64
}

func (t token) String() string {
	switch t.tType {
	case tkTrue:
		return "t"
	case tkFalse:
		return "f"
	case tkNil:
		return "nil"
	case tkAp:
		return "ap"
	case tkCons:
		return "cons"
	case tkEquals:
		return "="
	case tkFunc:
		return t.tStr
	case tkNumber:
		return fmt.Sprintf("%d", t.tNum)
	}
	return "<<UNKNOWN>>"
}

func parseDef(d []byte) error {
	if len(d) == 0 {
		return nil
	}
	tokens, err := getTokens(d)
	if err != nil {
		return err
	}
	fmt.Println("Found tokens %v in this line", tokens)
	return nil
}

func getTokens(d []byte) ([]token, error) {
	ws := bytes.Split(d, []byte(" "))
	tokens := make([]token, 0, len(ws))
	for _, w := range ws {
		switch {
		case len(w) == 0:
			continue
		case bytes.Equal(w, []byte("t")):
			tokens = append(tokens, token{tType: tkTrue})
		case bytes.Equal(w, []byte("f")):
			tokens = append(tokens, token{tType: tkFalse})
		case bytes.Equal(w, []byte("nil")):
			tokens = append(tokens, token{tType: tkNil})
		case bytes.Equal(w, []byte("ap")):
			tokens = append(tokens, token{tType: tkAp})
		case bytes.Equal(w, []byte("cons")):
			tokens = append(tokens, token{tType: tkCons})
		case bytes.Equal(w, []byte("=")):
			tokens = append(tokens, token{tType: tkEquals})
		default:
			r, _ := utf8.DecodeRune(w)
			if r == ':' || unicode.IsLetter(r) {
				tokens = append(tokens, token{tType: tkFunc, tStr: string(w)})
				continue
			}
			if num, err := strconv.ParseInt(string(w), 10, 64); err == nil {
				tokens = append(tokens, token{tType: tkNumber, tNum: num})
				continue
			}
			return nil, fmt.Errorf("Unknown token %q", w)
		}
	}
	return tokens, nil
}
