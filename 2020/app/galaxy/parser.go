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

type funcDef struct {
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

func parse(f string) (map[string]expr, error) {
	log.Printf("Reading the interaction-protocol from %q.", f)
	file, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	funcDefs := make(map[string]expr)

	scanner := bufio.NewScanner(file)
	for ln := 1; scanner.Scan(); ln++ {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var fd funcDef
		if fd, err = parseDef(line); err != nil {
			log.Printf("Parse-error at line #%d: %v", ln, err)
			log.Printf("Line #%d:\n\t%s", ln, string(line))
			return nil, err
		}
		log.Printf("Parsed:\n%s = %s", fd.name, fd.def)
		funcDefs[fd.name] = fd.def
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return funcDefs, nil
}

func parseDef(d []byte) (funcDef, error) {
	var f funcDef
	var err error
	var s parseState

	s.tokens, err = getTokens(d)
	if err != nil {
		return f, err
	}

	f.name, err = getFuncName(&s)
	if err != nil {
		return f, err
	}

	err = expectEquals(&s)
	if err != nil {
		return f, err
	}

	f.def, err = getExpression(&s)
	if err != nil {
		return f, err
	}

	return f, nil
}

func getFuncName(s *parseState) (string, error) {
	if s.idx < len(s.tokens) && s.tokens[s.idx].tType == tkName {
		n := s.tokens[s.idx].tStr
		s.idx++
		return n, nil
	}
	return "", fmt.Errorf("expected function-name not found")
}

func expectEquals(s *parseState) error {
	if s.idx < len(s.tokens) && s.tokens[s.idx].tType == tkEquals {
		s.idx++
		return nil
	}
	return fmt.Errorf("expected equals not found")
}

func getExpression(s *parseState) (expr, error) {
	if s.idx >= len(s.tokens) {
		return nil, fmt.Errorf("unexpected end of expression")
	}
	tk := &s.tokens[s.idx]
	s.idx++

	var e expr
	switch tk.tType {
	case tkTrue:
		e = &atom{expression: &expression{}, aType: atTrue}
	case tkFalse:
		e = &atom{expression: &expression{}, aType: atFalse}
	case tkNil:
		e = &atom{expression: &expression{}, aType: atNil}
	case tkCons:
		e = &atom{expression: &expression{}, aType: atCons}
	case tkNumber:
		e = &atom{expression: &expression{}, aType: atNumber, aNum: tk.tNum}
	case tkName:
		e = &atom{expression: &expression{}, aType: atName, aStr: tk.tStr}
	case tkAp:
		a := &ap{expression: &expression{}}
		var err error
		if a.fun, err = getExpression(s); err != nil {
			return nil, fmt.Errorf("error parsing ap-fun: %w", err)
		}
		if a.arg, err = getExpression(s); err != nil {
			return nil, fmt.Errorf("error parsing ap-arg: %w", err)
		}
		e = a
	default:
		return nil, fmt.Errorf("unexpected token %q in expression", tk)
	}
	return e, nil
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
			}
			return nil, fmt.Errorf("unknown token %q", w)
		}
	}
	return tokens, nil
}
