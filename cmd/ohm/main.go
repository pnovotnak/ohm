package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/pnovotnak/ohm/src/config"
	"github.com/pnovotnak/ohm/src/nextdns"
	"github.com/pnovotnak/ohm/src/ohm"
	"github.com/pnovotnak/ohm/src/types"
	"io"
	"log"
	"net/http"
	"regexp"
)

var (
	Config = &config.Config{}
)

func StreamLogs(logC chan *types.LogData) error {
	defer close(logC)

	req, err := nextdns.Get(nextdns.MakeUrl("profiles", nextdns.Profile, "logs", "stream"))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(resp.Body)
	// TODO cancel via context
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			// TODO backoff & retry
			break
		}
		prefix := nextdns.StreamingLogLineRegex.FindSubmatchIndex(line)
		// could be blank line or metadata
		if len(prefix) == 0 {
			continue
		}
		data := line[prefix[1]:]

		logData := &types.LogData{}
		err = json.Unmarshal(data, logData)
		if err != nil {
			log.Printf("unable to decode data: %s\n", data)
			continue
		}

		logC <- logData
	}
	return nil
}

func init() {
	var err error

	Config, err = config.Load()
	if err != nil {
		panic(err)
	}

	nextdns.APIKey = Config.Account.Key
	nextdns.Profile = Config.Account.Profile
}

type Route struct {
	re       *regexp.Regexp
	handlerC chan *types.LogData
}

type Router struct {
	Routes []Route
}

func (r *Router) Add(key string, bucket *config.BlockBucket) chan *types.LogData {
	// Give each chan a buffer so that they don't block the other pipeline stages
	handlerC := make(chan *types.LogData, 2)
	r.Routes = append(r.Routes, Route{
		regexp.MustCompile(fmt.Sprintf(".*%s$", key)),
		handlerC,
	})
	return handlerC
}

func (r *Router) Route(logC chan *types.LogData) {
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
	logC := make(chan *types.LogData)

	// Start the producer
	go func() {
		panic(StreamLogs(logC))
	}()

	// Start the router
	go func() {
		router.Route(logC)
	}()

	// Start the consumers
	for key, bucket := range Config.Buckets {
		go ohm.Run(key, bucket, router.Add(key, bucket))
	}

	fmt.Println("Ohm is running.")
	select {}
}
