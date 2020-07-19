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
		log.Printf("Parsed:\n%s = %s", fd.name, fd.def)
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
		if f, err = getExpression(s); err != nil {
			return nil, fmt.Errorf("error parsing ap-fun: %w", err)
		}
		if a, err = getExpression(s); err != nil {
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
			}
			return nil, fmt.Errorf("unknown token %q", w)
		}
	}
	return tokens, nil
}
