package listen

import (
	"os/exec"
	"fmt"
	"strings"
	"bytes"
	"regexp"
	"net"
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
