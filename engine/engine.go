package engine



import (
//	"time"
//	"sync"
)




type Engine struct {
	Routes map[string]*Route
}


func NewEngine() (e *Engine,err error) {
	return &Engine{
		Routes: make(map[string]*Route),
	}, nil
}


type RouteHealthcheck interface {
	IsRouteUp(route string, nexthop string) bool
}


// register a block of routes
// each route also needs to have healthcheck implementing RouteHealthcheck interface
func (e *Engine)RegisterRoutes (routes []string, nexthop string, extra string, healthcheck RouteHealthcheck) (err error) {
	for _, route := range routes {
		newRoute,err := NewRoute(route,nexthop,extra, healthcheck)
		if err != nil {return err }
		routeStr := newRoute.String()
		if _, ok := e.Routes[routeStr]; !ok {
			e.Routes[routeStr] = newRoute
		}
	}
	return err
}
