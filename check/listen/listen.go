package listen

import (
	"os/exec"
	"fmt"
	"strings"
	"bytes"
	"regexp"
	"net"
	"sync"
)
var reMulticharWhitespace = regexp.MustCompile(`\s+`)
// ss pisses over "how you are supposed to write IPv6" so functions like net.SplitHostPort do not work
var reSplitAddr = regexp.MustCompile(`^(.+)\:(.+?)$`)

func GetListeningIp(socketType string, filter string) (ip []net.IP, err error)  {
	execArgs := []string{ "-l", "-n"}
	if socketType == "tcp" {
		execArgs = append(execArgs, "-t")
	} else if socketType == "udp" {
		execArgs = append(execArgs, "-u")
	} else {
		return ip, fmt.Errorf("Unsupported socket type[%s]")
	}
	if len(filter) > 0 {
		execArgs = append(execArgs, "( " + filter + " )")
	}
	cmd := exec.Command("ss",execArgs...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return ip, fmt.Errorf("Error starting ss: %s", err)
	}
	if err := cmd.Wait(); err != nil {
		return ip, fmt.Errorf("Error while running ss: %s [stdout:%s, stderr:%s]",stdout.String(), stderr.String())
	}
	lines := strings.Split(stdout.String(), "\n")
	if len(lines) < 1 {
		return ip, fmt.Errorf("ss returned nothing [stderr:%s]",stderr.String())
	} else if len(lines) == 1 {
		return ip, nil
	}
	header := reMulticharWhitespace.Split(lines[0],6)
	if len(header) < 5 || !strings.Contains(header[3], "Local") || !strings.Contains(header[4],"Address")  {
		return ip, fmt.Errorf("Unexpected header returned from ss, aborting (version mismatch?)")
	}
	dedupeIp := make(map[string]bool)
	for _, line := range lines[1:] {
		data := reMulticharWhitespace.Split(line,6)
		if len(data) < 5 { continue }
		out := reSplitAddr.FindStringSubmatch(data[3])
		var addr net.IP
		if len(out) > 2 {
			addr = net.ParseIP(out[1])
		} else  {
			return ip,fmt.Errorf("Can't decode IP from %s", data[3])
		}
		if addr == nil && strings.Contains(out[1],"*") {
			addr = net.ParseIP(`0.0.0.0`)
		}
		if _, ok := dedupeIp[addr.String()]; !ok {
			dedupeIp[addr.String()]=true
			ip = append(ip, addr)
		}
	}
	return ip, nil
}

type Check struct {
	socketType string
	filter string
	ipAlive map[string]bool
	sync.RWMutex
	newIpHook func(ip net.IP)
}


func NewCheck(socketType string, filter string) (c *Check, err error) {
	c = &Check{
		socketType: socketType,
		filter: filter,
		ipAlive: make(map[string]bool),
	}
	return c, err
}


// hook that will be called when new listening IP is found
func (c *Check)SetNewIpHook(f func(ip net.IP)) {
	c.Lock()
	c.newIpHook = f
	c.Unlock()
}



// Check() runs listen check and updates state table
// It does keep a memory of IPs that were listening and stopped for debug purposes
// it can be accessed from DebugListenState()
func (c *Check)Check() (err error) {
	ip, err := GetListeningIp(c.socketType, c.filter)
	if err != nil { return err }
	ipMap := make(map[string]bool)
	c.Lock()
	defer c.Unlock()
	for _, ip := range ip {
		// new IPs are kicked to goroutine so they can lock
		// in peace until Check() is finished
		if _, ok := c.ipAlive[ip.String()]; !ok {
			go c.newIpHook(ip)
		}
		ipMap[ip.String()] = true
		c.ipAlive[ip.String()] = true

	}
	// clear any IPs that do not exist
	for ip, _ := range c.ipAlive  {
		if _, ok := ipMap[ip]; !ok {
			c.ipAlive[ip] = false
		}
	}
	return err
}


// check if route is up (we know it exists *and* last check indicates something is listening on it) or down (down or never existed in the first place)
func (c *Check)IsRouteUp(route string, nexthop string) bool {
	c.RLock()
	defer c.RUnlock()
	if up, ok := c.ipAlive[route]; ok {
		return up
	} else {
		return false
	}
}

// return internal state of check
func (c *Check)DebugListenState() (interface{}) {
	return c.ipAlive
}
