package sched

import (
	"sort"
	"strconv"

	"github.com/sdsph/dialysis-scheduling/internal/model"
)

func itoa(v int64) string { return strconv.FormatInt(v, 10) }

func sortShifts(in []*model.Shift) []*model.Shift {
	out := make([]*model.Shift, len(in))
	copy(out, in)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Sort < out[j].Sort })
	return out
}

func sortMachines(in []*model.Machine) {
	sort.SliceStable(in, func(i, j int) bool { return in[i].PositionIndex < in[j].PositionIndex })
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func derefI64(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}
