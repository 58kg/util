package util

import (
	"errors"
	"math/rand"
	"net"
	"regexp"
	"time"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

// RetryDo Retry sn表示这是第几次执行f, 至多执行times次f
func Retry(f func(sn int) (end bool), times int, maxSleepTime time.Duration) {
	for i := 1; i <= times; i++ {
		if f(i) || i == times {
			return
		}
		time.Sleep(time.Duration(random.Int63n(int64(maxSleepTime))))
	}
}

var trimWhiteReg = regexp.MustCompile(`(^\s+)|(\s+$)`)

func TrimWhite(s string) string {
	return trimWhiteReg.ReplaceAllString(s, "")
}

func ExternalIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			ip := getIpFromAddr(addr)
			if ip == nil {
				continue
			}
			return ip, nil
		}
	}
	return nil, errors.New("may be not connect to the network")
}

func getIpFromAddr(addr net.Addr) net.IP {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil || ip.IsLoopback() {
		return nil
	}
	ip = ip.To4()
	if ip == nil {
		return nil
	}
	return ip
}
