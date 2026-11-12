package app

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouteAuthentication(t *testing.T) {
	router := NewRouter(RouterDeps{Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), TokenManager: homeTestTokens(), HomeRepository: emptyRepository{}})
	for _, tc := range []struct {
		method, path string
		status       int
	}{
		{"GET", "/healthz", 200},
		{"POST", "/api/v1/user/login", 200},
		{"POST", "/api/v1/user/register", 200},
		{"GET", "/api/v1/home", 401},
		{"GET", "/api/v1/home/featured", 401},
		{"GET", "/api/v1/home/tools", 401},
		{"GET", "/api/v1/home/banners", 401},
		{"GET", "/api/v1/home/flash", 401},
		{"GET", "/api/v1/home/news", 401},
	} {
		t.Run(tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
			if w.Code != tc.status {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
	for _, header := range []string{"Bearer", "Basic abc", "Bearer invalid"} {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/home", nil)
		req.Header.Set("Authorization", header)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != 401 {
			t.Fatalf("header=%q status=%d", header, w.Code)
		}
	}
	fetchHome(t, router, "/api/v1/home", http.StatusOK)
}
