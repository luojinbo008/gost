package constant

import "math"

const (
	ServiceDiscoveryDefaultGroup = "DEFAULT_GROUP"

	GOSTIpToRegistryKey       = "GOST_IP_TO_REGISTRY"
	GOSTPortToRegistryKey     = "GOST_PORT_TO_REGISTRY"
	GOSTDefaultPortToRegistry = "80"
)

const (
	DefaultRegTimeout = "5s"
)

const (
	DefaultMaxServerRecvMsgSize = 1024 * 1024 * 4
	DefaultMaxServerSendMsgSize = math.MaxInt32

	DefaultMaxCallRecvMsgSize = 1024 * 1024 * 4
	DefaultMaxCallSendMsgSize = math.MaxInt32
)

const (
	ServiceFilterKey   = "service.filter"
	ReferenceFilterKey = "reference.filter"
)

const (
	LoadBalanceKeyRoundRobin = "roundrobin"
	DefaultLoadBalance       = LoadBalanceKeyRoundRobin
)
