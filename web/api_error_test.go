package web

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewErrorMessage(t *testing.T) {
	msg := NewErrorMessage(CodeBadRequest, "bad request")
	if msg.Code != CodeBadRequest {
		t.Fatalf("code = %d, want %d", msg.Code, CodeBadRequest)
	}
	if msg.Data != nil {
		t.Fatalf("data = %#v, want nil", msg.Data)
	}
	if msg.Msg != "请求参数不正确" {
		t.Fatalf("msg = %q, want friendly Chinese message", msg.Msg)
	}
	if msg.Type != APIErrorType {
		t.Fatalf("type = %q, want %q", msg.Type, APIErrorType)
	}
}

func TestNewErrorMessageKeepsBusinessDetail(t *testing.T) {
	msg := NewErrorMessage(CodeBadRequest, "不在打卡范围内")
	if msg.Msg != "不在打卡范围内" {
		t.Fatalf("msg = %q, want business detail", msg.Msg)
	}
}

func TestAPIErrorResponseFormat(t *testing.T) {
	server := DefaultServer()
	server.Get("/api/fail", func(*Request) (any, error) {
		return nil, NewBadRequest().WithDetail("数量不能为空")
	})
	server.Get("/api/validation", func(*Request) (any, error) {
		return nil, NewValidationError().WithDetail("数量格式不正确")
	})
	server.Get("/api/message", func(*Request) (any, error) {
		return &Message{Code: CodeBadRequest, Msg: "bad request"}, nil
	})
	server.Get("/api/panic", func(*Request) (any, error) {
		panic("unexpected")
	})

	ts := httptest.NewServer(server.GetHandler())
	t.Cleanup(ts.Close)

	tests := []struct {
		name string
		path string
		body int
		code int
		msg  string
	}{
		{name: "business error", path: "/api/fail", body: CodeBadRequest, code: http.StatusBadRequest, msg: "数量不能为空"},
		{name: "validation code", path: "/api/validation", body: CodeValidationFailed, code: http.StatusBadRequest, msg: "数量格式不正确"},
		{name: "direct message", path: "/api/message", body: CodeBadRequest, code: http.StatusBadRequest, msg: "请求参数不正确"},
		{name: "unknown route", path: "/api/missing", body: CodeNotFound, code: http.StatusNotFound, msg: "请求的资源不存在"},
		{name: "panic", path: "/api/panic", body: CodeInternalError, code: http.StatusInternalServerError, msg: "请求处理失败，请稍后重试"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(ts.URL + tt.path)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.code {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.code)
			}
			var payload map[string]any
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("decode %q: %v", body, err)
			}
			if payload["code"] != float64(tt.body) || payload["msg"] != tt.msg || payload["type"] != APIErrorType || payload["data"] != nil {
				t.Fatalf("unexpected error response: %#v", payload)
			}
		})
	}
}

func TestNonErrorMessageKeepsStatusCode(t *testing.T) {
	server := DefaultServer()
	server.Get("/api/created", func(*Request) (any, error) {
		return DataCode(http.StatusCreated, "created"), nil
	})

	ts := httptest.NewServer(server.GetHandler())
	t.Cleanup(ts.Close)

	resp, err := http.Get(ts.URL + "/api/created")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	var payload Message
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload.Code != http.StatusCreated || payload.Msg != "ok" || payload.Type != "" {
		t.Fatalf("unexpected success response: %#v", payload)
	}
}
