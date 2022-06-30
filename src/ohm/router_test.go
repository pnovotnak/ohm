package ohm

import (
	"encoding/json"
	"github.com/pnovotnak/ohm/src/config"
	"github.com/pnovotnak/ohm/src/types"
)

import "testing"

func routerFixture() (*Router, []config.BlockBucket, error) {
	var err error
	router := Router{}
	buckets := []config.BlockBucket{
		{},
		{},
	}
	if err = buckets[0].Init("matches.com"); err != nil {
		return nil, buckets, err
	}
	if err = buckets[1].Init("no-matches.com"); err != nil {
		return nil, buckets, err
	}
	router.Add("matches.com", &buckets[0])
	router.Add("no-matches.com", &buckets[1])
	return &router, buckets, nil
}

func TestRouter_Route(t *testing.T) {
	router, _, err := routerFixture()
	if err != nil {
		t.Fatal(err)
	}
	logC := make(chan types.LogData)
	go router.Route(logC)
	exampleData1 := []byte(" {\"timestamp\":\"2022-06-30T01:13:13.440Z\",\"domain\":\"some.matches.com\",\"root\":\"matches.com\",\"tracker\":\"\",\"encrypted\":true,\"protocol\":\"DNS-over-TLS\",\"status\":\"blocked\",\"reasons\":[{\"id\":\"denylist\",\"name\":\"Denylist\"}]}")
	exampleParsed := types.LogData{}
	if err = json.Unmarshal(exampleData1, &exampleParsed); err != nil {
		t.Fatal(err)
	}
	logC <- exampleParsed
}
