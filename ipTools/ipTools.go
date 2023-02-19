package ipTools

import (
	"net"
	"strings"
)

func DigitBitIsOne(digit uint64, index uint64) bool {
	return digit&(1<<index) == 1
}

func InterIsUp(flags uint) bool {
	return DigitBitIsOne(uint64(flags), 0)
}

// LocalIP 获取本机IP
func LocalIP() string {
	inters, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, inte := range inters {
		if InterIsUp(uint(inte.Flags)) {
			addrs, _ := inte.Addrs()
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipV4 := ipnet.IP.To4(); ipV4 != nil {
						ipaddrs := strings.Split(ipV4.String(), "/")[0]
						if !strings.HasSuffix(ipaddrs, ".1") && !strings.HasSuffix(ipaddrs, ".255") {
							return ipaddrs
						}
					}
				}
			}
		}
	}

	return ""
}
