package user

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
		code       int
		message    string
	}{
		{"missing both", `{}`, nil, 200, 4000, "请输入用户名；请输入密码"},
		{"short both", `{"username":"ab","password":"123456"}`, nil, 200, 4000, "用户名长度须为 4 到 24 个字符；密码长度须为 8 到 72 个字符"},
		{"malformed json", `{`, nil, 200, 4000, "请求格式错误"},
		{"wrong type", `{"username":12,"password":"testpass123"}`, nil, 200, 4000, "请求格式错误"},
		{"duplicate", `{"username":"alice","password":"testpass123"}`, ErrUsernameExists, 200, 5001, "用户名已存在"},
		{"unexpected error", `{"username":"alice","password":"testpass123"}`, errors.New("database error"), 500, 0, "注册失败"},
		{"success", `{"username":"alice","password":"testpass123"}`, nil, 200, 0, `"username":"alice"`},
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
			if tc.status == 500 {
				if w.Body.String() != "注册失败" || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/plain") {
					t.Fatalf("expected plain text server error: %s", w.Body.String())
				}
				return
			}
			var body struct {
				Code *int            `json:"code"`
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Code == nil || *body.Code != tc.code {
				t.Fatalf("unexpected response: %s, want code=%d", w.Body.String(), tc.code)
			}
			if tc.code == 0 && (len(body.Data) == 0 || string(body.Data) == "null") {
				t.Fatal("missing success data")
			}
			if tc.code != 0 && string(body.Data) != "null" {
				t.Fatal("business failure data must be null")
			}

		})
	}
}

// loginRepo 模拟用户查询结果，覆盖凭据失败与数据库故障的区别。
type loginRepo struct {
	registrationRepo
	user *User
	err  error
}

func (r loginRepo) FindUserByUsername(context.Context, string) (*User, error) {
	return r.user, r.err
}
func TestLoginResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hash := fmt.Sprintf("%x", md5.Sum([]byte("testpass123")))
	for _, tc := range []struct {
		name         string
		body         string
		user         *User
		err          error
		status, code int
	}{
		{"invalid params", `{}`, nil, nil, 200, 4000},
		{"missing user", `{"username":"alice","password":"testpass123"}`, nil, gorm.ErrRecordNotFound, 200, 5000},
		{"wrong password", `{"username":"alice","password":"wrongpass"}`, &User{PasswordHash: hash}, nil, 200, 5000},
		{"database failure", `{"username":"alice","password":"testpass123"}`, nil, errors.New("private database details"), 500, 0},
		{"success", `{"username":"alice","password":"testpass123"}`, &User{ID: 1, Username: "alice", PasswordHash: hash}, nil, 200, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			RegisterRoutes(router.Group("/api/v1"), NewService(loginRepo{user: tc.user, err: tc.err}), slog.New(slog.NewTextHandler(io.Discard, nil)))
			req := httptest.NewRequest(http.MethodPost, "/api/v1/user/login", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.status {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
			if tc.status == 500 {
				if w.Body.String() != "登录失败，请稍后重试" || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/plain") {
					t.Fatalf("unexpected server error: %s", w.Body.String())
				}
				return
			}
			var body struct {
				Code *int `json:"code"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Code == nil || *body.Code != tc.code {
				t.Fatalf("unexpected business response: %s", w.Body.String())
			}
		})
	}
}
