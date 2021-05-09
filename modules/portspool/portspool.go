package portspool

import (
	"errors"
	"net"
	"strconv"
	"sync"
)

var portsPool = struct {
	sync.Mutex
	inUse      map[string]bool
	rangeStart int
}{
	inUse:      make(map[string]bool),
	rangeStart: 4000,
}

func GetNext() (string, error) {
	portsPool.Lock()
	defer portsPool.Unlock()
	for port := portsPool.rangeStart; port < portsPool.rangeStart+1000; port++ {
		strPort := strconv.Itoa(port)
		_, ok := portsPool.inUse[strPort]
		if !ok {
			ln, err := net.Listen("tcp", ":"+strPort)
			portsPool.inUse[strPort] = true
			if err == nil {
				ln.Close()
				return strPort, nil
			}
		}
	}
	return "", errors.New("can't find an available port")
}

func Release(port string) {
	portsPool.Lock()
	defer portsPool.Unlock()
	delete(portsPool.inUse, port)
}
