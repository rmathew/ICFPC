package lol

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

// Room labels are represented as two-bit integers (0..3).
const invRoomLabel int = -1
const maxRoomLabel int = 3

// Doors are numbered from 0..5.
const invDoorNum int = -1
const numDoorsPerRoom int = 6

type edge struct {
	toRoom int
	toDoor int
}

type room struct {
	label int
	doors [numDoorsPerRoom]edge
}

type library struct {
	rooms        []room
	startingRoom int
}

type Mapper struct {
	client        *Client
	maxPathFactor int
}

func initRoom(r *room, label int) {
	r.label = label
	for i := 0; i < numDoorsPerRoom; i++ {
		r.doors[i].toRoom = invRoomLabel
		r.doors[i].toDoor = invDoorNum
	}
}

func newLibrary(size int, startingRoom int) *library {
	lib := &library{rooms: make([]room, size), startingRoom: startingRoom}
	initRoom(&lib.rooms[0], startingRoom)
	for i := 1; i < size; i++ {
		initRoom(&lib.rooms[i], invRoomLabel)
	}
	return lib
}

func NewMapper(clt *Client, maxPathFactor int) *Mapper {
	return &Mapper{client: clt, maxPathFactor: maxPathFactor}
}

func genPlan(maxSteps int) ([]int, string) {
	rand.Seed(time.Now().UnixNano())
	plan := make([]int, maxSteps)
	var b strings.Builder
	b.Grow(maxSteps)
	for i := 0; i < maxSteps; i++ {
		d := rand.Intn(numDoorsPerRoom)
		plan[i] = d
		b.WriteRune(rune('0' + d))
	}
	return plan, b.String()
}

func checkRoomLabel(label int) error {
	if label < 0 || label > maxRoomLabel {
		return fmt.Errorf("invalid room label %d", label)
	}
	return nil
}

func getRoomDoorTgts(rm *room) string {
	tgts := make([]rune, numDoorsPerRoom)
	for i := 0; i < numDoorsPerRoom; i++ {
		tr := rm.doors[i].toRoom
		if tr == invRoomLabel {
			tgts[i] = '.'
		} else {
			tgts[i] = rune('0' + tr)
		}
	}
	return string(tgts)
}

func revEngLibFromPath(plan []int, res [][]int, size int) (*library, error) {
	if len(res) != 1 {
		return nil, fmt.Errorf("expected 1 result; got #d", len(res))
	}
	pathRec := res[0]
	if len(pathRec) != len(plan)+1 {
		return nil, fmt.Errorf("expected %d long path-record; got #d",
			len(plan)+1, len(pathRec))
	}
	if err := checkRoomLabel(pathRec[0]); err != nil {
		return nil, err
	}
	lib := newLibrary(size, pathRec[0])
	currentRoomIdx := 0
	for step, doorNum := range plan {
		nextRoomLabel := pathRec[step+1]
		if err := checkRoomLabel(nextRoomLabel); err != nil {
			return nil, err
		}
		lib.rooms[currentRoomIdx].doors[doorNum].toRoom = nextRoomLabel
		nextRoomIdx := -1
		found := false
		for i := range lib.rooms {
			if lib.rooms[i].label == nextRoomLabel {
				nextRoomIdx = i
				found = true
				break
			}
		}
		if !found {
			for i := range lib.rooms {
				if lib.rooms[i].label == invRoomLabel {
					initRoom(&lib.rooms[i], nextRoomLabel)
					nextRoomIdx = i
					break
				}
			}
		}
		if nextRoomIdx < 0 {
			return nil, fmt.Errorf("no free rooms")
		}
		currentRoomIdx = nextRoomIdx
	}
	for i := range lib.rooms {
		rm := &lib.rooms[i]
		log.Printf("INFO: Room[%d]: %s", rm.label, getRoomDoorTgts(rm))
	}
	return lib, nil
}

func checkMapCoverage(lib *library, size int) error {
	for i := 0; i < size; i++ {
		if lib.rooms[i].label == invRoomLabel {
			return fmt.Errorf("unmapped room %d", i)
		}
	}
	roomFanOut := make([][]int, size)
	for i := range lib.rooms {
		rm := &lib.rooms[i]
		roomFanOut[rm.label] = make([]int, size)
		for i := 0; i < numDoorsPerRoom; i++ {
			if rm.doors[i].toRoom == invRoomLabel {
				return fmt.Errorf("room %d, door %d unmapped", rm.label, i)
			}
			roomFanOut[rm.label][rm.doors[i].toRoom] += 1
		}
	}
	for i, fanOut := range roomFanOut {
		for j, fo := range fanOut {
			if fo != roomFanOut[j][i] {
				return fmt.Errorf("%d->%d = %d, but %d->%d = %d", i, j, fo, j,
					i, roomFanOut[j][i])
			}
		}
	}
	log.Printf("INFO: Map coverage is satisfactory.")
	return nil
}

