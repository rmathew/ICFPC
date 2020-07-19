// Expression-evaluation mostly implements the pseudo-code in:
//   https://message-from-space.readthedocs.io/en/latest/implementation.html
package galaxy

import (
	"fmt"
	"log"
)

type expr interface {
	getCached() expr
	setCached(e expr)
	asNum() (int64, error)
	tryEval(fds *FuncDefs) (expr, error)
}

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

func evlOneNum(fds *FuncDefs, e expr) (int64, error) {
	ev, err1 := eval(fds, e)
	if err1 != nil {
		return 0, err1
	}
	num, err2 := ev.asNum()
	if err2 != nil {
		return 0, err2
	}
	return num, nil
}

func evlTwoNums(fds *FuncDefs, e1, e2 expr) (int64, int64, error) {
	n1, err1 := evlOneNum(fds, e1)
	if err1 != nil {
		return 0, 0, err1
	}
	n2, err2 := evlOneNum(fds, e2)
	if err2 != nil {
		return 0, 0, err2
	}
	return n1, n2, nil
}

func (a *atom) getCached() expr {
	return a.exp
}

func (a *atom) setCached(e expr) {
	a.exp = e
}

func (a *atom) asNum() (int64, error) {
	if a.aType == atNumber {
		return a.aNum, nil
	}
	return 0, fmt.Errorf("atom, but not numeric")
}

func (a *atom) tryEval(fds *FuncDefs) (expr, error) {
	if a.exp != nil {
		return a.exp, nil
	}
	if a.aType == atName {
		if e, ok := fds.fds[a.aStr]; ok {
			return e, nil
		}
	}
	return a, nil
}

func (a *ap) getCached() expr {
	return a.exp
}

func (a *ap) setCached(e expr) {
	a.exp = e
}

func (a *ap) asNum() (int64, error) {
	return 0, fmt.Errorf("not an atom")
}

func (a *ap) tryEval(fds *FuncDefs) (expr, error) {
	if a.exp != nil {
		return a.exp, nil
	}
	fe, err := eval(fds, a.fun)
	if err != nil {
		return nil, err
	}
	x := a.arg
	if at, ok := fe.(*atom); ok && at.aType == atName {
		switch at.aStr {
		case "neg":
			num, err1 := evlOneNum(fds, x)
			if err1 != nil {
				return nil, err1
			}
			return mkNum(-num), nil
		case "i":
			return x, nil
		case "nil":
			return mkTrue(), nil
		case "isnil":
			return mkAp(x, mkAp(mkTrue(), mkAp(mkTrue(), mkFalse()))), nil
		case "car":
			return mkAp(x, mkTrue()), nil
		case "cdr":
			return mkAp(x, mkFalse()), nil
		}
	} else if ap1, ok := fe.(*ap); ok {
		f2e, err2 := eval(fds, ap1.fun)
		if err2 != nil {
			return nil, err2
		}
		y := ap1.arg
		if at2, ok2 := f2e.(*atom); ok2 {
			var n1, n2 int64
			var berr error
			switch {
			case at2.aType == atTrue:
				return y, nil
			case at2.aType == atFalse:
				return x, nil
			case at2.aType == atName && at2.aStr == "add":
				n1, n2, berr = evlTwoNums(fds, x, y)
				if berr != nil {
					return nil, berr
				}
				return mkNum(n1 + n2), nil
			case at2.aType == atName && at2.aStr == "mul":
				n1, n2, berr = evlTwoNums(fds, x, y)
				if berr != nil {
					return nil, berr
				}
				return mkNum(n1 * n2), nil
			case at2.aType == atName && at2.aStr == "div":
				n1, n2, berr = evlTwoNums(fds, x, y)
				if berr != nil {
					return nil, berr
				}
				// Deliberate x/y-switch.
				return mkNum(n2 / n1), nil
			case at2.aType == atName && at2.aStr == "lt":
				n1, n2, berr = evlTwoNums(fds, x, y)
				if berr != nil {
					return nil, berr
				}
				// Deliberate x/y-switch.
				if n2 < n1 {
					return mkTrue(), nil
				}
				return mkFalse(), nil
			case at2.aType == atName && at2.aStr == "eq":
				n1, n2, berr = evlTwoNums(fds, x, y)
				if berr != nil {
					return nil, berr
				}
				if n1 == n2 {
					return mkTrue(), nil
				}
				return mkFalse(), nil
			case at2.aType == atName && at2.aStr == "cons":
				// Deliberate x/y-switch.
				return evalCons(fds, y, x)
			}
		} else if ap2, ok2 := f2e.(*ap); ok2 {
			f3e, err3 := eval(fds, ap2.fun)
			if err3 != nil {
				return nil, err3
			}
			z := ap2.arg
			if at3, ok3 := f3e.(*atom); ok3 {
				switch {
				case at3.aType == atName && at3.aStr == "s":
					return mkAp(mkAp(z, x), mkAp(y, x)), nil
				case at3.aType == atName && at3.aStr == "c":
					return mkAp(mkAp(z, x), y), nil
				case at3.aType == atName && at3.aStr == "b":
					return mkAp(z, mkAp(y, x)), nil
				case at3.aType == atCons:
					return mkAp(mkAp(x, z), y), nil
				}
			}
		}
	}

	return a, nil
}

func evalCons(fds *FuncDefs, a, b expr) (expr, error) {
	e1, err1 := eval(fds, a)
	if err1 != nil {
		return nil, err1
	}
	e2, err2 := eval(fds, b)
	if err2 != nil {
		return nil, err2
	}
	res := mkAp(mkAp(mkCons(), e1), e2)
	res.setCached(res)
	return res, nil
}

func eval(fds *FuncDefs, e expr) (expr, error) {
	if e == nil {
		return nil, fmt.Errorf("cannot evaluate NULL expr")
	}
	cached := e.getCached()
	if cached != nil {
		return cached, nil
	}
	ie := e
	for {
		res, err := e.tryEval(fds)
		if err != nil {
			return nil, err
		}
		if res == nil {
			log.Printf("ERROR: Expr %q evaluated to NULL.", e)
			return nil, fmt.Errorf("NULL evaluation-result ")
		}
		if res == e {
			ie.setCached(res)
			return res, nil
		}
		e = res
	}
}
