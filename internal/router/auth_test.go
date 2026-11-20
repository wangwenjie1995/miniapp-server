package router_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"miniapp-server/internal/platform"
	"miniapp-server/internal/router"
)

func TestPublicRoutes(t *testing.T) {
	router := router.NewWithDependencies(platform.Resources{Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Tokens: homeTestTokens()}, router.Repositories{Home: emptyRepository{}})
	for _, tc := range []struct {
		method, path string
		status       int
	}{
		{"GET", "/healthz", 200},
		{"POST", "/api/v1/user/login", 200},
		{"POST", "/api/v1/user/register", 200},
		{"GET", "/api/v1/home", 200},
		{"GET", "/api/v1/home/featured", 200},
		{"GET", "/api/v1/home/tools", 200},
		{"GET", "/api/v1/home/banners", 200},
		{"GET", "/api/v1/home/flash", 200},
		{"GET", "/api/v1/home/news", 200},
	} {
		t.Run(tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
			if w.Code != tc.status {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
	token, _, err := homeTestTokens().Generate(1)
	if err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"", "/featured", "/tools", "/banners", "/flash", "/news"} {
		for _, header := range []string{"", "Bearer " + token, "Bearer invalid"} {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/home"+suffix, nil)
			req.Header.Set("Authorization", header)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("path=%s header=%q status=%d", req.URL.Path, header, w.Code)
			}
		}
	}
}
