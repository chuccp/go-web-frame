package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
)

func BenchmarkHandles_Handle(b *testing.B) {
	handles := NewHandles()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handles.Handle(http.MethodGet, "/route", func(req *Request) (any, error) {
			return "ok", nil
		})
	}
}

func BenchmarkRouteTree_Lookup(b *testing.B) {
	handles := NewHandles()
	for i := 0; i < 1000; i++ {
		handles.Handle(http.MethodGet, "/route/%d", func(req *Request) (any, error) {
			return "ok", nil
		})
	}
	handlerMeta := handles.HandlerMeta(http.MethodGet, "/route/500")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handlerMeta
	}
}

func BenchmarkRouteTree_Has(b *testing.B) {
	handles := NewHandles()
	for i := 0; i < 1000; i++ {
		handles.Handle(http.MethodGet, "/route/%d", func(req *Request) (any, error) {
			return "ok", nil
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handles.HasHandler(http.MethodGet, "/route/500")
	}
}

func BenchmarkRequest_Query(b *testing.B) {
	queryParams := url.Values{}
	queryParams.Set("key", "value")
	c, _ := createTestContext("GET", "/test", queryParams, nil, nil)
	req := newTestRequest(c, NewHandlerMeta())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = req.Query("key")
	}
}

func BenchmarkRequest_Param(b *testing.B) {
	c, _ := createTestContext("GET", "/users/123", url.Values{}, nil, nil)
	c.Params = []gin.Param{{Key: "id", Value: "123"}}
	req := newTestRequest(c, NewHandlerMeta())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = req.Param("id")
	}
}

func BenchmarkRequest_BindJSON(b *testing.B) {
	body, _ := json.Marshal(map[string]any{"name": "John", "age": 30})
	headers := map[string]string{"Content-Type": "application/json"}
	c, _ := createTestContext("POST", "/test", url.Values{}, body, headers)
	req := newTestRequest(c, NewHandlerMeta())

	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var user User
		_ = req.BindJSON(&user)
	}
}

func BenchmarkJoinContextPath(b *testing.B) {
	b.Run("empty context path", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = joinContextPath("", "/users")
		}
	})
	b.Run("with context path", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = joinContextPath("/api/v1", "/users/:id")
		}
	})
	b.Run("with trailing slash", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = joinContextPath("/api/", "/users")
		}
	})
}

func BenchmarkMessage_JsonMarshal(b *testing.B) {
	msg := Data(map[string]any{
		"id":    123,
		"name":  "test",
		"items": []string{"a", "b", "c"},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(msg)
	}
}

func BenchmarkHttpServer_Handle(b *testing.B) {
	gin.SetMode(gin.TestMode)

	serverConfig := &ServerConfig{Port: 0}
	handles := NewHandles()
	handles.Handle(http.MethodGet, "/hello", func(req *Request) (any, error) {
		return map[string]string{"message": "Hello, World!"}, nil
	})

	server := NewHttpServer(serverConfig, NewCertManager())
	server.AddHandle(NewHandlerConfig(nil, handles, serverConfig))
	server.Handle()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/hello", nil)
		server.Engine().ServeHTTP(w, req)
	}
}

func BenchmarkErrorCode_ToMessage(b *testing.B) {
	err := NewErrorCode(CodeBadRequest, "invalid input")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.ToMessage()
	}
}

func BenchmarkNewBadRequest(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewBadRequest("test error")
	}
}
