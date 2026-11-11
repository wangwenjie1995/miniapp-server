package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"io"
	"log/slog"
	"miniapp-server/internal/config"
	"miniapp-server/internal/database"
	"miniapp-server/migrations"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"miniapp-server/internal/home"
)

// homeResponse 表示最外层业务响应，RawMessage 保留 data 的 JSON，供每个测试按需解析。
type homeResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// fetchHome 通过真实路由走完中间件和 handler，不直接调用业务函数。
func fetchHome(t *testing.T, router *gin.Engine, path string, status int) homeResponse {
	t.Helper()
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
	if recorder.Code != status {
		t.Fatalf("%s: status=%d, want=%d, body=%s", path, recorder.Code, status, recorder.Body.String())
	}
	if status == http.StatusInternalServerError {
		if recorder.Body.String() != "首页数据暂时无法获取" || recorder.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
			t.Fatalf("unexpected server error: %s", recorder.Body.String())
		}
		return homeResponse{Message: recorder.Body.String()}
	}
	var result homeResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Message == "" || (status == http.StatusOK && len(result.Data) == 0) {
		t.Fatalf("%s: incomplete response: %s", path, recorder.Body.String())
	}
	if (string(result.Data) == "null" && result.Code != 4000) || (string(result.Data) != "null" && result.Code != 0) {
		t.Fatalf("%s: success code=%d", path, result.Code)
	}
	return result
}

// TestHomeSections 验证聚合接口与各独立接口一致，且没有返回暂不启用的 AI 入口。
func TestHomeSections(t *testing.T) {
	router := newHomeRouter(t)
	result := fetchHome(t, router, "/api/v1/home", http.StatusOK)
	var overview home.Overview
	if err := json.Unmarshal(result.Data, &overview); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		path string
		want any
	}{
		{"featured", overview.Featured},
		{"tools", overview.Tools},
		{"banners", overview.Banners},
		{"flash", overview.FlashHeadlines},
	} {
		t.Run(tc.path, func(t *testing.T) {
			section := fetchHome(t, router, "/api/v1/home/"+tc.path, http.StatusOK)
			expected, err := json.Marshal(tc.want)
			if err != nil {
				t.Fatal(err)
			}
			var got, want any
			if err := json.Unmarshal(section.Data, &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(expected, &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("section differs from overview: %s", section.Data)
			}
			if items, ok := got.([]any); !ok || len(items) == 0 {
				t.Fatalf("expected nonempty list: %s", section.Data)
			}
		})
	}
	for _, tool := range overview.Tools {
		if tool.Key == "ai" {
			t.Fatal("AI entry must be excluded")
		}
	}
}

// TestHomeNewsPagination 检查全部栏目、跨页去重、真实总数、尾页和超范围空页。
func TestHomeNewsPagination(t *testing.T) {
	router := newHomeRouter(t)
	read := func(path string) home.NewsPage {
		t.Helper()
		result := fetchHome(t, router, path, http.StatusOK)
		var page home.NewsPage
		if err := json.Unmarshal(result.Data, &page); err != nil {
			t.Fatal(err)
		}
		return page
	}
	defaults := read("/api/v1/home/news")
	if defaults.Page != 1 || defaults.Size != 20 || defaults.Total == 0 || int64(len(defaults.Data)) != defaults.Total {
		t.Fatalf("unexpected defaults: %+v", defaults)
	}
	for _, tab := range []string{"discover", "news", "flash", "watchlist", "journey"} {
		t.Run(tab, func(t *testing.T) {
			base := "/api/v1/home/news?tab=" + tab
			all := read(base)
			first := read(base + "&page=1&size=3")
			second := read(base + "&page=2&size=3")
			last := read(base + "&page=3&size=3")
			combined := append(append(first.Data, second.Data...), last.Data...)
			if !reflect.DeepEqual(combined, all.Data) || all.Total != int64(len(combined)) {
				t.Fatalf("pagination must reproduce full list: %+v", combined)
			}
			seen := make(map[string]bool)
			for _, item := range combined {
				if item.ID == "" || item.Title == "" || seen[item.ID] {
					t.Fatalf("missing data or repeated ID: %+v", item)
				}
				seen[item.ID] = true
			}
			again := read(base + "&page=1&size=3")
			if !reflect.DeepEqual(first, again) {
				t.Fatal("repeating the same request must return stable IDs and order")
			}
			for _, pageNumber := range []string{"4", "9223372036854775807"} {
				empty := read(base + "&page=" + pageNumber + "&size=3")
				if empty.Data == nil || len(empty.Data) != 0 || empty.Total != all.Total {
					t.Fatalf("out-of-range page must return [] and actual total: %+v", empty)
				}
			}
		})
	}
}

// TestHomeNewsInvalidQuery 确认非法输入明确失败，不会被悄悄当作成功数据。
func TestHomeNewsInvalidQuery(t *testing.T) {
	router := NewRouter(RouterDeps{Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)), HomeRepository: emptyRepository{}})
	for _, query := range []string{
		"tab=unknown", "page=0", "page=-1", "page=abc", "page=1.5",
		"size=0", "size=-1", "size=101", "size=abc",
		"page=99999999999999999999999999999",
	} {
		t.Run(query, func(t *testing.T) {
			result := fetchHome(t, router, "/api/v1/home/news?"+query, http.StatusOK)
			if result.Code != 4000 || result.Message == "" || string(result.Data) != "null" {
				t.Fatalf("unexpected error body: %+v", result)
			}
		})
	}
}

