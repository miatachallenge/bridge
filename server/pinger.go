package main

import (
	"net"
	"time"

	"github.com/pkg/errors"
	fastping "github.com/tatsushid/go-fastping"
)

type pingResponse struct {
	addr *net.IPAddr
	rtt  time.Duration
}

func startPinger() error {
	pinger := fastping.NewPinger()

	resolvedIP, err := net.ResolveIPAddr("ip4:icmp", *modem)
	if err != nil {
		return errors.Wrap(err, "unable to resolve the ip addr")
	}

	pinger.AddIPAddr(resolvedIP)

	onRecv, onIdle := make(chan *pingResponse), make(chan bool)
	pinger.OnRecv = func(addr *net.IPAddr, t time.Duration) {
		onRecv <- &pingResponse{addr: addr, rtt: t}
	}
	pinger.OnIdle = func() {
		onIdle <- true
	}

	pinger.RunLoop()

	go func() {
		var result *pingResponse
		for {
			select {
			case res := <-onRecv:
				if res.addr.String() == resolvedIP.String() {
					result = res
				}
			case <-onIdle:
				stateLock.Lock()
				if result == nil {
					state.Modem.Ping = -1
				} else {
					state.Modem.Ping = int64(result.rtt)
				}
				stateLock.Unlock()

				result = nil
			}
		}
	}()

	return nil
}
