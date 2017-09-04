package main

import (
	//	"fmt"
	"github.com/op/go-logging"
	"github.com/urfave/cli"
	"os"
)

var version string
var log = logging.MustGetLogger("main")

// note the "#" -  it is added so exabgp treats the line as comment
var stdout_log_format = logging.MustStringFormatter("# %{color:bold}%{time:2006-01-02T15:04:05.0000Z-07:00}%{color:reset}%{color} [%{level:.1s}] %{color:reset}%{shortpkg}[%{longfunc}] %{message}")

func main() {
	stderrBackend := logging.NewLogBackend(os.Stderr, "", 0)
	stderrFormatter := logging.NewBackendFormatter(stderrBackend, stdout_log_format)
	logging.SetBackend(stderrFormatter)
	logging.SetFormatter(stdout_log_format)

	app := cli.NewApp()
	app.Name = "HA2BGP"
	app.Usage = "Announce BGP prefixes only when services are up"
	app.Description = "HA2BGP is bridge between bgp (so far only ExaBGP) and various healthchecks (so far only checks for listening socket\n    It's role is to announce routes when service is up and withdraw when they are down or flapping"
	app.Version = version
	app.HideHelp = true
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "help, h", Usage: "show help"},
		cli.StringFlag{
			Name:   "socket, s",
			Value:  "/run/haproxy/admin.sock",
			Usage:  "HAProxy socket path",
			EnvVar: "HA2BGP_HAPROXY_SOCKET",
		},
		cli.StringSliceFlag{
			Name:   "network, n",
			Usage:  "Networks allowed to be distributed, in CIDR format. Defaults to 127.0.0.1/8",
			EnvVar: "HA2BGP_NETWORK_FILTER",
		},
		cli.StringFlag{
			Name:   "listen-filter, l",
			Value:  "sport = :80 or sport = :443",
			Usage:  "ss-compatible filter. defaults to only looking for listening ports 80 and 443",
			EnvVar: "HA2BGP_LISTEN_FILTER",
		},
		// cli.StringSliceFlag{
		// 	Name:   "announce, a",
		// 	Usage:  "List of IPs to announce. This WILL ignore network filter but will not ignore backend state.",
		// 	EnvVar: "HA2BGP_ANNOUNCE",
		// },
		cli.StringFlag{
			Name:   "nexthop, t",
			Value:  "self",
			Usage:  "Next hop, defaults to self",
			EnvVar: "HA2BGP_NEXTHOP",
		},
	}
	app.Action = func(c *cli.Context) error {
		if c.Bool("help") {
			cli.ShowAppHelp(c)
			os.Exit(1)
		}
		log.Infof("Starting HA2BGP version: %s", version)
		MainLoop(c)
		return nil
	}
	app.Run(os.Args)

}