func assignEntrances(lib *library) error {
	for i := range lib.rooms {
		srcRoom := &lib.rooms[i]
		for j := 0; j < numDoorsPerRoom; j++ {
			srcDoor := &srcRoom.doors[j]
			if srcDoor.toDoor != invDoorNum {
				continue
			}
			tgtRoomIdx := -1
			for k := range lib.rooms {
				if lib.rooms[k].label == srcDoor.toRoom {
					tgtRoomIdx = k
					break
				}
			}
			if tgtRoomIdx < 0 {
				return fmt.Errorf("unable to locate target room %d[%d]->%d",
					srcRoom.label, j, srcDoor.toRoom)
			}
			tgtRoom := &lib.rooms[tgtRoomIdx]
			found := false
			for l := 0; l < numDoorsPerRoom; l++ {
				tgtDoor := &tgtRoom.doors[l]
				if tgtDoor.toRoom == srcRoom.label &&
					tgtDoor.toDoor == invDoorNum {
					srcDoor.toDoor = l
					tgtDoor.toDoor = j
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf(
					"unable to locate entrance in target room %d[%d]->%d[?]",
					srcRoom.label, j, srcDoor.toRoom)
			}
		}
	}
	log.Printf("INFO: Completed inter-room exit to entrance mapping.")
	return nil
}

func genGuessedMap(lib *library) (*GuessedMap, error) {
	gm := &GuessedMap{}

	gm.Rooms = make([]int, len(lib.rooms))
	roomLabelToIdx := make(map[int]int)
	for i := range lib.rooms {
		rm := &lib.rooms[i]
		gm.Rooms[i] = rm.label
		roomLabelToIdx[rm.label] = i
	}
	var ok bool
	gm.StartingRoom, ok = roomLabelToIdx[lib.startingRoom]
	if !ok {
		return nil, fmt.Errorf("unable to map room label %d to its index",
			lib.startingRoom)
	}

	gm.Connections = make([]Connection, len(lib.rooms)*numDoorsPerRoom)
	for i := range lib.rooms {
		rm := &lib.rooms[i]
		var rmIdx int
		rmIdx, ok = roomLabelToIdx[rm.label]
		if !ok {
			return nil, fmt.Errorf("unable to map room label %d to its index",
				rm.label)
		}
		for j := 0; j < numDoorsPerRoom; j++ {
			door := &rm.doors[j]
			conn := &gm.Connections[i*numDoorsPerRoom+j]
			conn.From.Room = rmIdx
			conn.From.Door = j
			conn.To.Room, ok = roomLabelToIdx[door.toRoom]
			if !ok {
				return nil, fmt.Errorf(
					"unable to map room label %d to its index", door.toRoom)
			}
			conn.To.Door = door.toDoor
		}
	}

	return gm, nil
}

func (m *Mapper) Map(prob string, size int) error {
	if size < 1 {
		return fmt.Errorf("non-positive problem-size: %d", size)
	}
	if err := m.client.SelectProblem(prob); err != nil {
		return fmt.Errorf("unable to select problem: %w", err)
	}

	plan, planStr := genPlan(m.maxPathFactor * size)
	log.Printf("INFO: Plan %s", planStr)

	res, err := m.client.Explore([]string{planStr})
	if err != nil {
		return fmt.Errorf("unable to explore: %w", err)
	}

	var lib *library
	lib, err = revEngLibFromPath(plan, res, size)
	if err != nil {
		return err
	}

	if err = checkMapCoverage(lib, size); err != nil {
		return err
	}
	if err = assignEntrances(lib); err != nil {
		return err
	}

	var guessedMap *GuessedMap
	if guessedMap, err = genGuessedMap(lib); err != nil {
		return err
	}
	success := false
	if success, err = m.client.Guess(guessedMap); err != nil {
		return err
	}
	if success {
		log.Print("INFO: Success! \\o/")
	} else {
		log.Print("INFO: Failure. /o\\")
	}

	return nil
}
