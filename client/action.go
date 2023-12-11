package client

import (
	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/common/constant"
	"github.com/luojinbo008/gost/common/extension"
	"github.com/luojinbo008/gost/internal/cluster/directory/static"
	"github.com/luojinbo008/gost/internal/protocol"
)

// ReferWithService retrieves invokers from urls.
func (refOpts *ReferOptions) ReferWithService(srv common.RPCService) {
	refOpts.refer(srv, nil)
}

func (refOpts *ReferOptions) ReferWithInfo(info *ClientInfo) {
	refOpts.refer(nil, info)
}

func (refOpts *ReferOptions) refer(srv common.RPCService, info *ClientInfo) {

	if info != nil {
		refOpts.id = info.InterfaceName
	} else {
		refOpts.id = common.GetReference(srv)
	}

	var (
		invoker protocol.Invoker
	)

	invokers := make([]protocol.Invoker, len(refOpts.urls))
	for i, u := range refOpts.urls {
		invoker = extension.GetProtocol(u.Protocol).Refer(u)
		invokers[i] = invoker
	}

	hitClu := constant.ClusterKeyFailfast
	cluster, err := extension.GetCluster(hitClu)
	if err != nil {
		panic(err)
	} else {
		refOpts.invoker = cluster.Join(static.NewDirectory(invokers))
	}

}
