package engine



import (
	"time"
	"sync"
	"github.com/op/go-logging"
)


var log = logging.MustGetLogger("main")

type Engine struct {
	Routes map[string]*Route
	// minimum interval between changing states
	MinFlapInterval time.Duration
	// delay between interface up and announcing it to world
	UpDelay time.Duration
	// delay between downing interface and announcement
	DownDelay time.Duration
	// how often to repeat already up announcements
	ReAnnounce time.Duration
	// how often to repeat already up withdraws
	ReWithdraw time.Duration
	routeEngine RouteEngine
	iter int
	sync.RWMutex
}


func NewEngine(r RouteEngine) (e *Engine,err error) {
	return &Engine{
		Routes: make(map[string]*Route),
		UpDelay: time.Duration(6 * time.Second),
		DownDelay: time.Duration(3 * time.Second),
		ReAnnounce: time.Duration(300 * time.Second),
		ReWithdraw: time.Duration(900 * time.Second),
		routeEngine: r,
	}, nil
}


type RouteHealthcheck interface {
	IsRouteUp(route string, nexthop string) bool
}


// register route
// each route also needs to have healthcheck implementing RouteHealthcheck interface
// each route will be set ONCE, which means any subsequent changes to same tuple of route+nexthop will be ignored.
// Drop it before if you want to change config
func (e *Engine)RegisterRoute (route string, nexthop string, extra string, healthcheck RouteHealthcheck) (err error) {
	newRoute,err := NewRoute(route,nexthop,extra, healthcheck)
	if err != nil {return err }
	routeStr := newRoute.String()
	e.Lock()
	defer e.Unlock()
	if _, ok := e.Routes[routeStr]; !ok {
		e.Routes[routeStr] = newRoute
		log.Noticef("Registered new route  %s -> %s",route, nexthop)
	}
	return err
}

type RouteEngine interface {
	AnnounceRoute(route string, nexthop string, extra ...string) (err error)
	WithdrawRoute(route string, nexthop string, extra ...string) (err error)
}



func (e *Engine) UpdateState() {
	// first we update state; then generate announcements based on timeouts and throttling
	e.iter++
	for name, route := range e.Routes {
		alive := route.Check()
		if alive {
			route.SetUp()
		} else {
			route.SetDown()
		}
		if (e.iter % 50) == 10 && route.FlapScore > 50 {
			log.Errorf("Route %s is flapping, score [%d]", route.Route, route.FlapScore)
		}
		switch {
			// drop anything that is flapping too much
		case route.FlapScore > 50 && route.Announced:
			err := e.routeEngine.WithdrawRoute(route.Route, route.NextHop, route.Extra)
			if err != nil {log.Errorf("error while withdrawing route %s: %s [%+v]",name,err, route)}
			route.SetWithdrawn()
		case route.FlapScore > 50:
			// is down already, nothing to do

		case route.Up &&  !route.Announced && route.UpCount < 3:
			// wait till it is stable
		case route.Up && !route.Announced &&  time.Now().Sub(route.LastStateChange) < e.UpDelay:
			// wait a delay before announcing it
		case route.Up && !route.Announced:
			err := e.routeEngine.AnnounceRoute(route.Route, route.NextHop, route.Extra)
			if err != nil {log.Errorf("error while announcing route %s: %s [%+v]",name,err, route)}
			route.SetAnnounced()
		case !route.Up && route.Announced && route.DownCount < 3:
			// don't drop route immediately (for stuff like daemon restart)
		case !route.Up && route.Announced:
			// drop the route
			err := e.routeEngine.WithdrawRoute(route.Route, route.NextHop, route.Extra)
			if err != nil {log.Errorf("error while withdrawing route %s: %s [%+v]",name,err, route)}
			route.SetWithdrawn()
		case route.Up && time.Now().Sub(route.LastAnnounced) > e.ReAnnounce && e.ReAnnounce != 0:
			// re-announce every ReAnnounce timer
			err := e.routeEngine.AnnounceRoute(route.Route, route.NextHop, route.Extra)
			if err != nil {log.Errorf("error while re-announcing route %s: %s [%+v]",name,err, route)}
			route.SetAnnounced()
		case !route.Up && time.Now().Sub(route.LastWithdrawn) > e.ReWithdraw && e.ReWithdraw != 0:
			err := e.routeEngine.WithdrawRoute(route.Route, route.NextHop, route.Extra)
			if err != nil {log.Errorf("error while re-announcing route %s: %s [%+v]",name,err, route)}
			route.SetWithdrawn()
		}
	}

}
