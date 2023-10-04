package main

import (
	_ "embed"
	"log"
	"net/http"
	"time"

	"github.com/pnovotnak/ohm/src/config"
	"github.com/pnovotnak/ohm/src/nextdns"
	"github.com/pnovotnak/ohm/src/ohm"
	"github.com/pnovotnak/ohm/src/types"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	Config = &config.Config{}
	//go:embed config.yaml
	configRaw   []byte
	MetricsAddr = ":9091"

	logStreamerRestarts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ohm_log_streamer_restarts",
		Help: "The total number of log streamer restart events",
	})
)

func init() {
	var err error

	Config, err = config.Parse(configRaw)
	if err != nil {
		panic(err)
	}

	if err = Config.Validate(); err != nil {
		panic(err)
	}

	nextdns.APIKey = Config.NextDNS.Key
	nextdns.Profile = Config.NextDNS.Profile
}

func main() {
	var router ohm.Router
	logC := make(chan types.LogData)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		_ = http.ListenAndServe(MetricsAddr, nil)
	}()

	// Start the producer
	go func() {
		var err error
		var retryCount int
		var lineID string

		// TODO move to constants
		clampMaxRetrySleep := 60 * time.Second
		resetAfter := 10 * time.Minute

		lastCrash := time.Now()
		for {
			lineID, err = nextdns.StreamLogs(logC, lineID)
			logStreamerRestarts.Inc()
			if time.Since(lastCrash) > resetAfter {
				retryCount = 0
				continue
			} else {
				log.Printf("log streamer restart attempt #%d: (previous error: %s)", retryCount, err)
			}
			toSleep := time.Duration(retryCount*retryCount) * time.Second
			if toSleep > clampMaxRetrySleep {
				toSleep = clampMaxRetrySleep
			}
			time.Sleep(toSleep)
			lastCrash = time.Now()
			retryCount += 1
		}
	}()

	// Start the router
	go func() {
		router.Route(logC)
	}()

	// Start the consumers
	for key, bucket := range Config.Buckets {
		go ohm.Run(key, bucket, router.Add(key, bucket))
	}

	log.Printf("Ω started, serving metrics on %s", MetricsAddr)
	select {}
}
