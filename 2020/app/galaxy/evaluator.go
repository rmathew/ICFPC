package galaxy

import (
	"fmt"
	"log"
)

func getCached(e expr) (expr, error) {
	switch v := e.(type) {
	case *atom:
		return v.exp, nil
	case *ap:
		return v.exp, nil
	}
	return nil, fmt.Errorf("unknown kind of expr %v in getCached()", e)
}

func setCached(e, ce expr) error {
	switch v := e.(type) {
	case *atom:
		v.exp = ce
		return nil
	case *ap:
		v.exp = ce
		return nil
	}
	return fmt.Errorf("unknown kind of expr %v in setCached()", e)
}

func asNum(e expr) (int64, error) {
	switch v := e.(type) {
	case *atom:
		if v.aType == atNumber {
			return v.aNum, nil
		}
		return 0, fmt.Errorf("%v is an atom, but not numeric", v)
	case *ap:
		return 0, fmt.Errorf("%v is not an atom", v)
	}
	return 0, fmt.Errorf("unknown kind of expr %v in asNum()", e)
}

func evalOneNum(fds *FuncDefs, e expr) (int64, error) {
	var ev expr
	var num int64
	var err error

	ev, err = eval(fds, e)
	if err != nil {
		return 0, err
	}
	num, err = asNum(ev)
	if err != nil {
		return 0, err
	}
	return num, nil
}

func evalTwoNums(fds *FuncDefs, e1, e2 expr) (int64, int64, error) {
	var n1, n2 int64
	var err error

	n1, err = evalOneNum(fds, e1)
	if err != nil {
		return 0, 0, err
	}
	n2, err = evalOneNum(fds, e2)
	if err != nil {
		return 0, 0, err
	}
	return n1, n2, nil
}

func tryEvalAtom(fds *FuncDefs, a *atom) (expr, error) {
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

func tryEvalAp(fds *FuncDefs, a *ap) (expr, error) {
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
			num, err1 := evalOneNum(fds, x)
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
				n1, n2, berr = evalTwoNums(fds, x, y)
				if berr != nil {
					return nil, berr
				}
				return mkNum(n1 + n2), nil
			case at2.aType == atName && at2.aStr == "mul":
				n1, n2, berr = evalTwoNums(fds, x, y)
				if berr != nil {
					return nil, berr
				}
				return mkNum(n1 * n2), nil
			case at2.aType == atName && at2.aStr == "div":
				n1, n2, berr = evalTwoNums(fds, x, y)
				if berr != nil {
					return nil, berr
				}
				// Deliberate x/y-switch.
				return mkNum(n2 / n1), nil
			case at2.aType == atName && at2.aStr == "lt":
				n1, n2, berr = evalTwoNums(fds, x, y)
				if berr != nil {
					return nil, berr
				}
				// Deliberate x/y-switch.
				if n2 < n1 {
					return mkTrue(), nil
				}
				return mkFalse(), nil
			case at2.aType == atName && at2.aStr == "eq":
				n1, n2, berr = evalTwoNums(fds, x, y)
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
	var e1, e2 expr
	var err error
	e1, err = eval(fds, a)
	if err != nil {
		return nil, err
	}
	e2, err = eval(fds, b)
	if err != nil {
		return nil, err
	}
	res := mkAp(mkAp(mkCons(), e1), e2)
	if err = setCached(res, res); err != nil {
		return nil, err
	}
	return res, nil
}

func tryEval(fds *FuncDefs, e expr) (expr, error) {
	switch v := e.(type) {
	case *atom:
		return tryEvalAtom(fds, v)
	case *ap:
		return tryEvalAp(fds, v)
	}
	return nil, fmt.Errorf("unknown kind of expr %v in tryEval()", e)
}

func eval(fds *FuncDefs, e expr) (expr, error) {
	if e == nil {
		return nil, fmt.Errorf("cannot evaluate NULL expr")
	}
	cached, err := getCached(e)
	if err != nil {
		return nil, err
	}
	if cached != nil {
		return cached, nil
	}
	ie := e
	for {
		var res expr
		res, err = tryEval(fds, e)
		if err != nil {
			return nil, err
		}
		if res == nil {
			log.Printf("ERROR: Expr %q evaluated to NULL.", e)
			return nil, fmt.Errorf("NULL evaluation-result ")
		}
		if res == e {
			if err = setCached(ie, res); err != nil {
				return nil, err
			}
			return res, nil
		}
		e = res
	}
}
