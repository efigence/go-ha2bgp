package check

import (
	"fmt"
)

type Check struct {
	checks map[string]RouteHealthcheck

}

type RouteHealthcheck interface {
	IsRouteUp(route string, nexthop string) bool
	Check() error
}

func NewCheck() (c *Check, err error) {
	return &Check{
		checks: make(map[string]RouteHealthcheck),
	}, err
}

func(c *Check)Register(name string, check RouteHealthcheck){
	c.checks[name]=check
}

func (c *Check)Check() error {
	for name, check := range c.checks {
		err := check.Check()
		if err != nil {
			return fmt.Errorf("Check() for [%s] failed with: %s", name, err)
		}
	}
	return nil
}

func (c *Check)IsRouteUp(route string, nexthop string) bool {
	// fail if none is registered
	if len(c.checks) < 1 {
		return false
	}
	// fail on first failing check
	for _, check := range c.checks {
		if !check.IsRouteUp(route,nexthop) {return false}
	}
	return true
}
