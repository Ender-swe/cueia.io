package main

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

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
	probes := []optionalProvider{
		{"first", nil},
		{"second", func(context.Context) error { return nil }},
		{"third", func(context.Context) error { return errors.New("private provider details") }},
	}

	for _, tc := range []struct {
		name, authorization string
		providers           []optionalProvider
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
			req := httptest.NewRequest("GET", "/api/providers", nil)
			req.Header.Set("Authorization", tc.authorization)
			rec := httptest.NewRecorder()
			routes(auth, tc.providers).ServeHTTP(rec, req)
			if rec.Code != tc.status || strings.TrimSpace(rec.Body.String()) != tc.body {
				t.Errorf("unexpected response: %d %s", rec.Code, rec.Body)
			}
		})
	}
}

func TestPublicHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	routes(nil, nil).ServeHTTP(rec, httptest.NewRequest("GET", "/api/health", nil))
	if rec.Code != 200 || strings.TrimSpace(rec.Body.String()) != `{"status":"ok"}` {
		t.Fatalf("unexpected health response: %d %s", rec.Code, rec.Body)
	}
}
