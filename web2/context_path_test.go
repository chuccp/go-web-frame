package web2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinContextPath(t *testing.T) {
	tests := []struct {
		name         string
		contextPath  string
		relativePath string
		expected     string
	}{
		{
			name:         "empty context path",
			contextPath:  "",
			relativePath: "/users",
			expected:     "/users",
		},
		{
			name:         "context path without leading slash",
			contextPath:  "api",
			relativePath: "/users",
			expected:     "/api/users",
		},
		{
			name:         "context path with leading slash",
			contextPath:  "/api",
			relativePath: "/users",
			expected:     "/api/users",
		},
		{
			name:         "context path with trailing slash",
			contextPath:  "/api/",
			relativePath: "/users",
			expected:     "/api/users",
		},
		{
			name:         "root path",
			contextPath:  "/api",
			relativePath: "/",
			expected:     "/api/",
		},
		{
			name:         "relative path without leading slash",
			contextPath:  "/api",
			relativePath: "users",
			expected:     "/api/users",
		},
		{
			name:         "both empty",
			contextPath:  "",
			relativePath: "/",
			expected:     "/",
		},
		{
			name:         "nested path",
			contextPath:  "/api/v1",
			relativePath: "/users/:id",
			expected:     "/api/v1/users/:id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinContextPath(tt.contextPath, tt.relativePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStripContextPath(t *testing.T) {
	tests := []struct {
		name        string
		contextPath string
		path        string
		expected    string
		ok          bool
	}{
		{
			name:        "empty context path matches all",
			contextPath: "",
			path:        "/anything",
			expected:    "/anything",
			ok:          true,
		},
		{
			name:        "exact match",
			contextPath: "/app",
			path:        "/app",
			expected:    "/",
			ok:          true,
		},
		{
			name:        "match with subpath",
			contextPath: "/app",
			path:        "/app/index.html",
			expected:    "/index.html",
			ok:          true,
		},
		{
			name:        "context path with trailing slash",
			contextPath: "/app/",
			path:        "/app/index.html",
			expected:    "/index.html",
			ok:          true,
		},
		{
			name:        "path with trailing slash",
			contextPath: "/app",
			path:        "/app/",
			expected:    "/",
			ok:          true,
		},
		{
			name:        "no match - different prefix",
			contextPath: "/app",
			path:        "/other/file.html",
			expected:    "",
			ok:          false,
		},
		{
			name:        "no match - boundary check",
			contextPath: "/app",
			path:        "/application/file.html",
			expected:    "",
			ok:          false,
		},
		{
			name:        "nested context path",
			contextPath: "/api/v1",
			path:        "/api/v1/users",
			expected:    "/users",
			ok:          true,
		},
		{
			name:        "nested context path no match",
			contextPath: "/api/v1",
			path:        "/api/v2/users",
			expected:    "",
			ok:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := stripContextPath(tt.contextPath, tt.path)
			assert.Equal(t, tt.ok, ok)
			if ok {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
