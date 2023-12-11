package constant

type GOSTCtxKey string

// golbal
const (
	GOST = "gost"

	// 超时
	TimeoutKey = "timeout"

	PathSeparator = "/"
	DotSeparator  = "."

	TimestampKey = "timestamp"
	ClusterKey   = "cluster"
	WeightKey    = "weight"

	// remote
	ClientNameKey = "remote-client-name"
)

// app key
const (
	EnvironmentKey = "environment"
	GroupKey       = "group"
	VersionKey     = "version"
)

// service key
const (
	ServiceKey = "service"

	InterfaceKey = "interface"

	MaxCallSendMsgSize   = "max-call-send-msg-size"
	MaxServerSendMsgSize = "max-server-send-msg-size"
	MaxCallRecvMsgSize   = "max-call-recv-msg-size"
	MaxServerRecvMsgSize = "max-server-recv-msg-size"
)

// registry keys
const (
	RegistryKey          = "registry"
	RegistryAccessKey    = "registry.accesskey"
	RegistrySecretKey    = "registry.secretkey"
	RegistryTimeoutKey   = "registry.timeout"
	RegistryNamespaceKey = "registry.namespace"
	RegistryGroupKey     = "registry.group"
	RegistryProtocol     = "registry.protocol"
	RegistryTTLKey       = "registry.ttl"
	RegistryLabelKey     = "label"
	RegistryZoneKey      = "zone"
)

const (
	Tagkey = "gost.tag" // key of tag

	AttachmentKey = GOSTCtxKey("attachment") // key in context in invoker
)

const (
	AsyncKey = "async" // it's value should be "true" or "false" of string type
)

// router
const (
	TagRouterFactoryKey = "tag"
)

const (
	LoadbalanceKey     = "loadbalance"
	ClusterKeyFailfast = "failfast"
)
