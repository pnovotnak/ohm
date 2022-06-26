package types

import "time"

type Status string

const StatusBlocked Status = "blocked"

type LogData struct {
	Timestamp time.Time `json:"timestamp"`
	Domain    string    `json:"domain"`
	Root      string    `json:"root"`
	Tracker   string    `json:"tracker,omitempty"`
	Encrypted bool      `json:"encrypted"`
	Protocol  string    `json:"protocol"`
	Status    Status    `json:"status"`
	Reasons   []string  `json:"reasons"`
}

type LogResponse struct {
	Data []LogData `json:"data"`
	Meta struct {
		Pagination struct {
			Cursor string `json:"cursor"`
		} `json:"pagination"`
		Stream struct {
			Id string `json:"id"`
		} `json:"stream"`
	} `json:"meta"`
}

type DenyEntry struct {
	// Id is also the TLD to use and is matched as `*.<id>`
	Id     string `json:"id,omitempty"`
	Active bool   `json:"active"`
}

type Denylist struct {
	Data []DenyEntry `json:"data"`
}
