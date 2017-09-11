package main

import (
	"github.com/efigence/go-ha2bgp/check"
	"github.com/efigence/go-ha2bgp/check/ifup"
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
		log.Notice("please specify whitelisted networks using --network parameter; running in test mode where only localhost(127.0.0.0/8) IPs will be recognized")
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
	// TODO redo that part, it is a bit messy
	log.Debugf("Full network filter: %+v", networkFilter)
	check, err := check.NewCheck()
	checkListen, err := listen.NewCheck(`tcp`, listenFilter)
	if err != nil {
		log.Panicf("Can't initialize listener checker: %s", err)
	}
	checkIfup, err := ifup.NewCheck(c.String(`device`), c.String(`device-label`))
	if err != nil {
		log.Panicf("Can't initialize ifup checker: %s", err)
	}
	check.Register(`listen`, checkListen)
	check.Register(`ifup`, checkIfup)
	// add any new listening IP to the list
	// it will be checked before announcing so in case of IP that is listening but not up (net.ipv4.ip_nonlocal_bind = 1)
	checkListen.SetNewIpHook(func(ip net.IP) {
		for _, n := range networkFilter {
			if n.Contains(ip) {
				log.Noticef("New listening socket IP added: %s,%s", ip.String(), n.String())
				core.RegisterRoute(ip.String(), nexthop, "", check)
				break
			}
		}

	})
	// test run if it runs correctly
	err = check.Check()
	if err != nil {
		log.Panicf("Error running checker: %s", err)
	}
	go func() {
		checkMin := time.Duration(time.Second)
		checkMax := time.Duration(time.Second * 10)
		log.Noticef("Running checks every 1..10s (1s by default unless slowdown is detected)")
		for {
			start := time.Now()
			err := check.Check()
			if err != nil {
				log.Errorf("Error running check: %s", err)
			}
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
