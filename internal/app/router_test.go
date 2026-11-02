// 本测试直接调用 Gin 引擎，不占用真实网络端口。
package app

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// 一起验证健康检查、panic 恢复和日志，确保请求处理链能正常配合。
func TestHealthAndRecovery(t *testing.T) {
	// bytes.Buffer 是内存缓冲区；将日志写到这里，测试就可以读取并检查内容。
	var logs bytes.Buffer
	// 复用生产代码的路由创建函数，测试实际安装的中间件和健康接口。
	router := NewRouter(slog.New(slog.NewJSONHandler(&logs, nil)), nil)
	// 这个故意 panic 的路由仅在当前测试实例中存在，不会暴露在正式服务里。
	router.GET("/test-panic", func(c *gin.Context) { panic("test panic") })

	// 同一个路由实例依次处理正常请求、异常请求和恢复后的正常请求。
	// 表中依次执行正常请求、异常请求、再次正常请求，验证异常不会破坏后续处理。
	for _, tc := range []struct {
		path   string
		status int
		body   string
	}{
		{"/healthz", http.StatusOK, `{"status":"ok"}`},
		{"/test-panic", http.StatusInternalServerError, `{"code":"INTERNAL_ERROR","message":"服务内部错误"}`},
		{"/healthz", http.StatusOK, `{"status":"ok"}`},
	} {
		// Recorder 模拟响应接收端，保存状态码、响应头和响应体供后续断言。
		recorder := httptest.NewRecorder()
		// NewRequest 构造内存中的 GET 请求；ServeHTTP 让它走完整的 Gin 处理链。
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
		// 检查正常请求返回 200、异常请求返回 500。Fatalf 标记失败并停止测试。
		if recorder.Code != tc.status {
			t.Fatalf("%s: status = %d, want %d", tc.path, recorder.Code, tc.status)
		}
		// 将实际响应 got 和预期响应 want 都解码为 Go 数据，避免依赖 JSON 键的排列顺序。
		var got, want any
		if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(tc.body), &want); err != nil {
			t.Fatal(err)
		}
		// DeepEqual 递归比较 map 等复合数据，确认 JSON 内容一致。
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s: body = %s, want %s", tc.path, recorder.Body.String(), tc.body)
		}
		// Content-Type 可以包含 charset 等附加内容，但必须以 application/json 开头。
		if !strings.HasPrefix(recorder.Header().Get("Content-Type"), "application/json") {
			t.Fatal("response must be JSON")
		}
	}
	// 三个请求应各产生一条访问日志，其中只有一个请求产生异常日志。
	var requests, panics int
	// slog 每行输出一个 JSON 对象，通过 Decoder 按顺序读取这些对象。
	decoder := json.NewDecoder(&logs)
	for decoder.More() {
		// map[string]any 的键是字符串，值允许不同类型，用于接收日志字段。
		var entry map[string]any
		if err := decoder.Decode(&entry); err != nil {
			t.Fatal(err)
		}
		// 按日志消息分类，分别检查访问日志和异常日志的关键字段。
		switch entry["msg"] {
		case "HTTP 请求":
			requests++
			if entry["method"] != "GET" || entry["path"] == nil || entry["status"] == nil || entry["duration"] == nil {
				t.Fatalf("missing request log fields: %v", entry)
			}
		case "HTTP 请求异常":
			panics++
			if entry["level"] != "ERROR" {
				t.Fatalf("panic log level = %v", entry["level"])
			}
		}
	}
	// 最终计数也确认异常恢复后的健康请求仍然执行，且访问日志没有遗漏。
	if requests != 3 || panics != 1 {
		t.Fatalf("request logs = %d, panic logs = %d", requests, panics)
	}
}
