package registry

import (
	"net/url"
	"strconv"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/common/constant"
)

// ServiceInstance is the interface  which is used for service registration and discovery.
type ServiceInstance interface {

	// GetID will return this instance's id. It should be unique.
	GetID() string

	// GetServiceName will return the serviceName
	GetServiceName() string

	// GetHost will return the hostname
	GetHost() string

	// GetPort will return the port.
	GetPort() int

	// IsEnable will return the enable status of this instance
	IsEnable() bool

	// IsHealthy will return the value represent the instance whether healthy or not
	IsHealthy() bool

	// GetMetadata will return the metadata
	GetMetadata() map[string]string

	// GetAddress will return the ip:Port
	GetAddress() string

	// GetProtocol will return the protocol
	GetProtocol() string

	ToURL() *common.URL
}

// DefaultServiceInstance the default implementation of ServiceInstance
// or change the ServiceInstance to be struct???
type DefaultServiceInstance struct {
	ID          string
	ServiceName string
	Host        string
	Port        int
	Enable      bool
	Healthy     bool

	Metadata map[string]string

	Address   string
	GroupName string
}

// GetID will return this instance's id. It should be unique.
func (d *DefaultServiceInstance) GetID() string {
	return d.ID
}

// GetServiceName will return the serviceName
func (d *DefaultServiceInstance) GetServiceName() string {
	return d.ServiceName
}

// GetHost will return the hostname
func (d *DefaultServiceInstance) GetHost() string {
	return d.Host
}

// GetPort will return the port.
func (d *DefaultServiceInstance) GetPort() int {
	return d.Port
}

// IsEnable will return the enable status of this instance
func (d *DefaultServiceInstance) IsEnable() bool {
	return d.Enable
}

// IsHealthy will return the value represent the instance whether healthy or not
func (d *DefaultServiceInstance) IsHealthy() bool {
	return d.Healthy
}

// GetAddress will return the ip:Port
func (d *DefaultServiceInstance) GetAddress() string {
	if d.Address != "" {
		return d.Address
	}
	if d.Port <= 0 {
		d.Address = d.Host
	} else {
		d.Address = d.Host + ":" + strconv.Itoa(d.Port)
	}
	return d.Address
}

// GetMetadata will return the metadata, it will never return nil
func (d *DefaultServiceInstance) GetMetadata() map[string]string {
	if d.Metadata == nil {
		d.Metadata = make(map[string]string)
	}
	return d.Metadata
}

func (d *DefaultServiceInstance) GetProtocol() string {
	if d.Metadata == nil {
		return ""
	}

	if val, ok := d.Metadata[constant.MetaProtocolKey]; ok {
		return val
	}
	return ""
}

func (d *DefaultServiceInstance) ToURL() *common.URL {
	params := url.Values{}
	for key, val := range d.Metadata {
		if len(val) != 0 {
			params.Set(key, val)
		}
	}

	instanceUrl := common.NewURLWithOptions(common.WithProtocol(d.GetProtocol()),
		common.WithParams(params),
		common.WithIp(d.Host),
		common.WithPort(strconv.Itoa(d.Port)),
		common.WithParamsValue(constant.MetaGroupKey, d.GroupName),
	)

	return instanceUrl
}
