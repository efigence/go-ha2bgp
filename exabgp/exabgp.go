package exabgp




import (
	"fmt"
	"strings"
)

func NewExaBGP3() (*Exa3, error) {
	return &Exa3{}, nil
}

type Exa3 struct {
	sendCmd func(string) error

}

// announce route to host
func (s *Exa3)AnnounceRoute(route string, nexthop string, extra ...string) (err error) {
	s.SendCmd(fmt.Sprintf("announce route %s next-hop %s %s",route, nexthop, strings.Join(extra, " ")))
	// exabgp 3.x doesnt provide any feedback soo
	// 4.x does
	return
}

// announce slice of routes to host
func (s *Exa3)AnnounceRouteSlice(route []string, nexthop string,extra ...string) (err error) {
	for id, r := range route {
		err := s.AnnounceRoute(r, nexthop, extra...)
		if err != nil { return fmt.Errorf("failed to announce route %s[%d]: %s", r, id, err) }
	}
	return nil
}


//withdraw a route to host
func (s *Exa3)WithdrawRoute(route string, nexthop string, extra ...string) (err error) {
	s.SendCmd(fmt.Sprintf("withdraw route %s next-hop %s %s",route, nexthop, strings.Join(extra, " ")))
	// exabgp 3.x doesnt provide any feedback soo
	// 4.x does
	return
}

// withdraw slice of routes from host
func (s *Exa3)WithdrawRouteSlice(route []string, nexthop string,extra ...string) (err error) {
	for id, r := range route {
		err := s.WithdrawRoute(r, nexthop, extra...)
		if err != nil { return fmt.Errorf("failed to withdraw route %s[%d]: %s", r, id, err) }
	}
	return nil
}

func (s *Exa3)SendCmd(cmd string) (err error) {
	if s.sendCmd == nil {
		fmt.Printf("%s\n", cmd)
		return
	} else {
		return s.sendCmd(cmd)
	}
}
