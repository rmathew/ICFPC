package lol

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

const invRoomIdx int = -1

// Room labels are represented as two-bit integers (0..3).
const invRoomLabel int = -1
const maxRoomLabel int = 3

// Doors are numbered from 0..5.
const invDoorNum int = -1
const numDoorsPerRoom int = 6

type problemInfo struct {
	size          int
	maxPathFactor int
	needsMarkers  bool
}

type roomRecord struct {
	idx       int
	label     int
	exitDoor  int
	exitLabel int
}

type edge struct {
	toRoomIdx   int
	toRoomLabel int
	toRoomDoor  int
}

type room struct {
	label int
	doors [numDoorsPerRoom]edge
}

type library struct {
	rooms []room
}

type phantomRoom struct {
	partRoom room
	supRecs  []int
}

type Mapper struct {
	client *Client
}

func initRoom(r *room, label int) {
	r.label = label
	for i := 0; i < numDoorsPerRoom; i++ {
		r.doors[i].toRoomIdx = invRoomIdx
		r.doors[i].toRoomLabel = invRoomLabel
		r.doors[i].toRoomDoor = invDoorNum
	}
}

func newLibrary(size int) *library {
	lib := &library{rooms: make([]room, size)}
	for i := 0; i < size; i++ {
		initRoom(&lib.rooms[i], invRoomLabel)
	}
	return lib
}

func NewMapper(clt *Client) *Mapper {
	return &Mapper{client: clt}
}

func getProblemInfo(prob string) (*problemInfo, error) {
	// The problems from the Lightning round (probatio, primus, ..., quintus)
	// can use use 18x paths relative to their size and do not need the use of
	// charcoal markers to be solved, while it is 6x for the ones posted
	// thereafter (aleph, beth, ..., iod) and they do need the use of markers.
	probToInfo := map[string]*problemInfo{
		"probatio": &problemInfo{3, 18, false},
		"primus":   &problemInfo{6, 18, false},
		"secundus": &problemInfo{12, 18, false},
		"tertius":  &problemInfo{18, 18, false},
		"quartus":  &problemInfo{24, 18, false},
		"quintus":  &problemInfo{30, 18, false},
		"aleph":    &problemInfo{12, 6, true},
		"beth":     &problemInfo{24, 6, true},
		"gimel":    &problemInfo{36, 6, true},
		"daleth":   &problemInfo{48, 6, true},
		"he":       &problemInfo{60, 6, true},
		"vau":      &problemInfo{18, 6, true},
		"zain":     &problemInfo{36, 6, true},
		"hhet":     &problemInfo{54, 6, true},
		"teth":     &problemInfo{72, 6, true},
		"iod":      &problemInfo{90, 6, true},
	}
	if pi, ok := probToInfo[prob]; ok {
		return pi, nil
	}
	return nil, fmt.Errorf("unknown problem '%s'", prob)
}

