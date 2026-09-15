package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const spaIndexHTML = "<html><body>spa-index</body></html>"

// newSPATestServer 起一个"API 挂在 ContextPath 下、前端产物放在静态目录根路径"的服务端。
func newSPATestServer(t *testing.T, contextPath string) *httptest.Server {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(spaIndexHTML), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte("console.log(1)"), 0o644); err != nil {
		t.Fatal(err)
	}

	server := DefaultServer()
	server.serverConfig.ContextPath = contextPath
	server.serverConfig.Locations = []string{dir}
	server.serverConfig.Page404 = "/index.html"
	server.Get("/ping", func(r *Request) (any, error) { return "pong", nil })

	ts := httptest.NewServer(server.GetHandler())
	t.Cleanup(ts.Close)
	return ts
}

// getNoRedirect 发请求但不跟随重定向，便于断言 301/302 本身。
func getNoRedirect(t *testing.T, ts *httptest.Server, path, accept string) (*http.Response, string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, ts.URL+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	client := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp, string(body)
}

// ContextPath 只作用于路由，静态资源（SPA 产物）从根路径访问时不该被前缀判断挡掉。
// 回归：此前 stripContextPath 返回 false 就直接 return，带 ContextPath 的服务端
// 无法在根路径托管前端（API 在 /api、界面在 /）。
func TestNoRouteServesStaticFromRootWithContextPath(t *testing.T) {
	ts := newSPATestServer(t, "/api")

	// ContextPath 下的 API 照常工作
	resp, body := getNoRedirect(t, ts, "/api/ping", "")
	if resp.StatusCode != http.StatusOK || !strings.Contains(body, "pong") {
		t.Errorf("GET /api/ping = %d %q, want 200 且含 pong", resp.StatusCode, body)
	}

	// 根路径的 SPA 页面
	resp, body = getNoRedirect(t, ts, "/", "text/html")
	if resp.StatusCode != http.StatusOK || !strings.Contains(body, spaIndexHTML) {
		t.Errorf("GET / = %d %q, want 200 且返回 index.html", resp.StatusCode, body)
	}

	// 根路径的静态资源
	resp, body = getNoRedirect(t, ts, "/app.js", "")
	if resp.StatusCode != http.StatusOK || !strings.Contains(body, "console.log") {
		t.Errorf("GET /app.js = %d %q, want 200 且返回 app.js", resp.StatusCode, body)
	}
}

// 深链接（SPA 路由）应拿到 404 页内容本身，且不能被 301 送回站点根目录。
// 回归：c.FileFromFS 会把请求路径改写成文件名，http.FileServer 见到 /index.html
// 结尾就 301 到 "./"，浏览器被送回根目录、深链接丢失。
func TestNoRoutePage404DoesNotRedirect(t *testing.T) {
	// 无 ContextPath 时能单独验到 301 那个问题；带 ContextPath 时是用户实际场景。
	for _, cp := range []string{"", "/api"} {
		name := cp
		if name == "" {
			name = "无ContextPath"
		}
		t.Run(name, func(t *testing.T) {
			ts := newSPATestServer(t, cp)

			resp, body := getNoRedirect(t, ts, "/deep/link", "text/html")
			if resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusFound {
				t.Fatalf("GET /deep/link 被重定向到 %q，深链接会丢失", resp.Header.Get("Location"))
			}
			if resp.StatusCode != http.StatusOK || !strings.Contains(body, spaIndexHTML) {
				t.Errorf("GET /deep/link = %d %q, want 200 且返回 SPA 页面", resp.StatusCode, body)
			}
			if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "html") {
				t.Errorf("Content-Type = %q, want text/html", ct)
			}

			// 非 html 请求不该落到 SPA 页面
			resp, body = getNoRedirect(t, ts, "/deep/link", "application/json")
			if resp.StatusCode != http.StatusNotFound {
				t.Errorf("GET /deep/link (Accept: application/json) = %d %q, want 404", resp.StatusCode, body)
			}
		})
	}
}
