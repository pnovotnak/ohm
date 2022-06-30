package nextdns

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pnovotnak/ohm/src/types"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

const (
	NextDNSAPI = "https://api.nextdns.io"
	APIHeader  = "x-api-key" // HTTP headers are case-insensitive
)

var (
	StreamingLogLineRegex = regexp.MustCompile(`^data:\s+`)
	APIKey                string
	Profile               string
)

func MakeUrl(pathParts ...string) string {
	return fmt.Sprintf("%s/%s", NextDNSAPI, strings.Join(pathParts, "/"))
}

func Get(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add(APIHeader, APIKey)
	req.Header.Add("content-type", "application/json")
	return req, err
}

func Patch(url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPatch, url, body)
	req.Header.Add(APIHeader, APIKey)
	req.Header.Add("content-type", "application/json")
	return req, err
}

func GetLogs() (*http.Response, error) {
	req, err := Get(MakeUrl("profiles", Profile, "logs"))
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func StreamLogs(logC chan types.LogData) error {
	req, err := Get(MakeUrl("profiles", Profile, "logs", "stream"))
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
		prefix := StreamingLogLineRegex.FindSubmatchIndex(line)
		// could be blank line or metadata
		if len(prefix) == 0 {
			continue
		}
		data := line[prefix[1]:]

		logData := types.LogData{}
		err = json.Unmarshal(data, &logData)
		if err != nil {
			log.Printf("unable to unmarshal data: '%s\n' (%s)", data, err)
			continue
		}

		logC <- logData
	}
}

func SetBlock(key string, value bool) (*http.Response, error) {
	payload, err := json.Marshal(types.DenyEntry{Active: value})
	if err != nil {
		return nil, err
	}
	url := MakeUrl("profiles", Profile, "denylist", key)
	req, err := Patch(url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp, err
	}
	return resp, err
}
