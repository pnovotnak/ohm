package nextdns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pnovotnak/ohm/src/types"
	"io"
	"net/http"
	"regexp"
	"strings"
)

const (
	NextDNSAPI = "https://api.nextdns.io"
	APIHeader  = "x-api-key" // HTTP headers are case-insensitive
)

var (
	StreamingLogLineRegex = regexp.MustCompile("(^data:\\s+)")
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

func SetBlock(key string, value bool) error {
	payload, err := json.Marshal(types.DenyEntry{Active: value})
	url := MakeUrl("profiles", Profile, "denylist", key)
	req, err := Patch(url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	fmt.Printf("block request to %s returned: %d\n", url, resp.StatusCode)
	return err
}
