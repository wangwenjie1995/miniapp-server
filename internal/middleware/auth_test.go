package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetUserIDFromContext(t *testing.T) {
	for _, tc := range []struct {
		name   string
		set    bool
		value  any
		wantID uint64
		wantOK bool
	}{
		{name: "missing"},
		{name: "wrong type", set: true, value: "42"},
		{name: "zero", set: true, value: uint64(0)},
		{name: "valid", set: true, value: uint64(42), wantID: 42, wantOK: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			if tc.set {
				c.Set(userIDKey, tc.value)
			}
			id, ok := GetUserIDFromContext(c)
			if id != tc.wantID || ok != tc.wantOK {
				t.Fatalf("GetUserIDFromContext() = %d, %v; want %d, %v", id, ok, tc.wantID, tc.wantOK)
			}
		})
	}
}
