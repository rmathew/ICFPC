package galaxy

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
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
	tkCons
	tkNumber
	tkEquals
	tkAp
	tkName
)

type token struct {
	tType tokenType
	tStr  string
	tNum  int64
}

type fnDef struct {
	name string
	def  expr
}

type parseState struct {
	tokens []token
	idx    int
}

func (t token) String() string {
	switch t.tType {
	case tkTrue:
		return "t"
	case tkFalse:
		return "f"
	case tkNil:
		return "nil"
	case tkCons:
		return "cons"
	case tkAp:
		return "ap"
	case tkEquals:
		return "="
	case tkNumber:
		return fmt.Sprintf("%d", t.tNum)
	case tkName:
		return t.tStr
	}
	return "<<UNKNOWN>>"
}

func strToExpr(s string) (expr, error) {
	var e expr
	var ps parseState
	var err error

	ps.tokens, err = getTokens([]byte(s))
	if err != nil {
		return nil, err
	}
	e, err = parseExpression(&ps)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func ParseFunctions(f string) (*FuncDefs, error) {
	log.Printf("Reading the interaction-protocol from %q.", f)
	file, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fds := &FuncDefs{ip: "", fds: make(map[string]expr)}

	scanner := bufio.NewScanner(file)
	for ln := 1; scanner.Scan(); ln++ {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var fd fnDef
		if fd, err = parseFuncDef(line); err != nil {
			log.Printf("Parse-error at line #%d: %v", ln, err)
			log.Printf("Line #%d:\n\t%s", ln, string(line))
			return nil, err
		}
		// log.Printf("Parsed:\n%s = %s", fd.name, fd.def)
		fds.fds[fd.name] = fd.def
		fds.ip = fd.name
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}

	log.Printf("Found %d func-def(s) in %q.", len(fds.fds), f)
	return fds, nil
}

func parseFuncDef(d []byte) (fnDef, error) {
	var f fnDef
	var err error
	var ps parseState

	ps.tokens, err = getTokens(d)
	if err != nil {
		return f, err
	}

	f.name, err = parseFuncName(&ps)
	if err != nil {
		return f, err
	}

	err = parseEquals(&ps)
	if err != nil {
		return f, err
	}

	f.def, err = parseExpression(&ps)
	if err != nil {
		return f, err
	}

	return f, nil
}

func parseFuncName(ps *parseState) (string, error) {
	if ps.idx < len(ps.tokens) && ps.tokens[ps.idx].tType == tkName {
		n := ps.tokens[ps.idx].tStr
		ps.idx++
		return n, nil
	}
	return "", fmt.Errorf("expected function-name not found")
}

func parseEquals(ps *parseState) error {
	if ps.idx < len(ps.tokens) && ps.tokens[ps.idx].tType == tkEquals {
		ps.idx++
		return nil
	}
	return fmt.Errorf("expected equals not found")
}

func parseExpression(ps *parseState) (expr, error) {
	if ps.idx >= len(ps.tokens) {
		return nil, fmt.Errorf("unexpected end of expression")
	}
	tk := &ps.tokens[ps.idx]
	ps.idx++

	switch tk.tType {
	case tkTrue:
		return mkTrue(), nil
	case tkFalse:
		return mkFalse(), nil
	case tkNil:
		return mkNil(), nil
	case tkCons:
		return mkCons(), nil
	case tkNumber:
		return mkNum(tk.tNum), nil
	case tkName:
		return mkName(tk.tStr), nil
	case tkAp:
		var f, a expr
		var err error
		if f, err = parseExpression(ps); err != nil {
			return nil, fmt.Errorf("error parsing ap-fun: %w", err)
		}
		if a, err = parseExpression(ps); err != nil {
			return nil, fmt.Errorf("error parsing ap-arg: %w", err)
		}
		return mkAp(f, a), nil
	}
	return nil, fmt.Errorf("unexpected token %q in expression", tk)
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
				tokens = append(tokens, token{tType: tkName, tStr: string(w)})
				continue
			}
			if num, err := strconv.ParseInt(string(w), 10, 64); err == nil {
				tokens = append(tokens, token{tType: tkNumber, tNum: num})
				continue
			} else {
				nerr := err.(*strconv.NumError)
				if nerr.Err == strconv.ErrRange {
					return nil, fmt.Errorf("number %q too large", nerr.Num)
				}
			}
			return nil, fmt.Errorf("unknown token %q", w)
		}
	}
	return tokens, nil
}
