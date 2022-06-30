package ohm

import (
	"encoding/json"
	"github.com/pnovotnak/ohm/src/config"
	"github.com/pnovotnak/ohm/src/types"
)

import "testing"

func TestRouter_Add(t *testing.T) {
	var err error
	router := Router{}
	bucketMatch := config.BlockBucket{}
	if err = bucketMatch.Init("matches.com"); err != nil {
		t.Fatal(err)
	}
	bucketNoMatch := config.BlockBucket{}
	if err = bucketNoMatch.Init("no-matches.com"); err != nil {
		t.Fatal(err)
	}
	router.Add("matches.com", &bucketMatch)
	router.Add("no-matches.com", &bucketNoMatch)
	logC := make(chan types.LogData)
	go router.Route(logC)
	exampleData1 := []byte(" {\"timestamp\":\"2022-06-30T01:13:13.440Z\",\"domain\":\"some.matches.com\",\"root\":\"matches.com\",\"tracker\":\"\",\"encrypted\":true,\"protocol\":\"DNS-over-TLS\",\"status\":\"blocked\",\"reasons\":[{\"id\":\"denylist\",\"name\":\"Denylist\"}]}")
	exampleParsed := types.LogData{}
	if err = json.Unmarshal(exampleData1, &exampleParsed); err != nil {
		t.Fatal(err)
	}
	logC <- exampleParsed

}

func TestRouter_Route(t *testing.T) {

}
