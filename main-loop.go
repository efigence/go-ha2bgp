package main

import (
	"github.com/efigence/go-ha2bgp/check/listen"
	"github.com/efigence/go-ha2bgp/exabgp"
	"github.com/efigence/go-haproxy"
	"github.com/urfave/cli"
	"os"
	"time"
)

func MainLoop(c *cli.Context) {
	bgp, err := exabgp.NewExaBGP3()
	if err != nil {
		log.Errorf("ExaBGP error, exiting: %s", err)
		return
	}
	check, err := listen.NewCheck(`tcp`, "")
	log.Infof("listen check: %s", err)
	for {
		log.Noticef("check state: %+v", check.DebugListenState())
		if WaitForHaproxySocketOk(c.String(`socket`)) {
			sock := haproxy.New(c.String(`socket`))
			errCnt := 0
			for {
				if _, err := sock.RunCmd(`quit`); err == nil {
					log.Info("HAProxy OK, announcing routes")
					bgp.AnnounceRouteSlice(c.StringSlice(`announce`), `self`)
					check.Check()
					log.Noticef("check state: %+v", check.DebugListenState())
				} else {
					log.Errorf("Error when communicating with haproxy, withdrawing routes: %s", err)
					for _, v := range c.StringSlice(`announce`) {
						bgp.WithdrawRoute(v, `self`)
						errCnt = errCnt + 1
					}
				}
				time.Sleep(time.Second * 10)
				if errCnt > 10 {
					break
				}
			}
		} else {
			log.Errorf("No HAProxy communication, withdrawing routes")
			bgp.WithdrawRouteSlice(c.StringSlice(`announce`), `self`)
		}
	}

}

func Announce(bgp *exabgp.Exa3, s []string) {
	for _, v := range s {
		bgp.AnnounceRoute(v, `self`)
	}
}

func Withdraw(bgp *exabgp.Exa3, s []string) {
	for _, v := range s {
		bgp.WithdrawRoute(v, `self`)
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
