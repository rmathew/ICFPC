// Expression-evaluation mostly implements the pseudo-code in:
//   https://message-from-space.readthedocs.io/en/latest/implementation.html
package galaxy

import (
	"fmt"
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
	return fmt.Sprintf("(ap %s %s)", a.fun, a.arg)
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

func vec2e(v vect) expr {
	return mkAp(mkAp(mkCons(), mkNum(v.x)), mkNum(v.y))
}
