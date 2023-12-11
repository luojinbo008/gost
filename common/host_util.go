package common

import (
	"os"
	"strconv"

	"github.com/luojinbo008/gost/common/constant"
	gxnet "github.com/luojinbo008/gost/utils/net"
)

var localIp string

func GetLocalIp() string {
	if len(localIp) != 0 {
		return localIp
	}
	localIp, _ = gxnet.GetLocalIP()
	return localIp
}

func HandleRegisterIPAndPort(url *URL) {
	// if developer define registry port and ip, use it first.
	if ipToRegistry := os.Getenv(constant.GOSTIpToRegistryKey); len(ipToRegistry) > 0 {
		url.Ip = ipToRegistry
	}
	if len(url.Ip) == 0 {
		url.Ip = GetLocalIp()
	}
	if portToRegistry := os.Getenv(constant.GOSTPortToRegistryKey); isValidPort(portToRegistry) {
		url.Port = portToRegistry
	}
	if len(url.Port) == 0 || url.Port == "0" {
		url.Port = constant.GOSTDefaultPortToRegistry
	}
}

func isValidPort(port string) bool {
	if len(port) == 0 {
		return false
	}

	portInt, err := strconv.Atoi(port)
	return err == nil && portInt > 0 && portInt < 65536
}
