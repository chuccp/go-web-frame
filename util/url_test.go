package util

import "testing"

func TestJoinUrl(t *testing.T) {
	tests := []struct {
		name     string
		root     string
		args     []string
		expected string
	}{
		{
			name:     "simple path join",
			root:     "/api",
			args:     []string{"users", "1"},
			expected: "/api/users/1",
		},
		{
			name:     "with trailing slash",
			root:     "/api/",
			args:     []string{"users"},
			expected: "/api/users",
		},
		{
			name:     "with leading slash in args",
			root:     "/api",
			args:     []string{"/users", "/1"},
			expected: "/api/users/1",
		},
		{
			name:     "http URL",
			root:     "http://example.com",
			args:     []string{"api", "users"},
			expected: "http://example.com/api/users",
		},
		{
			name:     "https URL with trailing slash",
			root:     "https://example.com/",
			args:     []string{"/api/", "/users/"},
			expected: "https://example.com/api/users",
		},
		{
			name:     "empty args",
			root:     "/api",
			args:     []string{},
			expected: "/api",
		},
		{
			name:     "empty root",
			root:     "",
			args:     []string{"api", "users"},
			expected: "/api/users",
		},
		{
			name:     "windows backslash path",
			root:     "C:\\static\\voice",
			args:     []string{"58\\file.mp3"},
			expected: "C:/static/voice/58/file.mp3",
		},
		{
			name:     "http URL with windows backslash",
			root:     "http://127.0.0.1:8080/static/voice",
			args:     []string{"58\\5834d6318bf6c645d7fac1d95aa9eea8.mp3"},
			expected: "http://127.0.0.1:8080/static/voice/58/5834d6318bf6c645d7fac1d95aa9eea8.mp3",
		},
		{
			name:     "mixed slashes and backslashes",
			root:     "/static/voice/",
			args:     []string{"58\\file.mp3"},
			expected: "/static/voice/58/file.mp3",
		},
		{
			name:     "double backslash issue",
			root:     "http://127.0.0.1:8080/static/voice/58",
			args:     []string{"\\5834d6318bf6c645d7fac1d95aa9eea8.mp3"},
			expected: "http://127.0.0.1:8080/static/voice/58/5834d6318bf6c645d7fac1d95aa9eea8.mp3",
		},
		{
			name:     "empty parts in args",
			root:     "/api",
			args:     []string{"", "users", ""},
			expected: "/api/users",
		},
		{
			name:     "http URL without path",
			root:     "http://example.com",
			args:     []string{},
			expected: "http://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinUrl(tt.root, tt.args...)
			if result != tt.expected {
				t.Errorf("JoinUrl(%q, %v) = %q, want %q", tt.root, tt.args, result, tt.expected)
			}
		})
	}
}