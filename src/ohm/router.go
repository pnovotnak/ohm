package ohm

import (
	"fmt"
	"github.com/pnovotnak/ohm/src/config"
	"github.com/pnovotnak/ohm/src/types"
	"regexp"
)

type Route struct {
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
		regexp.MustCompile(fmt.Sprintf(".*%s$", key)),
		handlerC,
	})
	return handlerC
}

func (r *Router) Route(logC chan types.LogData) {
	for {
		logEntry := <-logC
		for _, handler := range r.Routes {
			if handler.re.MatchString(logEntry.Domain) {
				handler.handlerC <- logEntry
			}
		}
	}
}
