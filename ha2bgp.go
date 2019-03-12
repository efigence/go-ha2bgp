package main

import (
	//	"fmt"
	"github.com/op/go-logging"
	"github.com/urfave/cli"
	"os"
	"regexp"
	"strings"
)

var version string
var log = logging.MustGetLogger("main")

// note the "#" -  it is added so exabgp treats the line as comment
var colorLogFormat = logging.MustStringFormatter("# %{color:bold}%{time:2006-01-02T15:04:05Z-07:00}%{color:reset}%{color} [%{level:.1s}]%{color:reset} %{message}")

func main() {
	stderrBackend := logging.NewLogBackend(os.Stderr, "", 0)
	//	stderrFormatter := logging.NewBackendFormatter(stderrBackend, colorLogFormat)
	logging.SetBackend(stderrBackend)
	logging.SetFormatter(colorLogFormat)
	logging.SetLevel(logging.NOTICE, "")

	app := cli.NewApp()
	app.Name = "HA2BGP"
	app.Usage = "Announce BGP prefixes only when services are up"
	app.Description = "HA2BGP is bridge between bgp (so far only ExaBGP) and various healthchecks\n    It's role is to announce routes when service is up and withdraw when they are down or flapping\n    For more https://github.com/efigence/go-ha2bgp"
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
			Usage:  "ss-compatible filter, set if you use non HTTP/S ports",
			EnvVar: "HA2BGP_LISTEN_FILTER",
		},
		cli.StringFlag{
			Name:   "device",
			Usage:  "specify device to check for listening IPs. Defaults to empty string which means all devices",
			EnvVar: "HA2BGP_DEVICE",
		},
		cli.StringFlag{
			Name:   "device-label",
			Usage:  "filter IPs by address label string. Accepts full name or globs (like lo:*)",
			EnvVar: "HA2BGP_DEVICE",
		},
		// cli.StringSliceFlag{
		// 	Name:   "announce, a",
		// 	Usage:  "List of IPs to announce. This WILL ignore network filter but will not ignore backend state.",
		// 	EnvVar: "HA2BGP_ANNOUNCE",
		// },
		cli.StringFlag{
			Name:   "nexthop, t",
			Value:  "self",
			Usage:  "Next hop",
			EnvVar: "HA2BGP_NEXTHOP",
		},
		cli.BoolFlag{
			Name:  "no-color",
			Usage: "disable colors",
		},
		cli.StringFlag{
			Name:  "bgp-backend",
			Value: "exabgp",
			Usage: "BGP backend. So far only 'test' and 'exabgp' are supported.",
		},
		cli.BoolFlag{
			Name:  "debug,d",
			Usage: "Debug",
		},
	}
	app.Action = func(c *cli.Context) error {
		if c.Bool("help") {
			cli.ShowAppHelp(c)
			os.Exit(1)
		}
		if !regexp.MustCompile("^(exabgptest|exabgp)").MatchString(c.GlobalString("bgp-backend")) {
			log.Errorf("bgp-backend must be one of [exabgptest,exabgp] but got [%s]", c.GlobalString("bgp-backeend"))
			os.Exit(1)
		}
		setLoggerFormat(c)

		log.Infof("Starting HA2BGP version: %s", version)
		MainLoop(c)
		return nil
	}
	app.Run(os.Args)

}

func setLoggerFormat(c *cli.Context) {
	logFormat := ""
	// add timestamp by default except for exabgp
	if c.GlobalString("bgp-backend") == "exabgp" {
		if c.Bool("debug") {
			logFormat = logFormat + "[" + strings.ToLower(c.App.Name) + "] %{time:05.0000} "
		} else {
			logFormat = logFormat + "[" + strings.ToLower(c.App.Name) + "] "
		}
	} else {
		if c.Bool("no-color") {
			logFormat = logFormat + "%{time:2006-01-02T15:04:05.0000Z-07:00} "
		} else {
			logFormat = logFormat + "%{color:bold}%{time:2006-01-02T15:04:05.0000Z-07:00}%{color:reset} "
		}
	}
	// those probably should differ but function names are useful even when not debugging
	if c.Bool("debug") {
		if c.Bool("no-color") || c.GlobalString("bgp-backend") == "exabgp" {
			logFormat = logFormat + "[%{level:.1s}] %{shortpkg}[%{longfunc}] %{message}"
		} else {
			logFormat = logFormat + "%{color}[%{level:.1s}] %{color:reset}%{shortpkg}[%{longfunc}] %{message}"
		}
	} else {
		if c.Bool("no-color") || c.GlobalString("bgp-backend") == "exabgp" {
			logFormat = logFormat + "[%{level:.1s}] %{shortpkg}[%{longfunc}] %{message}"
		} else {
			logFormat = logFormat + "%{color}[%{level:.1s}] %{color:reset}%{shortpkg}[%{longfunc}] %{message}"
		}
	}
	logFormatter := logging.MustStringFormatter(logFormat)
	if c.Bool("debug") {
		logging.SetFormatter(logFormatter)
		logging.SetLevel(logging.DEBUG, "")
	} else {
		logging.SetFormatter(logFormatter)

	}
}
