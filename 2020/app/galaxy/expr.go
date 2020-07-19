package galaxy

import (
	"fmt"
)

type expr interface{}

type expression struct {
	evaled expr
}

type atomType int

type atom struct {
	*expression
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
	*expression
	fun expr
	arg expr
}

func (a atom) String() string {
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
	return "<<UNKNOWN>>"
}

func (a ap) String() string {
	return fmt.Sprintf("(ap %s %s)", a.fun, a.arg)
}
