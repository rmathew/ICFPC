package galaxy

import (
	"fmt"
	"log"
)

type InterCtx struct {
	BaseUrl   string
	ApiKey    string
	PlayerKey int64
	Protocol  *FuncDefs
}

func DoInteraction(ctx *InterCtx) error {
	var images expr
	var err error
	state := mkNil()
	v := vect{x: 0, y: 0}
	// for {
	click := vec2e(v)
	state, images, err = interact(ctx, state, click)
	if err != nil {
		return err
	}
	// v = requestClick()
	log.Printf("New state: %s", state)
	log.Printf("Images: %s", images)
	// }

	return nil
}

func interact(ctx *InterCtx, state, event expr) (expr, expr, error) {
	fds := ctx.Protocol
	e := mkAp(mkAp(mkName(fds.ip), state), event)
	res, err := eval(fds, e)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("Result: %s", res)

	var resList []expr
	resList, err = extrList(fds, res)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("Result-list has %d elements", len(resList))
	for i, v := range resList {
		log.Printf("#%d: %s", i, v)
	}
	if len(resList) != 3 {
		return nil, nil, fmt.Errorf(
			"result-list has %d elements instead of 3", len(resList))
	}

	var flag int64
	flag, err = evalOneNum(fds, resList[0])
	if err != nil {
		return nil, nil, err
	}
	if flag == 0 {
		return resList[1], resList[2], nil
	}
	// TODO: Implement flag-fetching and sending to aliens.
	return nil, nil, nil
}
