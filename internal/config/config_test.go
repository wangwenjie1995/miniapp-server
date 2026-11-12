// _test.go 文件仅在 go test 时用于测试，不参与业务服务的启动流程。
package config

import (
	"os"
	"testing"
)

// Test 开头、接收 *testing.T 的函数会被 go test 自动执行。
// 这个测试验证完全未设置 HTTP_ADDR 时使用默认值 :8080。
func TestDefaultAddress(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-only-secret-32-bytes-long-000")
	t.Setenv("JWT_DURATION", "2h")
	// Setenv 会登记测试结束后的恢复操作，避免破坏运行测试前的环境变量。
	t.Setenv("HTTP_ADDR", "")
	// 删除环境变量，模拟“未设置”；它与显式设置为空字符串不同。
	if err := os.Unsetenv("HTTP_ADDR"); err != nil {
		t.Fatal(err)
	}
	// 调用真实配置函数；t.Fatal 或 t.Fatalf 会标记失败并停止当前测试。
	cfg, err := Load()
	if err != nil || cfg.HTTPAddr != ":8080" {
		t.Fatalf("Load() = %+v, %v", cfg, err)
	}
}

// 表驱动测试：把多组输入及预期行为写成表，再逐项验证。
func TestAddressConfiguration(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-only-secret-32-bytes-long-000")
	t.Setenv("JWT_DURATION", "2h")
	// range 遍历切片，_ 忽略序号，tc 是当前测试数据，valid 表示是否预期通过。
	for _, tc := range []struct {
		addr  string
		valid bool
	}{
		{":8081", true},
		{"127.0.0.1:8081", true},
		{"[::1]:8081", true},
		{"", false},
		{"8080", false},
		{":abc", false},
		{":0", false},
		{":65536", false},
	} {
		// 每个地址作为一个子测试名称，出错时可以直接定位失败的输入。
		t.Run(tc.addr, func(t *testing.T) {
			t.Setenv("HTTP_ADDR", tc.addr)
			// 调用真实配置函数；t.Fatal 或 t.Fatalf 会标记失败并停止当前测试。
			cfg, err := Load()
			// err == nil 表示实际成功；如果与预期 valid 不一致，则测试失败。
			if (err == nil) != tc.valid {
				t.Fatalf("Load() = %+v, %v", cfg, err)
			}
			// 合法输入还应保留原始配置，不能悄悄换成默认值。
			if tc.valid && cfg.HTTPAddr != tc.addr {
				t.Fatalf("address = %q, want %q", cfg.HTTPAddr, tc.addr)
			}
		})
	}
}

// 数据库配置校验不需要连接数据库，验证显式错误不会被默认值覆盖。
func TestDatabaseConfiguration(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-only-secret-32-bytes-long-000")
	t.Setenv("JWT_DURATION", "2h")
	for _, tc := range []struct{ key, value string }{
		{"DB_PORT", "abc"}, {"DB_PORT", "0"}, {"DB_PORT", "65536"},
		{"DB_HOST", ""}, {"DB_NAME", ""}, {"DB_USER", ""},
	} {
		t.Run(tc.key+tc.value, func(t *testing.T) {
			t.Setenv(tc.key, tc.value)
			if _, err := Load(); err == nil {
				t.Fatal("invalid database configuration must fail")
			}
		})
	}
}

func TestJWTConfiguration(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":8080")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_NAME", "test")
	t.Setenv("DB_USER", "test")
	for _, tc := range []struct {
		name, secret, duration string
		valid                  bool
	}{
		{"valid", "test-secret-at-least-32-bytes-long", "2h", true},
		{"empty secret", "", "2h", false},
		{"short secret", "short", "2h", false},
		{"empty duration", "test-secret-at-least-32-bytes-long", "", false},
		{"invalid duration", "test-secret-at-least-32-bytes-long", "abc", false},
		{"zero duration", "test-secret-at-least-32-bytes-long", "0s", false},
		{"negative duration", "test-secret-at-least-32-bytes-long", "-1h", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("JWT_SECRET", tc.secret)
			t.Setenv("JWT_DURATION", tc.duration)
			_, err := Load()
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v err=%v", tc.valid, err)
			}
		})
	}
}
