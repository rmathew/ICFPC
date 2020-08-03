// Expression-evaluation mostly implements the pseudo-code in:
//   https://message-from-space.readthedocs.io/en/latest/implementation.html
package galaxy

import (
	"fmt"
	"strings"
)

type expr interface{}

type FuncDefs struct {
	ip  string
	fds map[string]expr
}

type atomType int

type atom struct {
	exp expr

	aType atomType
	aStr  string
	aNum  int64
}

const (
	atUnknown atomType = iota

	atTrue
	atFalse
	atNil
	atCons
	atNumber
	atName
)

type ap struct {
	exp expr

	fun expr
	arg expr
}

type vect struct {
	x, y int64
}

func (a *atom) String() string {
	if a == nil {
		return "<<NIL atom>>"
	}
	switch a.aType {
	case atTrue:
		return "t"
	case atFalse:
		return "f"
	case atNil:
		return "nil"
	case atCons:
		return "cons"
	case atNumber:
		return fmt.Sprintf("%d", a.aNum)
	case atName:
		return a.aStr
	}
	return "<<UNKNOWN atom>>"
}

func (a *ap) String() string {
	if a == nil {
		return "<<NIL ap>>"
	}
	if ok, _, _ := isPair(a); ok {
		if s, err := listToStr(a); err == nil {
			return s
		}
		if v, err := e2vec(a); err == nil {
			return fmt.Sprintf("%v", v)
		}
		return fmt.Sprintf("(ap %s %s)", a.fun, a.arg)
	}
	return fmt.Sprintf("(ap %s %s)", a.fun, a.arg)
}

func (v *vect) String() string {
	if v == nil {
		return "<<NIL vect>>"
	}
	return fmt.Sprintf("(%d, %d)", v.x, v.y)
}

func mkNil() expr {
	return &atom{exp: nil, aType: atNil}
}

func mkTrue() expr {
	return &atom{exp: nil, aType: atTrue}
}

func mkFalse() expr {
	return &atom{exp: nil, aType: atFalse}
}

func mkNum(n int64) expr {
	return &atom{exp: nil, aType: atNumber, aNum: n}
}

func mkName(s string) expr {
	return &atom{exp: nil, aType: atName, aStr: s}
}

func mkCons() expr {
	return &atom{exp: nil, aType: atCons}
}

func mkAp(f, a expr) expr {
	return &ap{exp: nil, fun: f, arg: a}
}

func mkPair(e1, e2 expr) expr {
	return mkAp(mkAp(mkCons(), e1), e2)
}

func eqExprs(e1, e2 expr) bool {
	if e1 == e2 {
		return true
	}
	if e1 == nil || e2 == nil {
		return false
	}
	switch v1 := e1.(type) {
	case *atom:
		switch v2 := e2.(type) {
		case *atom:
			return v1.aType == v2.aType && v1.aNum == v2.aNum &&
				v1.aStr == v2.aStr
		}
		return false
	case *ap:
		switch v2 := e2.(type) {
		case *ap:
			return eqExprs(v1.fun, v2.fun) && eqExprs(v1.arg, v2.arg)
		}
		return false
	}
	return false
}

func isAtomOfType(e expr, at atomType) bool {
	if e == nil {
		return false
	}
	if v, ok := e.(*atom); ok {
		return v.aType == at
	}
	return false
}

func isNil(e expr) bool {
	return isAtomOfType(e, atNil)
}

func isCons(e expr) bool {
	return isAtomOfType(e, atCons)
}

func isNumber(e expr) (bool, int64) {
	if e == nil {
		return false, 0
	}
	if v, ok := e.(*atom); ok && v.aType == atNumber {
		return true, v.aNum
	}
	return false, 0
}

func isName(e expr) (bool, string) {
	if e == nil {
		return false, ""
	}
	if v, ok := e.(*atom); ok && v.aType == atName {
		return true, v.aStr
	}
	return false, ""
}

func isPair(e expr) (bool, expr, expr) {
	if e == nil {
		return false, nil, nil
	}
	v, vOk := e.(*ap)
	if !vOk {
		return false, nil, nil
	}
	vf, vfOk := v.fun.(*ap)
	if !vfOk {
		return false, nil, nil
	}
	if !isCons(vf.fun) {
		return false, nil, nil
	}
	return true, vf.arg, v.arg
}

func vec2e(v *vect) expr {
	return mkPair(mkNum(v.x), mkNum(v.y))
}

func e2vec(e expr) (*vect, error) {
	var e1, e2 expr
	var n1, n2 int64
	var ok bool
	if ok, e1, e2 = isPair(e); !ok {
		return nil, fmt.Errorf("not a pair: %v", e)
	}
	if ok, n1 = isNumber(e1); !ok {
		return nil, fmt.Errorf("%v not a number in %v", e1, e)
	}
	if ok, n2 = isNumber(e2); !ok {
		return nil, fmt.Errorf("%v not a number in %v", e2, e)
	}
	return &vect{x: n1, y: n2}, nil
}

func extrList(e expr) ([]expr, error) {
	if e == nil {
		return nil, fmt.Errorf("NULL expression for list-extraction")
	}
	list := make([]expr, 0)
	if isNil(e) {
		return list, nil
	}

	var e1, e2 expr
	var ok bool
	if ok, e1, e2 = isPair(e); !ok {
		return nil, fmt.Errorf("not a pair: %v", e)
	}

	list = append(list, e1)
	if tailList, err := extrList(e2); err == nil {
		return append(list, tailList...), nil
	} else {
		return nil, err
	}
}

func listToStr(e expr) (string, error) {
	var el []expr
	var err error
	if el, err = extrList(e); err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteRune('[')
	if len(el) > 0 {
		les := make([]string, 0, len(el))
		for _, v := range el {
			var s string
			if s, err = listToStr(v); err != nil {
				s = fmt.Sprintf("%v", v)
			}
			les = append(les, s)
		}
		b.WriteString(strings.Join(les, ", "))
	}
	b.WriteRune(']')
	return b.String(), nil
}
