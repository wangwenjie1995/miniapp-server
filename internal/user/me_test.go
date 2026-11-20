package user_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"miniapp-server/internal/auth"
	"miniapp-server/internal/user"

	"gorm.io/gorm"
)

type currentUserRepo struct {
	registrationRepo
	gotID uint64
	err   error
}

func (r *currentUserRepo) FindUserByID(_ context.Context, id uint64) (*user.User, error) {
	r.gotID = id
	if r.err != nil {
		return nil, r.err
	}
	return &user.User{ID: id, Username: "alice", PasswordHash: "private-hash"}, nil
}

func TestGetCurrentUser(t *testing.T) {
	tokens := auth.NewTokenManager([]byte("test-only-secret-32-bytes-long!!"), time.Hour)
	token, _, err := tokens.Generate(42)
	if err != nil {
		t.Fatal(err)
	}
	expired, _, err := auth.NewTokenManager([]byte("test-only-secret-32-bytes-long!!"), -time.Hour).Generate(42)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, header string
		repoErr      error
		status       int
		queried      bool
	}{
		{"missing token", "", nil, http.StatusUnauthorized, false},
		{"invalid token", "Bearer invalid", nil, http.StatusUnauthorized, false},
		{"expired token", "Bearer " + expired, nil, http.StatusUnauthorized, false},
		{"user deleted", "Bearer " + token, gorm.ErrRecordNotFound, http.StatusUnauthorized, true},
		{"database failure", "Bearer " + token, errors.New("private database details"), http.StatusInternalServerError, true},
		{"success", "Bearer " + token, nil, http.StatusOK, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &currentUserRepo{err: tc.repoErr}
			router := newUserTestRouter(repo, tokens)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/user/me", strings.NewReader(`{"id":99}`))
			req.Header.Set("Authorization", tc.header)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.status {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
			if tc.queried && repo.gotID != 42 {
				t.Fatalf("queried ID=%d, want token subject 42", repo.gotID)
			}
			if !tc.queried && repo.gotID != 0 {
				t.Fatal("unauthenticated request queried repository")
			}
			if strings.Contains(w.Body.String(), "private") {
				t.Fatal("internal data leaked")
			}
			if tc.status == http.StatusOK {
				var response struct {
					Data map[string]any `json:"data"`
				}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatal(err)
				}
				if len(response.Data) != 4 || response.Data["ID"] != float64(42) || response.Data["Username"] != "alice" || response.Data["CreatedAt"] == nil || response.Data["UpdatedAt"] == nil {
					t.Fatalf("unexpected data: %s", w.Body.String())
				}
			}
		})
	}
}
