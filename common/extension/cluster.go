package extension

import (
	"fmt"

	"github.com/luojinbo008/gost/internal/cluster/cluster"
	"github.com/pkg/errors"
)

var clusters = make(map[string]func() cluster.Cluster)

// SetCluster sets the cluster fault-tolerant mode with @name
// For example: failfast
func SetCluster(name string, fcn func() cluster.Cluster) {
	clusters[name] = fcn
}

// GetCluster finds the cluster fault-tolerant mode with @name
func GetCluster(name string) (cluster.Cluster, error) {
	if clusters[name] == nil {
		return nil, errors.New(fmt.Sprintf("Cluster for %s is not existing, make sure you have import the package.", name))
	}
	return clusters[name](), nil
}
