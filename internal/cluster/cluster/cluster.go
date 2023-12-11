package cluster

import (
	"github.com/luojinbo008/gost/internal/cluster/directory"
	"github.com/luojinbo008/gost/internal/protocol"
)

type Cluster interface {
	Join(directory.Directory) protocol.Invoker
}
