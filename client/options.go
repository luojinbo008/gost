package client

import (
	"strconv"
	"time"

	"github.com/creasty/defaults"
	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/config"
	"github.com/luojinbo008/gost/internal/protocol"
)

type ReferOptions struct {
	id      string
	invoker protocol.Invoker
	urls    []*common.URL
}

func defaultReferptions() *ReferOptions {
	return &ReferOptions{}
}

func (refOpts *ReferOptions) init(cli *Client, opts ...ReferOption) error {
	for _, opt := range opts {
		opt(refOpts)
	}
	return nil
}

type ReferOption func(*ReferOptions)

func WithReferUrl(url *common.URL) ReferOption {
	return func(opts *ReferOptions) {
		opts.urls = append(opts.urls, url)
	}
}

type ClientOptions struct {
	// Consumer    *global.ConsumerConfig
	Application *config.ApplicationConfig
	Registries  map[string]*config.RegistryConfig
	// Shutdown    *global.ShutdownConfig
	// Metrics     *global.MetricsConfig
	// Otel        *global.OtelConfig
}

func defaultClientOptions() *ClientOptions {
	return &ClientOptions{
		// Consumer:    global.DefaultConsumerConfig(),
		Registries:  make(map[string]*config.RegistryConfig),
		Application: config.GetApplicationConfig(),
		// Shutdown:    global.DefaultShutdownConfig(),
		// Metrics:     global.DefaultMetricsConfig(),
		// Otel:        global.DefaultOtelConfig(),
	}
}

func (cliOpts *ClientOptions) init(opts ...ClientOption) error {
	for _, opt := range opts {
		opt(cliOpts)
	}
	if err := defaults.Set(cliOpts); err != nil {
		return err
	}
	return nil
}

type ClientOption func(*ClientOptions)

func SetApplication(application *config.ApplicationConfig) ClientOption {
	return func(opts *ClientOptions) {
		opts.Application = application
	}
}

func SetRegistries(regs map[string]*config.RegistryConfig) ClientOption {
	return func(opts *ClientOptions) {
		opts.Registries = regs
	}
}

// todo: need to be consistent with MethodConfig
type CallOptions struct {
	RequestTimeout string
	Retries        string
	Group          string
	Version        string
}

type CallOption func(*CallOptions)

func newDefaultCallOptions() *CallOptions {
	return &CallOptions{}
}

// WithCallRequestTimeout the maximum waiting time for one specific call, only works for 'tri' and 'dubbo' protocol
func WithCallRequestTimeout(timeout time.Duration) CallOption {
	return func(opts *CallOptions) {
		opts.RequestTimeout = timeout.String()
	}
}

// WithCallRetries the maximum retry times on request failure for one specific call, only works for 'tri' and 'dubbo' protocol
func WithCallRetries(retries int) CallOption {
	return func(opts *CallOptions) {
		opts.Retries = strconv.Itoa(retries)
	}
}

func WithCallGroup(group string) CallOption {
	return func(opts *CallOptions) {
		opts.Group = group
	}
}

func WithCallVersion(version string) CallOption {
	return func(opts *CallOptions) {
		opts.Version = version
	}
}
