// 排序
package gxsort

import (
	"sort"
)

func Int64(slice []int64) {
	sort.Sort(Int64Slice(slice))
}

func Int32(slice []int32) {
	sort.Sort(Int32Slice(slice))
}

func Uint32(slice []uint32) {
	sort.Sort(Uint32Slice(slice))
}

type Int64Slice []int64

func (p Int64Slice) Len() int {
	return len(p)
}

func (p Int64Slice) Less(i, j int) bool {
	return p[i] < p[j]
}

func (p Int64Slice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

type Int32Slice []int32

func (p Int32Slice) Len() int {
	return len(p)
}

func (p Int32Slice) Less(i, j int) bool {
	return p[i] < p[j]
}

func (p Int32Slice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

type Uint32Slice []uint32

func (p Uint32Slice) Len() int {
	return len(p)
}

func (p Uint32Slice) Less(i, j int) bool {
	return p[i] < p[j]
}

func (p Uint32Slice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
