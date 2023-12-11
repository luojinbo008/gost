package nacos

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/common/constant"
	nacosClient "github.com/luojinbo008/gost/utils/database/kv/nacos"
	nacosConstant "github.com/nacos-group/nacos-sdk-go/v2/common/constant"

	perrors "github.com/pkg/errors"
)

// GetNacosConfig will return the nacos config
func GetNacosConfig(url *common.URL) ([]nacosConstant.ServerConfig, nacosConstant.ClientConfig, error) {
	if url == nil {
		return []nacosConstant.ServerConfig{}, nacosConstant.ClientConfig{}, perrors.New("url is empty!")
	}

	if len(url.Location) == 0 {
		return []nacosConstant.ServerConfig{}, nacosConstant.ClientConfig{},
			perrors.New("url.location is empty!")
	}

	addresses := strings.Split(url.Location, ",")
	serverConfigs := make([]nacosConstant.ServerConfig, 0, len(addresses))

	for _, addr := range addresses {
		ip, portStr, err := net.SplitHostPort(addr)
		if err != nil {
			return []nacosConstant.ServerConfig{}, nacosConstant.ClientConfig{},
				perrors.WithMessagef(err, "split [%s] ", addr)
		}
		portContextPath := strings.Split(portStr, constant.PathSeparator)
		port, err := strconv.Atoi(portContextPath[0])
		if err != nil {
			return []nacosConstant.ServerConfig{}, nacosConstant.ClientConfig{},
				perrors.WithMessagef(err, "port [%s] ", portContextPath[0])
		}
		var contextPath string
		if len(portContextPath) > 1 {
			contextPath = constant.PathSeparator + strings.Join(portContextPath[1:], constant.PathSeparator)
		}
		serverConfigs = append(serverConfigs, nacosConstant.ServerConfig{IpAddr: ip, Port: uint64(port), ContextPath: contextPath})
	}

	timeout := url.GetParamDuration(constant.NacosTimeout, constant.DefaultRegTimeout)

	clientConfig := nacosConstant.ClientConfig{
		TimeoutMs:            uint64(int32(timeout / time.Millisecond)),
		NamespaceId:          url.GetParam(constant.NacosNamespaceID, ""),
		Username:             url.GetParam(constant.NacosUsername, ""),
		Password:             url.GetParam(constant.NacosPassword, ""),
		BeatInterval:         url.GetParamInt(constant.NacosBeatIntervalKey, 5000),
		AppName:              url.GetParam(constant.NacosAppNameKey, ""),
		Endpoint:             url.GetParam(constant.NacosEndpoint, ""),
		RegionId:             url.GetParam(constant.NacosRegionIDKey, ""),
		AccessKey:            url.GetParam(constant.NacosAccessKey, ""),
		SecretKey:            url.GetParam(constant.NacosSecretKey, ""),
		OpenKMS:              url.GetParamBool(constant.NacosOpenKmsKey, false),
		CacheDir:             url.GetParam(constant.NacosCacheDirKey, ""),
		UpdateThreadNum:      url.GetParamByIntValue(constant.NacosUpdateThreadNumKey, 20),
		NotLoadCacheAtStart:  url.GetParamBool(constant.NacosNotLoadLocalCache, true),
		LogDir:               url.GetParam(constant.NacosLogDirKey, ""),
		LogLevel:             url.GetParam(constant.NacosLogLevelKey, "info"),
		UpdateCacheWhenEmpty: url.GetParamBool(constant.NacosUpdateCacheWhenEmpty, false),
	}
	return serverConfigs, clientConfig, nil
}

// NewNacosClientByURL created
func NewNacosClientByURL(url *common.URL) (*nacosClient.NacosNamingClient, error) {
	scs, cc, err := GetNacosConfig(url)

	if err != nil {
		return nil, err
	}
	clientName := url.GetParam(constant.ClientNameKey, "")
	if len(clientName) <= 0 {
		return nil, perrors.New("nacos client name must set")
	}
	return nacosClient.NewNacosNamingClient(clientName, true, scs, cc)
}
