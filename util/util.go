package util

import (
	"net"
)

func GetIpv4Addrs() (list []string) {
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if inet, ok := addr.(*net.IPNet); ok && !inet.IP.IsLoopback() {
				if inet.IP.To4() != nil {
					list = append(list, inet.IP.String())
				}
			}
		}
	}
	return
}

func ResolveDomain(domain string) string {
	addrs, err := net.LookupHost(domain)
	if err != nil {
		return ""
	}
	return addrs[0]
}
