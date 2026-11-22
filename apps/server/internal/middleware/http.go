package middleware

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func Decode(w http.ResponseWriter, r *http.Request, value any) error {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	d.DisallowUnknownFields()
	if d.Decode(value) != nil || d.Decode(new(any)) != io.EOF {
		return fs.ErrInvalid
	}
	return nil
}

func Result[T any](w http.ResponseWriter, value T, err error) {
	if err != nil {
		RespondError(w, err)
		return
	}
	WriteJSON(w, 200, value)
}
func Content(w http.ResponseWriter, kind string, data []byte, err error) {
	if err != nil {
		RespondError(w, err)
		return
	}
	w.Header().Set("Content-Type", kind)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(data)
}

func RespondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, fs.ErrInvalid):
		Reject(w, 400, "invalid_request")
	case errors.Is(err, fs.ErrNotExist), errors.Is(err, pgx.ErrNoRows):
		Reject(w, 404, "not_found")
	default:
		Reject(w, 503, "unavailable")
	}
}

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
