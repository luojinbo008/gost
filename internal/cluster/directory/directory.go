package directory

import (
	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/internal/protocol"
)

// Directory
// Extension - Directory
type Directory interface {
	common.Node

	// List candidate invoker list for the current Directory.
	// NOTICE: The invoker list returned to the caller may be backed by the same data hold by the current Directory
	// implementation for the sake of performance consideration. This requires the caller of List() shouldn't modify
	// the return result directly.
	List(invocation protocol.Invocation) []protocol.Invoker
}
