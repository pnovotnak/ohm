package main

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/pnovotnak/ohm/src/config"
	"github.com/pnovotnak/ohm/src/nextdns"
	"github.com/pnovotnak/ohm/src/ohm"
	"github.com/pnovotnak/ohm/src/types"
	"log"
	"net/http"
	"regexp"
	"time"
)

var (
	Config = &config.Config{}
	//go:embed config.yaml
	configRaw []byte
)

func StreamLogs(logC chan types.LogData) error {
	req, err := nextdns.Get(nextdns.MakeUrl("profiles", nextdns.Profile, "logs", "stream"))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	log.Printf("log streamer started")

	reader := bufio.NewReader(resp.Body)
	// TODO cancel via context
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return err
		}
		prefix := nextdns.StreamingLogLineRegex.FindSubmatchIndex(line)
		// could be blank line or metadata
		if len(prefix) == 0 {
			continue
		}
		data := line[prefix[1]:]

		logData := types.LogData{}
		err = json.Unmarshal(data, &logData)
		if err != nil {
			log.Printf("unable to decode data: %s\n", data)
			continue
		}

		logC <- logData
	}
}

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

func main() {
	var router Router
	logC := make(chan types.LogData)

	// Start the producer
	go func() {
		var retryCount int

		// TODO move to constants
		clampMax := 60 * time.Second
		resetAfter := 10 * time.Minute

		lastCrash := time.Now()
		for {
			err := StreamLogs(logC)
			log.Printf("log streamer crashed: %s", err)
			if time.Now().Sub(lastCrash) > resetAfter {
				retryCount = 0
				continue
			}
			toSleep := time.Duration(retryCount*retryCount) * time.Second
			if toSleep > clampMax {
				toSleep = clampMax
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
