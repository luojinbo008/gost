package failfast

import (
	"github.com/luojinbo008/gost/common/constant"
	"github.com/luojinbo008/gost/common/extension"
	clusterpkg "github.com/luojinbo008/gost/internal/cluster/cluster"
	"github.com/luojinbo008/gost/internal/cluster/directory"
	"github.com/luojinbo008/gost/internal/protocol"
)

func init() {
	extension.SetCluster(constant.ClusterKeyFailfast, newFailfastCluster)
}

type failfastCluster struct{}

// newFailfastCluster returns a failfastCluster instance.
//
// Fast failure, only made a call, failure immediately error. Usually used for non-idempotent write operations,
// such as adding records.
func newFailfastCluster() clusterpkg.Cluster {
	return &failfastCluster{}
}

// Join returns a baseClusterInvoker instance
func (cluster *failfastCluster) Join(directory directory.Directory) protocol.Invoker {
	return clusterpkg.BuildInterceptorChain(newFailfastClusterInvoker(directory))
}