// newHomeDB 仅在显式启用 MYSQL_INTEGRATION=1 时创建独立测试库。
// 测试不会清空用户的业务库；只删除本次自动生成且由测试创建的库。
func newHomeDB(t *testing.T) *gorm.DB {
	t.Helper()
	if os.Getenv("MYSQL_INTEGRATION") != "1" {
		t.Skip("设置 MYSQL_INTEGRATION=1 运行真实 MySQL 首页测试")
	}
	local := filepath.Join("..", "..", ".env.local")
	if values, err := godotenv.Read(local); err == nil {
		for key, value := range values {
			if _, exists := os.LookupEnv(key); !exists {
				t.Setenv(key, value)
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	// 不选择业务库，用同一账号创建本次测试专用的空数据库。
	adminConfig := cfg.DB
	adminConfig.Name = ""
	_, admin, err := database.Open(ctx, adminConfig)
	if err != nil {
		t.Fatal(err)
	}
	name := fmt.Sprintf("miniapp_test_%d", time.Now().UnixNano())
	if _, err := admin.ExecContext(ctx, "CREATE DATABASE `"+name+"` CHARACTER SET utf8mb4"); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.ExecContext(context.Background(), "DROP DATABASE `"+name+"`"); err != nil {
			t.Error(err)
		}
		admin.Close()
	})
	cfg.DB.Name = name
	db, pool, err := database.Open(ctx, cfg.DB)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Close() })
	// 连续执行两次，验证迁移和种子数据能安全重跑。
	for range 2 {
		if err := migrations.Apply(ctx, pool); err != nil {
			t.Fatal(err)
		}
		if err := home.Seed(ctx, db); err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func newHomeRouter(t *testing.T) *gin.Engine {
	t.Helper()
	return NewRouter(RouterDeps{Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)), HomeRepository: home.NewRepository(newHomeDB(t))})
}

// emptyRepository 用于不依赖 MySQL 的接口测试；正常服务不会使用这个测试对象。
// err 非空时模拟存储失败，否则返回空数据，检验“没有内容”和“查询失败”被正确区分。
type emptyRepository struct{ err error }

func (r emptyRepository) Featured(context.Context) ([]home.Featured, error) {
	return []home.Featured{}, r.err
}
func (r emptyRepository) Tools(context.Context) ([]home.Tool, error) { return []home.Tool{}, r.err }
func (r emptyRepository) Banners(context.Context) ([]home.Banner, error) {
	return []home.Banner{}, r.err
}
func (r emptyRepository) Flash(context.Context) ([]string, error) { return []string{}, r.err }
func (r emptyRepository) News(_ context.Context, q home.NewsQuery) (home.NewsPage, error) {
	return home.NewsPage{Data: []home.NewsItem{}, Page: q.Page, Size: q.Size}, r.err
}

func TestHomeEmptyAndDatabaseFailure(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
	}{
		{"empty", nil, http.StatusOK},
		{"failure", errors.New("private database details"), http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			router := NewRouter(RouterDeps{Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)), HomeRepository: emptyRepository{err: tc.err}})
			for _, suffix := range []string{"", "/featured", "/tools", "/banners", "/flash", "/news"} {
				body := fetchHome(t, router, "/api/v1/home"+suffix, tc.status)
				if tc.err != nil && (body.Message != "首页数据暂时无法获取" || len(body.Data) != 0) {
					t.Fatalf("database errors must not leak or become success: %+v", body)
				}
				if tc.err == nil && suffix != "" && suffix != "/news" && string(body.Data) != "[]" {
					t.Fatalf("empty list must be []: %s", body.Data)
				}
			}
		})
	}
}

// TestHomeDatabaseChanges 验证改库后接口立即变化、未发布数据被过滤、取消请求停止查询。
func TestHomeDatabaseChanges(t *testing.T) {
	db := newHomeDB(t)
	ctx := context.Background()
	repo := home.NewRepository(db)
	if _, err := gorm.G[home.Article](db).Where("id = ?", "news-0").Update(ctx, "title", "数据库修改后的标题"); err != nil {
		t.Fatal(err)
	}
	// 再次初始化不能覆盖已修改的数据。
	if err := home.Seed(ctx, db); err != nil {
		t.Fatal(err)
	}
	page, err := repo.News(ctx, home.NewsQuery{Tab: "news", Page: 1, Size: 20})
	if err != nil || page.Data[0].Title != "数据库修改后的标题" {
		t.Fatalf("database change not visible: %+v, %v", page, err)
	}
	if _, err := gorm.G[home.Article](db).Where("id = ?", "news-0").Update(ctx, "published", false); err != nil {
		t.Fatal(err)
	}
	if _, err := gorm.G[home.Article](db).Where("id = ?", "news-1").Update(ctx, "published_at", time.Now().UTC().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	page, err = repo.News(ctx, home.NewsQuery{Tab: "news", Page: 1, Size: 20})
	if err != nil || page.Total != 6 {
		t.Fatalf("publication filter failed: %+v, %v", page, err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := repo.News(cancelled, home.NewsQuery{Tab: "news", Page: 1, Size: 20}); err == nil {
		t.Fatal("cancelled request must fail")
	}
}
