// Copyright 2016 HcNet Development Foundation and contributors. Licensed
// under the Apache License, Version 2.0. See the COPYING file at the root
// of this distribution or at http://www.apache.org/licenses/LICENSE-2.0

package historyarchive

import (
	"fmt"
	"sort"
	"strings"
)

const CheckpointFreq = uint32(64)

type Range struct {
	Low  uint32
	High uint32
}

func PrevCheckpoint(i uint32) uint32 {
	freq := CheckpointFreq
	if i < freq {
		return freq - 1
	}
	return ((i / freq) * freq) - 1
}

func NextCheckpoint(i uint32) uint32 {
	if i == 0 {
		return CheckpointFreq - 1
	}
	freq := uint64(CheckpointFreq)
	v := uint64(i)
	n := (((v + freq - 1) / freq) * freq) - 1
	if n >= 0xffffffff {
		return 0xffffffff
	}
	return uint32(n)
}

func MakeRange(low uint32, high uint32) Range {
	if high < low {
		high = low
	}
	return Range{
		Low:  PrevCheckpoint(low),
		High: NextCheckpoint(high),
	}
}

func (r Range) clamp(other Range) Range {
	low := r.Low
	high := r.High
	if low < other.Low {
		low = other.Low
	}
	if high > other.High {
		high = other.High
	}
	return MakeRange(low, high)
}

func (r Range) String() string {
	return fmt.Sprintf("[0x%8.8x, 0x%8.8x]", r.Low, r.High)
}

func (r Range) Checkpoints() chan uint32 {
	ch := make(chan uint32)
	go func() {
		for i := uint64(r.Low); i < uint64(r.High); i += uint64(CheckpointFreq) {
			ch <- uint32(i)
		}
		close(ch)
	}()
	return ch
}

func (r Range) Size() int {
	return int(r.High-r.Low) / int(CheckpointFreq)
}

func (r Range) collapsedString() string {
	if r.Low == r.High {
		return fmt.Sprintf("0x%8.8x", r.Low)
	} else {
		return fmt.Sprintf("[0x%8.8x-0x%8.8x]", r.Low, r.High)
	}
}

type byUint32 []uint32

func (a byUint32) Len() int           { return len(a) }
func (a byUint32) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byUint32) Less(i, j int) bool { return a[i] < a[j] }

func fmtRangeList(vs []uint32) string {

	sort.Sort(byUint32(vs))

	s := make([]string, 0, 10)
	var curr *Range

	for _, t := range vs {
		if curr != nil {
			if curr.High+CheckpointFreq == t {
				curr.High = t
				continue
			} else {
				s = append(s, curr.collapsedString())
				curr = nil
			}
		}
		curr = &Range{Low: t, High: t}
	}
	if curr != nil {
		s = append(s, curr.collapsedString())
	}

	return strings.Join(s, ", ")
}
