package workspace

import (
	"encoding/json"
	"errors"
)

var ErrUnavailable = errors.New("workspace unavailable")

type Snapshot struct {
	ID          string            `json:"id"`
	Threads     json.RawMessage   `json:"threads"`
	Events      json.RawMessage   `json:"events"`
	Preferences map[string]string `json:"preferences"`
}

type Entry struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}
