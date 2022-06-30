package main

import (
	_ "embed"
	"github.com/pnovotnak/ohm/src/config"
	"github.com/pnovotnak/ohm/src/nextdns"
	"github.com/pnovotnak/ohm/src/ohm"
	"github.com/pnovotnak/ohm/src/types"
	"log"
	"time"
)

var (
	Config = &config.Config{}
	//go:embed config.yaml
	configRaw []byte
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

	// Start the producer
	go func() {
		var retryCount int

		// TODO move to constants
		clampMaxRetrySleep := 60 * time.Second
		resetAfter := 10 * time.Minute

		lastCrash := time.Now()
		for {
			err := nextdns.StreamLogs(logC)
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

	log.Println("Ω")
	select {}
}
