package ohm

import (
	"fmt"
	"regexp"

	"github.com/pnovotnak/ohm/src/config"
	"github.com/pnovotnak/ohm/src/types"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	blockedQueries = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ohm_blocked_queries",
		Help: "Blocked queries by domain",
	}, []string{
		"domain",
	})
	allowedQueries = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ohm_allowed_queries",
		Help: "Allowed queries by domain",
	}, []string{
		"domain",
	})
)

type Route struct {
	key      string
	re       *regexp.Regexp
	handlerC chan types.LogData
}

type Router struct {
	Routes []Route
}

func (r *Router) Add(key string, bucket *config.BlockBucket) chan types.LogData {
	// Give each chan a buffer so that they don't block the other pipeline stages
	handlerC := make(chan types.LogData, 2)
	r.Routes = append(r.Routes, Route{
		key,
		regexp.MustCompile(fmt.Sprintf(".*%s$", key)),
		handlerC,
	})
	return handlerC
}

func (r *Router) Route(logC chan types.LogData) {
	for {
		logEntry := <-logC
		for _, route := range r.Routes {
			if route.re.MatchString(logEntry.Domain) {
				if logEntry.Status == types.StatusBlocked {
					blockedQueries.With(prometheus.Labels{"domain": route.key}).Inc()
				} else {
					allowedQueries.With(prometheus.Labels{"domain": route.key}).Inc()
				}
				route.handlerC <- logEntry
			}
		}
	}
}
