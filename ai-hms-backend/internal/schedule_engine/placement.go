package schedule_engine

import "math"

// PlaceHdfSession HDF次落位: 双固定→就近→报警(规范 §5)
func PlaceHdfSession(beds []BedInfo, occupied map[int64][]Occupancy, wardID int64, cell Cell, session SessionItem) *BedInfo {
	// 1) 双固定: 回到固定HDF机
	if session.FixedHdfBedID != nil {
		for _, b := range beds {
			if b.ID == *session.FixedHdfBedID && b.WardID == wardID &&
				MachineSupports(b.MachineType, ModeHDF) && !b.IsDisabled &&
				!IsOccupied(occupied, b.ID, cell) {
				return &b
			}
		}
	}
	// 2) 就近: 同区空闲HDF机
	hdfBeds := FindFreeBeds(beds, occupied, wardID, cell, ModeHDF)
	if len(hdfBeds) == 0 {
		return nil // 3) 报警
	}
	return nearest(hdfBeds, session.FixedHdfBedID, session.FixedHdBedID)
}

// PlaceHdSession HD次落位: 固定HD机→同区就近HD机→溢出HDF机(规范 §4.2/§4.1)
func PlaceHdSession(beds []BedInfo, occupied map[int64][]Occupancy, wardID int64, cell Cell, session SessionItem) *BedInfo {
	// 1) 固定HD机位
	if session.FixedHdBedID != nil {
		for _, b := range beds {
			if b.ID == *session.FixedHdBedID && b.WardID == wardID &&
				MachineSupports(b.MachineType, ModeHD) && !b.IsDisabled &&
				!IsOccupied(occupied, b.ID, cell) {
				return &b
			}
		}
	}
	// 2) 同区空闲HD机
	hdBeds := FindFreeBeds(beds, occupied, wardID, cell, ModeHD)
	hdOnly := make([]BedInfo, 0)
	for _, b := range hdBeds {
		if b.MachineType == MachineHD {
			hdOnly = append(hdOnly, b)
		}
	}
	if len(hdOnly) > 0 {
		return pickBest(hdOnly, session)
	}
	// 3) HD机满→溢出HDF机
	hdfFree := FindFreeBeds(beds, occupied, wardID, cell, ModeHD)
	hdfOnly := make([]BedInfo, 0)
	for _, b := range hdfFree {
		if b.MachineType == MachineHDF {
			hdfOnly = append(hdfOnly, b)
		}
	}
	if len(hdfOnly) > 0 {
		return pickBest(hdfOnly, session)
	}
	return nil
}

// pickBest 候选机位评分: 集中连片+组团, 不做负载均衡(规范 §4.4)
func pickBest(candidates []BedInfo, session SessionItem) *BedInfo {
	if len(candidates) == 0 {
		return nil
	}
	// 按PositionIndex排序取最优
	best := &candidates[0]
	bestDist := math.MaxInt32
	refPos := -1
	if session.FixedHdBedID != nil {
		refPos = findBedPosition(candidates, *session.FixedHdBedID)
	}
	if refPos < 0 && session.FixedHdfBedID != nil {
		refPos = findBedPosition(candidates, *session.FixedHdfBedID)
	}
	for i := range candidates {
		dist := abs(candidates[i].PositionIndex - refPos)
		if refPos < 0 {
			dist = candidates[i].PositionIndex
		}
		if dist < bestDist {
			bestDist = dist
			best = &candidates[i]
		}
	}
	return best
}

// nearest 取离参考机位最近的候选(规范 §5 就近)
func nearest(candidates []BedInfo, refID1, refID2 *int64) *BedInfo {
	refID := refID1
	if refID == nil {
		refID = refID2
	}
	if refID == nil || len(candidates) == 0 {
		return &candidates[0]
	}
	refPos := -1
	for _, b := range candidates {
		if b.ID == *refID {
			refPos = b.PositionIndex
			break
		}
	}
	best := &candidates[0]
	bestDist := math.MaxInt32
	for i := range candidates {
		dist := abs(candidates[i].PositionIndex - refPos)
		if refPos < 0 {
			dist = candidates[i].PositionIndex
		}
		if dist < bestDist {
			bestDist = dist
			best = &candidates[i]
		}
	}
	return best
}

func findBedPosition(beds []BedInfo, id int64) int {
	for _, b := range beds {
		if b.ID == id {
			return b.PositionIndex
		}
	}
	return -1
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
