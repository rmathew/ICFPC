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

func tryAtomEval(fds *FuncDefs, a *atom) (expr, error) {
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

func tryApEval(fds *FuncDefs, a *ap) (expr, error) {
	if a.exp != nil {
		return a.exp, nil
	}

	fe, err := eval(fds, a.fun)
	if err != nil {
		return nil, err
	}

	x := a.arg
	if at, ok := fe.(*atom); ok {
		switch {
		case at.aType == atName && at.aStr == "neg":
			var num int64
			num, err = evalOneNum(fds, x)
			if err != nil {
				return nil, err
			}
			return mkNum(-num), nil
		case at.aType == atName && at.aStr == "i":
			return x, nil
		case at.aType == atNil:
			return mkTrue(), nil
		case at.aType == atName && at.aStr == "isnil":
			return mkAp(x, mkAp(mkTrue(), mkAp(mkTrue(), mkFalse()))), nil
		case at.aType == atName && at.aStr == "car":
			return mkAp(x, mkTrue()), nil
		case at.aType == atName && at.aStr == "cdr":
			return mkAp(x, mkFalse()), nil
		}
	} else if ap1, ok := fe.(*ap); ok {
		var f2e expr
		f2e, err = eval(fds, ap1.fun)
		if err != nil {
			return nil, err
		}
		y := ap1.arg
		if at2, ok2 := f2e.(*atom); ok2 {
			var nx, ny int64
			switch {
			case at2.aType == atTrue:
				return y, nil
			case at2.aType == atFalse:
				return x, nil
			case at2.aType == atName && at2.aStr == "add":
				nx, ny, err = evalTwoNums(fds, x, y)
				if err != nil {
					return nil, err
				}
				return mkNum(nx + ny), nil
			case at2.aType == atName && at2.aStr == "mul":
				nx, ny, err = evalTwoNums(fds, x, y)
				if err != nil {
					return nil, err
				}
				return mkNum(nx * ny), nil
			case at2.aType == atName && at2.aStr == "div":
				nx, ny, err = evalTwoNums(fds, x, y)
				if err != nil {
					return nil, err
				}
				// Deliberate x/y-switch.
				return mkNum(ny / nx), nil
			case at2.aType == atName && at2.aStr == "lt":
				nx, ny, err = evalTwoNums(fds, x, y)
				if err != nil {
					return nil, err
				}
				// Deliberate x/y-switch.
				if ny < nx {
					return mkTrue(), nil
				}
				return mkFalse(), nil
			case at2.aType == atName && at2.aStr == "eq":
				nx, ny, err = evalTwoNums(fds, x, y)
				if err != nil {
					return nil, err
				}
				if nx == ny {
					return mkTrue(), nil
				}
				return mkFalse(), nil
			case at2.aType == atCons:
				// Deliberate x/y-switch.
				return evalCons(fds, y, x)
			}
		} else if ap2, ok2 := f2e.(*ap); ok2 {
			var f3e expr
			f3e, err = eval(fds, ap2.fun)
			if err != nil {
				return nil, err
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
	var ea, eb expr
	var err error
	ea, err = eval(fds, a)
	if err != nil {
		return nil, err
	}
	eb, err = eval(fds, b)
	if err != nil {
		return nil, err
	}
	res := mkAp(mkAp(mkCons(), ea), eb)
	if err = setCached(res, res); err != nil {
		return nil, err
	}
	return res, nil
}

func tryExprEval(fds *FuncDefs, e expr) (expr, error) {
	switch v := e.(type) {
	case *atom:
		return tryAtomEval(fds, v)
	case *ap:
		return tryApEval(fds, v)
	}
	return nil, fmt.Errorf("unknown kind of expr %v in tryExprEval()", e)
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
	const maxIters = 1000000
	for i := 0; i < maxIters; i++ {
		res, err := tryExprEval(fds, e)
		if err != nil {
			return nil, err
		}
		if res == nil {
			log.Printf("ERROR: Expr %q evaluated to NULL.", e)
			return nil, fmt.Errorf("NULL evaluation-result ")
		}
		if eqExprs(res, e) {
			if err = setCached(ie, res); err != nil {
				return nil, err
			}
			return res, nil
		}
		e = res
	}
	return nil, fmt.Errorf("could not converge after %d iterations", maxIters)
}
