package galaxy

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

type InterCtx struct {
	BaseUrl   string
	ApiKey    string
	PlayerKey int64
	Protocol  *FuncDefs
	Viewer    *GalaxyViewer
}

func DoInteraction(ctx *InterCtx) error {
	var images expr
	var err error
	state := mkNil()
	v := &vect{x: 0, y: 0}
	run := true

	const maxIters = 1000000
	for i := 0; i < maxIters && run; i++ {
		click := vec2e(v)
		t0 := time.Now()
		log.Printf("BEGIN interact(): #%d", i)
		state, images, err = interact(ctx, state, click)
		if err != nil {
			return err
		}

		err = drawImages(ctx, images)
		if err != nil {
			return err
		}
		log.Printf("END interact(): #%d after %v", i, time.Since(t0))

		run, v = requestClick(ctx)
		// log.Printf("New state: %s", state)
		// log.Printf("Images: %s", images)
	}

	return nil
}

func interact(ctx *InterCtx, state, event expr) (expr, expr, error) {
	fds := ctx.Protocol

	st := state
	ev := event
	run := true
	const maxIters = 1000000
	for i := 0; i < maxIters && run; i++ {
		// log.Printf("interact(): #%d", i)

		e := mkAp(mkAp(mkName(fds.ip), st), ev)
		res, err := eval(fds, e)
		if err != nil {
			return nil, nil, err
		}
		// log.Printf("Result: %s", res)

		var resList []expr
		resList, err = extrList(fds, res)
		if err != nil {
			return nil, nil, err
		}
		/*
			log.Printf("Result-list has %d elements", len(resList))
			for i, v := range resList {
				log.Printf("#%d: %s", i, v)
			}
		*/
		if len(resList) != 3 {
			return nil, nil, fmt.Errorf(
				"result-list has %d elements instead of 3", len(resList))
		}
		newSt := resList[1]
		data := resList[2]

		var flag int64
		flag, err = evalOneNum(fds, resList[0])
		if err != nil {
			return nil, nil, err
		}
		if flag == 0 {
			return newSt, data, nil
		}

		var ar expr
		ar, err = sendToAliens(ctx, data)
		if err != nil {
			return nil, nil, err
		}

		st = newSt
		ev = ar

		run = !shouldBreak(ctx)
	}
	return nil, nil, fmt.Errorf("incomplete after %d iterations", maxIters)
}

func extrDrawLists(fds *FuncDefs, imgs expr) ([][]*vect, error) {
	var il []expr
	var err error
	if il, err = extrList(fds, imgs); err != nil {
		log.Printf("Error converting expr to images list: %v", err)
		return nil, err
	}

	dls := make([][]*vect, 0, len(il))
	// log.Printf("Images-list has %d elements", len(il))
	for i, ili := range il {
		var vl []expr
		if vl, err = extrList(fds, ili); err != nil {
			log.Printf("Error converting img[%d] to vectors: %v", i, err)
			return nil, err
		}
		dls = append(dls, make([]*vect, 0, len(vl)))
		for j, vlj := range vl {
			var v *vect
			if v, err = e2vec(vlj); err != nil {
				log.Printf(
					"Error converting img[%d][%d] to vector: %v", i, j, err)
				return nil, err
			}
			// log.Printf("Vec[%d][%d]: %v", i, j, v)
			dls[i] = append(dls[i], v)
		}
	}
	return dls, nil
}

func drawImages(ctx *InterCtx, imgs expr) error {
	dls, err := extrDrawLists(ctx.Protocol, imgs)
	if err != nil {
		return err
	}
	/*
		for i, vi := range dls {
			for j, vj := range vi {
				log.Printf("img[%d][%d]=%v", i, j, vj)
			}
		}
	*/
	// log.Printf("#Imgs: %d", len(dls))
	if ctx.Viewer != nil {
		ctx.Viewer.update(dls)
	}
	return nil
}

func shouldBreak(ctx *InterCtx) bool {
	if ctx.Viewer != nil {
		return ctx.Viewer.shouldBreak()
	}
	return false
}

func requestClick(ctx *InterCtx) (bool, *vect) {
	if ctx.Viewer != nil {
		return ctx.Viewer.requestClick()
	}
	return true, &vect{x: int64(rand.Intn(64)), y: int64(rand.Intn(64))}
}
