package routes_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Ender-swe/cueia.io/apps/server/cmd/api/routes"
	"github.com/jackc/pgx/v5"
)

func TestProviders(t *testing.T) {
	auth := func(_ context.Context, token string) (string, error) {
		switch token {
		case "valid":
			return "workspace", nil
		case "db-down":
			return "", errors.New("private database details")
		default:
			return "", pgx.ErrNoRows
		}
	}
	probes := []routes.OptionalProvider{
		{Name: "first"},
		{Name: "second", Check: func(context.Context) error { return nil }},
		{Name: "third", Check: func(context.Context) error { return errors.New("private provider details") }},
	}

	for _, tc := range []struct {
		name, authorization string
		providers           []routes.OptionalProvider
		status              int
		body                string
	}{
		{"bootstrap", "Bearer valid", nil, 200, "[]"},
		{"states", "Bearer valid", probes, 200, `[{"name":"first","status":"disabled"},{"name":"second","status":"configured"},{"name":"third","status":"unavailable"}]`},
		{"missing", "", probes, 401, `{"code":"unauthorized","message":"Unauthorized"}`},
		{"invalid", "Bearer expired", probes, 401, `{"code":"unauthorized","message":"Unauthorized"}`},
		{"database unavailable", "Bearer db-down", probes, 503, `{"code":"unavailable","message":"Service Unavailable"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/providers", nil)
			req.Header.Set("Authorization", tc.authorization)
			rec := httptest.NewRecorder()
			mux := http.NewServeMux()
			routes.RegisterStatus(mux, auth, tc.providers)
			mux.ServeHTTP(rec, req)
			if rec.Code != tc.status || strings.TrimSpace(rec.Body.String()) != tc.body {
				t.Errorf("unexpected response: %d %s", rec.Code, rec.Body)
			}
		})
	}
}

func TestPublicHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	mux := http.NewServeMux()
	routes.RegisterStatus(mux, nil, nil)
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/health", nil))
	if rec.Code != 200 || strings.TrimSpace(rec.Body.String()) != `{"status":"ok"}` {
		t.Fatalf("unexpected health response: %d %s", rec.Code, rec.Body)
	}
}