func genPlan(pi *problemInfo) ([]int, string) {
	maxSteps := pi.size * pi.maxPathFactor
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

func getRoomRecords(plan []int, pathRec []int) ([]roomRecord, error) {
	if len(pathRec) != len(plan)+1 {
		return nil, fmt.Errorf("expected path-record to be of size %d; got #d",
			len(plan)+1, len(pathRec))
	}
	for i, l := range pathRec {
		if l < 0 || l > maxRoomLabel {
			return nil, fmt.Errorf("invalid room-label %d in result[%d]", l, i)
		}
	}
	roomRecs := make([]roomRecord, len(plan))
	for step, doorNum := range plan {
		rr := &roomRecs[step]
		rr.idx = invRoomIdx
		rr.label = pathRec[step]
		rr.exitDoor = doorNum
		rr.exitLabel = pathRec[step+1]
	}
	if len(roomRecs) == 0 {
		return nil, fmt.Errorf("empty set of room-records")
	}
	return roomRecs, nil
}

func roomLabelToLetter(l int) rune {
	if l < 0 || l > maxRoomLabel {
		return '?'
	}
	roomLabels := [maxRoomLabel + 1]rune{'A', 'B', 'C', 'D'}
	return roomLabels[l]
}

func isRoomLabelLetter(b byte) bool {
	return b >= 'A' && b <= 'D'
}

func tracePath(rrs []roomRecord) string {
	var b strings.Builder
	b.Grow(2*len(rrs) + 1)
	b.WriteRune(roomLabelToLetter(rrs[0].label))
	for _, rr := range rrs {
		b.WriteRune(rune('0' + rr.exitDoor))
		b.WriteRune(roomLabelToLetter(rr.exitLabel))
	}
	return b.String()
}

func newPhantomRooms(roomRecs []roomRecord, size int) map[int]*phantomRoom {
	pRooms := make(map[int]*phantomRoom, len(roomRecs))
	nPlaceholders := 0
	for i := range roomRecs {
		idx := size + nPlaceholders
		rr := &roomRecs[i]
		rr.idx = idx
		pr := &phantomRoom{supRecs: []int{i}}
		initRoom(&pr.partRoom, rr.label)
		pr.partRoom.doors[rr.exitDoor].toRoomLabel = rr.exitLabel
		pRooms[idx] = pr
		nPlaceholders++
	}
	log.Printf("INFO: %d phantom rooms created.", len(pRooms))
	return pRooms
}

func mergeDoors(src, tgt *edge) error {
	srl := src.toRoomLabel
	trl := tgt.toRoomLabel
	if srl != trl {
		if trl == invRoomLabel {
			tgt.toRoomLabel = srl
		} else if srl != invRoomLabel {
			return fmt.Errorf("door-label mis-match (%d vs %d)", srl, trl)
		}
	}

	sri := src.toRoomIdx
	tri := tgt.toRoomIdx
	if sri != tri {
		if tri == invRoomIdx {
			tgt.toRoomIdx = sri
		} else if sri != invRoomIdx {
			return fmt.Errorf("door-index mis-match (%d vs %d)", sri, tri)
		}
	}

	srd := src.toRoomDoor
	trd := tgt.toRoomDoor
	if srd != trd {
		if trd == invDoorNum {
			tgt.toRoomDoor = srd
		} else if srd != invDoorNum {
			return fmt.Errorf("door-target mis-match (%d vs %d)", srd, trd)
		}
	}
	return nil
}

func mergePhantomRooms(pRooms map[int]*phantomRoom, idxs []int,
	rrs []roomRecord) error {
	if len(idxs) < 2 {
		return nil
	}
	mainIdx := idxs[0]
	mainPr := pRooms[mainIdx]
	mr := &mainPr.partRoom
	for i := 1; i < len(idxs); i++ {
		subIdx := idxs[i]
		subPr := pRooms[subIdx]
		sr := &subPr.partRoom
		if sr.label != mr.label {
			return fmt.Errorf("label mis-match (%d vs %d)", sr.label, mr.label)
		}
		for j := 0; j < numDoorsPerRoom; j++ {
			if err := mergeDoors(&sr.doors[j], &mr.doors[j]); err != nil {
				return err
			}
		}
		for _, j := range subPr.supRecs {
			if rrs[j].idx != subIdx {
				return fmt.Errorf("index mis-match (%d vs %d)", rrs[j].idx,
					subIdx)
			}
			rrs[j].idx = mainIdx
		}
		mainPr.supRecs = append(mainPr.supRecs, subPr.supRecs...)

		delete(pRooms, subIdx)
	}
	return nil
}

/* BEGIN: DEPRECATED
func getCandidateRoomIdxForRec(lib *library, rr *roomRecord) (int, error) {
	for i := range lib.rooms {
		rm := &lib.rooms[i]
		if rm.label == invRoomLabel {
			return invRoomIdx, nil
		}
		if rm.label != rr.label {
			continue
		}
		dr := rm.doors[rr.exitDoor]
		if dr.toRoomLabel == invRoomLabel || dr.toRoomLabel == rr.exitLabel {
			return i, nil
		}
	}
	return invRoomIdx, fmt.Errorf("exhausted all library rooms")
}

func setPlaceholderRoomIdxs(roomRecs []roomRecord, size int) int {
	nPlaceholders := 0
	for i := range roomRecs {
		rr := &roomRecs[i]
		if rr.idx == invRoomIdx {
			rr.idx = size + nPlaceholders
			nPlaceholders++
		}
	}
	return nPlaceholders
}

func useLabelAndDoorInRoomRecs(lib *library, roomRecs []roomRecord) error {
	nRooms := 0
	for i := range roomRecs {
		rr := &roomRecs[i]
		idx, err := getCandidateRoomIdxForRec(lib, rr)
		if err != nil {
			return err
		}
		if idx != invRoomIdx {
			continue
		}
		rr.idx = nRooms
		rm := &lib.rooms[nRooms]
		rm.label = rr.label
		rm.doors[rr.exitDoor].toRoomLabel = rr.exitLabel
		nRooms++
	}
	log.Printf("INFO: Added %d of %d rooms on the first pass.", nRooms,
		len(lib.rooms))

	nP := setPlaceholderRoomIdxs(roomRecs, len(lib.rooms))
	log.Printf("INFO: Set %d placeholders across %d room-indexes.", nP,
		len(roomRecs))

	return nil
}
END: DEPRECATED */

func findAllDupeSubPaths(pt string, minSubPathLen int) map[string][]int {
	minSubStrLen := minSubPathLen*2 + 1
	if len(pt) < minSubStrLen {
		return nil
	}
	allSubs := make(map[string][]int)
	for i := 0; i <= len(pt)-minSubStrLen; i++ {
		for j := minSubStrLen; i+j <= len(pt); j += 2 {
			ss := pt[i : i+j]
			if !isRoomLabelLetter(ss[0]) || !isRoomLabelLetter(ss[len(ss)-1]) {
				continue
			}
			allSubs[ss] = append(allSubs[ss], i)
		}
	}
	// Note that we will find sub-paths that are themselves sub-paths of others:
	//     DupeSubPath 'B3C5A3A' at '[90 104 160]'
	//     DupeSubPath 'B3C5A3A2C' at '[90 160]'
	//     DupeSubPath 'C5A3A2C' at '[92 162]'
	// We will also find overlapping duplicates for a given sub-path. E.g.
	//     DupeSubPath 'B4B4B4B' at '[82 84]'
	// AFAICT, these cases *should* be fine for correctness, albeit a bit
	// wasteful to process.
	dupes := make(map[string][]int)
	for ss, idxs := range allSubs {
		if len(idxs) > 1 {
			dupes[ss] = idxs
		}
	}
	return dupes
}

func useSubPathsInRoomRecs(pRooms map[int]*phantomRoom, rrs []roomRecord,
	pt string, size int) error {
	const minSubPathLen = 3
	log.Printf("INFO: Finding duplicate %d+ long sub-paths in the path-trace.",
		minSubPathLen)
	dupeSubPaths := findAllDupeSubPaths(pt, minSubPathLen)

	for k, v := range dupeSubPaths {
		log.Printf("INFO: DupeSubPath '%s' at '%s'.", k, fmt.Sprint(v))

		// `k` is of the form "C2C1D4A" and `v` is like "[78, 158]".
		for i := 0; i < len(k)-1; i += 2 {
			rrIdxs := make([]int, 0, len(v))
			for _, j := range v {
				rri := (i + j) / 2
				rr := rrs[rri]
				if rune(k[i]) != roomLabelToLetter(rr.label) {
					return fmt.Errorf(
						"sub-path label-mismatch: %c != %c at roomRecs[%d]",
						rune(k[i]), roomLabelToLetter(rr.label), rri)
				}
				rrIdxs = append(rrIdxs, rri)
			}

			priSet := make(map[int]bool, len(rrIdxs))
			for _, rri := range rrIdxs {
				priSet[rrs[rri].idx] = true
			}
			priList := make([]int, 0, len(priSet))
			for pri := range priSet {
				priList = append(priList, pri)
			}
			if err := mergePhantomRooms(pRooms, priList, rrs); err != nil {
				return err
			}
		}
	}
	log.Printf("INFO: %d phantom rooms remain after using dupe sub-paths.",
		len(pRooms))
	return nil
}

func getRoomDoorTgts(rm *room) string {
	var b strings.Builder
	b.Grow(4 * numDoorsPerRoom)
	for i := 0; i < numDoorsPerRoom; i++ {
		tri := rm.doors[i].toRoomIdx
		if tri == invRoomIdx {
			b.WriteString("??")
		} else {
			b.WriteString(fmt.Sprintf("%d", tri))
		}
		b.WriteRune(' ')
	}
	return b.String()
}

func getMapFromPlanResult(plan []int, res []int, size int) (*library, error) {
	var err error
	var roomRecs []roomRecord
	if roomRecs, err = getRoomRecords(plan, res); err != nil {
		return nil, err
	}

	pathTrace := tracePath(roomRecs)
	log.Printf("INFO: Results of exploration:\n%s", pathTrace)

	/* BEGIN: DEPRECATED
	if err = useLabelAndDoorInRoomRecs(lib, roomRecs); err != nil {
		return nil, err
	}
	END: DEPRECATED */

	pRooms := newPhantomRooms(roomRecs, size)
	err = useSubPathsInRoomRecs(pRooms, roomRecs, pathTrace, size)
	if err != nil {
		return nil, err
	}

	lib := newLibrary(size)

	currentRoomIdx := 0
	for _, rr := range roomRecs {
		nextRoomLabel := rr.exitLabel
		lib.rooms[currentRoomIdx].doors[rr.exitDoor].toRoomLabel = nextRoomLabel
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
		log.Printf("INFO: Room[%d]: %s", i, getRoomDoorTgts(rm))
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
			if rm.doors[i].toRoomIdx == invRoomIdx {
				return fmt.Errorf("room %d, door %d unmapped", rm.label, i)
			}
			roomFanOut[rm.label][rm.doors[i].toRoomIdx]++
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
			if srcDoor.toRoomDoor != invDoorNum {
				continue
			}
			tgtRoomIdx := -1
			for k := range lib.rooms {
				if k == srcDoor.toRoomIdx {
					tgtRoomIdx = k
					break
				}
			}
			if tgtRoomIdx < 0 {
				return fmt.Errorf("unable to locate target room %d[%d]->%d",
					i, j, srcDoor.toRoomIdx)
			}
			tgtRoom := &lib.rooms[tgtRoomIdx]
			found := false
			for l := 0; l < numDoorsPerRoom; l++ {
				tgtDoor := &tgtRoom.doors[l]
				if tgtDoor.toRoomIdx == i &&
					tgtDoor.toRoomDoor == invDoorNum {
					srcDoor.toRoomDoor = l
					tgtDoor.toRoomDoor = j
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf(
					"unable to locate entrance in target room %d[%d]->%d[?]",
					i, j, srcDoor.toRoomIdx)
			}
		}
	}
	log.Printf("INFO: Completed inter-room exit to entrance mapping.")
	return nil
}

func genGuessedMap(lib *library) (*GuessedMap, error) {
	gm := &GuessedMap{}

	gm.Rooms = make([]int, len(lib.rooms))
	for i := range lib.rooms {
		gm.Rooms[i] = lib.rooms[i].label
	}
	gm.StartingRoom = 0
	gm.Connections = make([]Connection, len(lib.rooms)*numDoorsPerRoom)
	for i := range lib.rooms {
		rm := &lib.rooms[i]
		for j := 0; j < numDoorsPerRoom; j++ {
			door := &rm.doors[j]
			conn := &gm.Connections[i*numDoorsPerRoom+j]
			conn.From.Room = i
			conn.From.Door = j
			conn.To.Room = door.toRoomIdx
			conn.To.Door = door.toRoomDoor
		}
	}

	return gm, nil
}

func (m *Mapper) Map(prob string) error {
	probInfo, err := getProblemInfo(prob)
	if err != nil {
		return err
	}
	log.Printf("INFO: Solving '%s' (size: %d, maxPath: %d, markers: %t).",
		prob, probInfo.size, probInfo.size*probInfo.maxPathFactor,
		probInfo.needsMarkers)

	if err = m.client.SelectProblem(prob); err != nil {
		return fmt.Errorf("unable to select problem: %w", err)
	}

	plan, planStr := genPlan(probInfo)
	log.Printf("INFO: Plan %s", planStr)

	var res [][]int
	res, err = m.client.Explore([]string{planStr})
	if err != nil {
		return fmt.Errorf("unable to explore: %w", err)
	}
	if len(res) != 1 {
		return fmt.Errorf("expected 1 result; got #d", len(res))
	}

	var lib *library
	lib, err = getMapFromPlanResult(plan, res[0], probInfo.size)
	if err != nil {
		return err
	}

	if err = checkMapCoverage(lib, probInfo.size); err != nil {
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
