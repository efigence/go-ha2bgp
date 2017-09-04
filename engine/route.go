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
	LastStateChange time.Time
	UpCount int
	DownCount int
	// 0-100 score that tracks how often routes flap. 1 flap is worth 10 points, deduced 1 each time state is steady
	FlapScore int
	checkFunc RouteHealthcheck
	sync.Mutex
}

func NewRoute(route string, nexthop string, extra string,check RouteHealthcheck) (r *Route, err error) {
	r = &Route{
		Route: route,
		NextHop: nexthop,
		Extra: extra,
		checkFunc: check,

	}
	return r, err
}

func (r *Route)String() (string) {
	return r.Route + "|" + r.NextHop
}
func (r *Route)Check() (bool) {
	up := r.checkFunc.IsRouteUp(r.Route, r.NextHop)
	r.Lock()
	defer r.Unlock()
	if up != r.Up {
		r.LastStateChange = time.Now()
		// reset counters but only if in stable state
		// that can be used for flap detection
		if up {
			if r.UpCount > 10 {
				r.DownCount = 0
			}
		} else {
			if r.DownCount > 10 {
				r.UpCount = 0
			}
		}
		if r.FlapScore <= 100 {
			r.FlapScore = r.FlapScore + 10
		}
	} else {
		if r.FlapScore > 0 {
			r.FlapScore = r.FlapScore - 1
		}
	}
	if up {
		r.UpCount++
	} else {
		r.DownCount++
	}
	return up
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
