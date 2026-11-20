package volume

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"time"
)

// A request runs in a single authenticated SSH subsystem session.
type Request struct {
	Operation string `json:"operation"`
	Name      string `json:"name,omitempty"`
	Data      []byte `json:"data,omitempty"`
}
type Entry struct {
	Filename string
	FileMode fs.FileMode
	Length   int64
	Modified time.Time
}

func (e Entry) Name() string               { return e.Filename }
func (e Entry) Mode() fs.FileMode          { return e.FileMode }
func (e Entry) Type() fs.FileMode          { return e.FileMode.Type() }
func (e Entry) Size() int64                { return e.Length }
func (e Entry) ModTime() time.Time         { return e.Modified }
func (e Entry) IsDir() bool                { return e.FileMode.IsDir() }
func (e Entry) Sys() any                   { return nil }
func (e Entry) Info() (fs.FileInfo, error) { return e, nil }
func entry(f fs.FileInfo) Entry            { return Entry{f.Name(), f.Mode(), f.Size(), f.ModTime()} }

type Response struct {
	Error   string  `json:"error,omitempty"`
	Data    []byte  `json:"data,omitempty"`
	Key     string  `json:"key,omitempty"`
	Stat    Entry   `json:"stat,omitempty"`
	Entries []Entry `json:"entries,omitempty"`
}

func (s *Store) Serve(in io.Reader, out io.Writer) error {
	var q Request
	if err := json.NewDecoder(in).Decode(&q); err != nil {
		return err
	}
	var r Response
	var err error
	switch q.Operation {
	case "reconcile":
		err = s.Reconcile()
	case "version":
		r.Data, err = s.ReadVersion(q.Name)
	case "publish":
		r.Key, err = s.Publish(q.Name, q.Data)
	default:
		if !fs.ValidPath(q.Name) {
			err = fs.ErrInvalid
			break
		}
		switch q.Operation {
		case "write":
			err = s.WriteFile(q.Name, q.Data, 0600)
		case "read":
			var f fs.File
			f, err = s.Open(q.Name)
			if err == nil {
				defer f.Close()
				var info fs.FileInfo
				info, err = f.Stat()
				if err == nil {
					r.Stat = entry(info)
					if info.Mode().IsRegular() {
						r.Data, err = io.ReadAll(f)
					} else if !info.IsDir() {
						err = fs.ErrInvalid
					}
				}
			}
		case "list":
			var entries []fs.DirEntry
			entries, err = fs.ReadDir(s, q.Name)
			for _, e := range entries {
				info, infoErr := e.Info()
				if infoErr != nil {
					err = infoErr
					break
				}
				r.Entries = append(r.Entries, entry(info))
			}
		default:
			err = fs.ErrInvalid
		}
	}
	if err != nil {
		r = Response{Error: "unavailable"}
		if errors.Is(err, fs.ErrInvalid) {
			r.Error = "invalid"
		}
		if errors.Is(err, fs.ErrNotExist) || errors.Is(err, fs.ErrPermission) {
			r.Error = "missing"
		}
	}
	return json.NewEncoder(out).Encode(r)
}
