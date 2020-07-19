package galaxy

import (
	"log"
)

func DoInteraction(fds *FuncDefs) error {
	var err error
	state := mkNil()
	v := vect{x: 0, y: 0}
	// for {
	click := vec2e(v)
	state /*images=*/, _, err = interact(fds, state, click)
	if err != nil {
		return err
	}
	// v = requestClick()
	log.Printf("New state: %s", state)
	// }

	return nil
}

func interact(fds *FuncDefs, state, event expr) (expr, expr, error) {
	e := mkAp(mkAp(mkName(fds.ip), state), event)
	res, err := eval(fds, e)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("Result: %s", res)
	// TODO: Implement flag-fetching and sending to aliens.
	return nil, nil, nil
}
