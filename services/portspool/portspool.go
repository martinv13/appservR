package portspool

import (
	"errors"
	"net"
	"strconv"
)

var inUse = make(map[string]bool)
var rangeStart = 4000

func Init(startPort int) {
	rangeStart = startPort
}

func GetNext() (string, error) {
	for port := rangeStart; port < rangeStart+1000; port++ {
		strPort := strconv.Itoa(port)
		_, ok := inUse[strPort]
		if !ok {
			ln, err := net.Listen("tcp", ":"+strPort)
			if err == nil {
				ln.Close()
				inUse[strPort] = true
				return strPort, nil
			}
		}
	}
	return "", errors.New("can't find an available port")
}

func Release(port string) {
	delete(inUse, port)
}
