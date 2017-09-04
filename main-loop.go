package main

import (
	"github.com/efigence/go-ha2bgp/check/listen"
	"github.com/efigence/go-ha2bgp/engine"
	"github.com/efigence/go-ha2bgp/exabgp"
	"github.com/efigence/go-haproxy"
	"github.com/urfave/cli"
	"net"
	"os"
	"time"
)

func MainLoop(c *cli.Context) {
	bgp, err := exabgp.NewExaBGP3()
	if err != nil {
		log.Errorf("ExaBGP error, exiting: %s", err)
		return
	}
	core, err := engine.NewEngine(bgp)
	if err != nil {
		log.Panicf("Error when starting engine: %s", err)
	}
	nexthop := c.String("nexthop")
	listenFilter := c.String("listen-filter")
	log.Noticef("ss filter: %s", listenFilter)
	// prepare list of network filters
	rawNets := c.StringSlice("network")
	if len(rawNets) == 0 {
		rawNets = []string{"127.0.0.1/8"}
	}
	networkFilter := make([]net.IPNet, len(rawNets))
	for id, rawNet := range rawNets {
		_, ipNet, err := net.ParseCIDR(rawNet)
		if err != nil || ipNet == nil {
			log.Panicf("can't parse network: %s", rawNet)
		}
		log.Noticef("adding network %s to filter", ipNet.String())
		networkFilter[id] = *ipNet

	}
	log.Errorf("filter: %+v", networkFilter)
	check, err := listen.NewCheck(`tcp`, listenFilter)
	check.SetNewIpHook(func(ip net.IP) {
		for _, n := range networkFilter {
			if n.Contains(ip) {
				log.Warningf("New IP added: %+v,%+v", ip, n)
				core.RegisterRoute(ip.String(), nexthop, "", check)
			}
		}

	})
	go func() {
		checkMin := time.Duration(time.Second)
		checkMax := time.Duration(time.Second * 10)
		log.Noticef("Running listen checks every 1..10s (1s by default unless slowdown is detected)")
		for {
			start := time.Now()
			check.Check()
			diff := time.Since(start)
			delay := diff * 10
			switch {
			case delay < checkMin:
				time.Sleep(checkMin)
			case delay > checkMax:
				time.Sleep(checkMax)
			default:
				time.Sleep(delay)
			}
		}
	}()
	log.Noticef("Running core state update every second")
	for {
		time.Sleep(time.Second * 1)
		core.UpdateState()
	}
}

func WaitForHaproxySocketOk(socketPath string) bool {
	for _, d := range []int{1, 2, 4, 8, 10} {
		if _, err := os.Stat(socketPath); os.IsNotExist(err) {
			log.Errorf("Can't open HAProxy socket [%s]: %s", socketPath, err)
		} else {
			sock := haproxy.New(socketPath)
			if _, err := sock.RunCmd(`quit`); err == nil {
				return true
			} else {
				log.Errorf("Can't send command to HAProxy socket [%s]: %s", socketPath, err)
			}
		}
		time.Sleep(time.Duration(d) * time.Second)
	}
	return false
}
