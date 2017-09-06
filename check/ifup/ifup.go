package ifup


import (
	"net"
	"os/exec"
	"bufio"
	"bytes"
	"strings"
	"fmt"
	"sync"
)

// iproute2 binary location. If you need to change it for some reason cange that before calling anything else/
var IpCmd = `/sbin/ip`

// GetLocalIp returns a list of IPs that are up. Specify device or pass blank string for all
// `device` is name of the device
// `label` is pattern (glob) of device labels, so for example you can search for all eth0 IPs that have label "frontend"
func GetLocalIp (device string, label string) (ip []net.IP, err error) {
	execArgs := []string{
		`-o`, //oneline mode
		// ^-global iproute2 modifiers must be before ip command
		`address`,
		`show`,
		`up`, // only ones that are up interest us
	}
	if len(device) > 0 {
		execArgs = append(execArgs, `dev`,device)
	}
	if len(label) > 0 {
		execArgs = append(execArgs,`label`,label)
	}

	var stderr bytes.Buffer
	cmd := exec.Command(IpCmd,execArgs...)
	cmd.Stderr = &stderr
	stdout, err := cmd.StdoutPipe();
	scanner := bufio.NewScanner(stdout)
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf(`error running ip[%s]: %s, stderr: %s`, IpCmd, err, stderr.String())
	}
	for scanner.Scan() {
		// example:
		//1: lo    inet 127.0.0.1/8 scope host lo\       valid_lft forever preferred_lft forever
		fields := strings.Fields(scanner.Text())
		addr,_,err := net.ParseCIDR(fields[3])
		if addr != nil && err == nil {
			ip = append(ip, addr)
		}
	}
	err = cmd.Wait()
	if err != nil {
		return nil, fmt.Errorf(`error running ip[%s]: %s, stderr: %s`, IpCmd, err, stderr.String())
	}
	return ip, err
}

type Check struct {
	device string
	label string
	ipAlive *map[string]bool
	sync.RWMutex
}

func NewCheck(device string, label string) (c *Check, err error) {
	_, err = GetLocalIp(`lo`,``)
	m := make(map[string]bool);
	return &Check{
		device: device,
		label: label,
		ipAlive: &m,
	}, err
}



func (c *Check)Check() (err error) {
	ipList, err := GetLocalIp(c.device, c.label)
	aliveIpMap := make(map[string]bool,len(ipList))
	// iterate if ok, leave empty list if not
	// clearing IP list is correct behaviour if check for some reason dies
	// that is because node is probably either sick (OOM etc.) or misconfigured
	if err == nil {
		for _, ip := range ipList {
			aliveIpMap[ip.String()] = true
		}
	}
	c.Lock()
	c.ipAlive = &aliveIpMap
	c.Unlock()
	return err
}


func (c *Check)IsRouteUp(route string, nexthop string) bool {
	c.RLock()
	defer c.RUnlock()
	if up, ok := (*c.ipAlive)[route]; ok {
		return up
	} else {
		return false
	}
}
