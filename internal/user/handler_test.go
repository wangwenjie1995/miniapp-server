package user

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type registrationRepo struct{ err error }

func (r registrationRepo) CreateUser(_ context.Context, username, hash string) (*User, error) {
	if r.err != nil {
		return nil, r.err
	}
	return &User{ID: 1, Username: username, PasswordHash: hash}, nil
}
func (registrationRepo) FindUserByUsername(context.Context, string) (*User, error) { return nil, nil }

func TestRegisterResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name, body string
		repoErr    error
		status     int
		message    string
	}{
		{"missing both", `{}`, nil, 400, "请输入用户名；请输入密码"},
		{"short both", `{"username":"ab","password":"123456"}`, nil, 400, "用户名长度须为 4 到 24 个字符；密码长度须为 8 到 72 个字符"},
		{"malformed json", `{`, nil, 400, "请求格式错误"},
		{"wrong type", `{"username":12,"password":"testpass123"}`, nil, 400, "请求格式错误"},
		{"duplicate", `{"username":"alice","password":"testpass123"}`, ErrUsernameExists, 409, "用户名已存在"},
		{"unexpected error", `{"username":"alice","password":"testpass123"}`, errors.New("database error"), 500, "注册失败"},
		{"success", `{"username":"alice","password":"testpass123"}`, nil, 200, `"username":"alice"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			RegisterRoutes(router.Group("/api/v1"), NewService(registrationRepo{tc.repoErr}), slog.New(slog.NewTextHandler(io.Discard, nil)))
			req := httptest.NewRequest(http.MethodPost, "/api/v1/user/register", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.status || !strings.Contains(w.Body.String(), tc.message) {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
}
