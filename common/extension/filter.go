package extension

import (
	"github.com/luojinbo008/gost/internal/filter"
)

var (
	filters = make(map[string]func() filter.Filter)
)

func SetFilter(name string, v func() filter.Filter) {
	filters[name] = v
}

func GetFilter(name string) (filter.Filter, bool) {
	if filters[name] == nil {
		return nil, false
	}
	return filters[name](), true
}
