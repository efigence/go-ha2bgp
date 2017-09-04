package engine

import (
	"time"
	"sync"
)

// Route represents single route object. "Extra" is not used as a key so there can be no 2 routes with same extra field
type Route struct {
	// route
	Route string
	// next hop of the route
	NextHop string
	// extra parameters
	Extra string
	// whether route was announced
	Announced bool
	// whether route target is up
	Up bool
	LastUp time.Time
	LastDown time.Time
	LastAnnounced time.Time
	LastWithdrawn time.Time
	check RouteHealthcheck
	sync.RWMutex
}

func NewRoute(route string, nexthop string, extra string,check RouteHealthcheck) (r *Route, err error) {
	r = &Route{
		Route: route,
		NextHop: nexthop,
		Extra: extra,

	}
	return r, err
}

func (r *Route)String() (string) {
	return r.Route + "|" + r.NextHop
}

func (r *Route)SetAnnounced() {
	r.Lock()
	r.Announced = true
	r.LastAnnounced = time.Now()
	r.Unlock()
}

func (r *Route)SetWithdrawn() {
	r.Lock()
	r.Announced = false
	r.LastWithdrawn = time.Now()
	r.Unlock()
}


func (r *Route)SetUp() {
	r.Lock()
	r.Up = true
	r.LastUp = time.Now()
	r.Unlock()
}

func (r *Route)SetDown() {
	r.Lock()
	r.Up = false
	r.LastDown = time.Now()
	r.Unlock()
}
