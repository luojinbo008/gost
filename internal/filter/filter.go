package filter

import (
	"context"

	"github.com/luojinbo008/gost/internal/protocol"
)

// Filter is the interface which wraps Invoke and OnResponse method and defines the functions of a filter.
// Invoke method is the core function of a filter, it determines the process of the filter.
// OnResponse method updates the results from Invoke and then returns the modified results.
type Filter interface {
	Invoke(context.Context, protocol.Invoker, protocol.Invocation) protocol.Result
	OnResponse(context.Context, protocol.Result, protocol.Invoker, protocol.Invocation) protocol.Result
}
