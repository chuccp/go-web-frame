package util

import (
	"testing"
	"unicode/utf8"
)

func TestRemovePunctuation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "中文标点符号",
			input:    "你好，世界！",
			expected: "你好世界",
		},
		{
			name:     "英文标点符号",
			input:    "Hello, World!",
			expected: "HelloWorld",
		},
		{
			name:     "混合标点符号",
			input:    "Hello，World！你好。test",
			expected: "HelloWorld你好test",
		},
		{
			name:     "无标点符号",
			input:    "HelloWorld你好世界",
			expected: "HelloWorld你好世界",
		},
		{
			name:     "包含空格",
			input:    "Hello World",
			expected: "HelloWorld",
		},
		{
			name:     "包含数字",
			input:    "Test123！Test",
			expected: "Test123Test",
		},
		{
			name:     "空字符串",
			input:    "",
			expected: "",
		},
		{
			name:     "特殊符号",
			input:    "a@b#c$d%e^f&g*h(i)j",
			expected: "abcdefghij",
		},
		{
			name:     "问号和感叹号",
			input:    "什么？真的吗！",
			expected: "什么真的吗",
		},
		{
			name:     "引号和书名号",
			input:    "他说\u201c你好\u201d\u300a书名\u300b",
			expected: "他说你好书名",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemovePunctuation(tt.input)
			if result != tt.expected {
				t.Errorf("RemovePunctuation(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEqualsAnyIgnorePunctuation(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		strs     []string
		expected bool
	}{
		{
			name:     "中文标点匹配",
			s:        "你好，世界！",
			strs:     []string{"你好世界"},
			expected: true,
		},
		{
			name:     "英文标点匹配",
			s:        "Hello, World!",
			strs:     []string{"HelloWorld"},
			expected: true,
		},
		{
			name:     "不匹配",
			s:        "你好世界",
			strs:     []string{"HelloWorld"},
			expected: false,
		},
		{
			name:     "多个候选值匹配",
			s:        "测试！",
			strs:     []string{"abc", "测试", "def"},
			expected: true,
		},
		{
			name:     "空字符串匹配",
			s:        "",
			strs:     []string{""},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EqualsAnyIgnorePunctuation(tt.s, tt.strs...)
			if result != tt.expected {
				t.Errorf("EqualsAnyIgnorePunctuation(%q, %v) = %v, want %v", tt.s, tt.strs, result, tt.expected)
			}
		})
	}
}

func TestEqualsAnyIgnorePunctuationAndCase(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		strs     []string
		expected bool
	}{
		{
			name:     "忽略大小写和标点",
			s:        "Hello, World!",
			strs:     []string{"helloworld"},
			expected: true,
		},
		{
			name:     "中文标点忽略大小写",
			s:        "你好，世界！",
			strs:     []string{"你好世界"},
			expected: true,
		},
		{
			name:     "混合大小写",
			s:        "TEST！",
			strs:     []string{"test"},
			expected: true,
		},
		{
			name:     "不匹配",
			s:        "abc",
			strs:     []string{"def"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EqualsAnyIgnorePunctuationAndCase(tt.s, tt.strs...)
			if result != tt.expected {
				t.Errorf("EqualsAnyIgnorePunctuationAndCase(%q, %v) = %v, want %v", tt.s, tt.strs, result, tt.expected)
			}
		})
	}
}
func TestSubStringMaxLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		maxLength int
		expected  string
	}{
		{name: "汉字不截断", value: "中国人民万岁", maxLength: 2, expected: "中国"},
		{name: "汉字正好", value: "中国人民", maxLength: 4, expected: "中国人民"},
		{name: "汉字不足", value: "中国", maxLength: 5, expected: "中国"},
		{name: "中英混合", value: "ab中国cd", maxLength: 4, expected: "ab中国"},
		{name: "emoji 不截断", value: "a😀b😀c", maxLength: 3, expected: "a😀b"},
		{name: "纯 ASCII", value: "abcdef", maxLength: 3, expected: "abc"},
		{name: "长度 0", value: "中国", maxLength: 0, expected: ""},
		{name: "带空格会被 Trim", value: "  中国  ", maxLength: 2, expected: "中国"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SubStringMaxLength(tt.value, tt.maxLength)
			if result != tt.expected {
				t.Errorf("SubStringMaxLength(%q, %d) = %q, want %q", tt.value, tt.maxLength, result, tt.expected)
			}
		})
	}
}

func TestSubStringLastMaxLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		maxLength int
		expected  string
	}{
		{name: "汉字不截断", value: "中国人民万岁", maxLength: 2, expected: "万岁"},
		{name: "汉字正好", value: "中国人民", maxLength: 4, expected: "中国人民"},
		{name: "汉字不足", value: "中国", maxLength: 5, expected: "中国"},
		{name: "中英混合", value: "ab中国cd", maxLength: 4, expected: "中国cd"},
		{name: "emoji 不截断", value: "a😀b😀c", maxLength: 3, expected: "b😀c"},
		{name: "纯 ASCII", value: "abcdef", maxLength: 3, expected: "def"},
		{name: "长度 0", value: "中国", maxLength: 0, expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SubStringLastMaxLength(tt.value, tt.maxLength)
			if result != tt.expected {
				t.Errorf("SubStringLastMaxLength(%q, %d) = %q, want %q", tt.value, tt.maxLength, result, tt.expected)
			}
		})
	}
}

// 截断后的结果必须是合法 UTF-8（按字节截断会产生乱码）
func TestSubStringMaxLength_ValidUTF8(t *testing.T) {
	value := "中文字符串测试"
	for i := 0; i <= len([]rune(value))+1; i++ {
		if result := SubStringMaxLength(value, i); !utf8.ValidString(result) {
			t.Errorf("SubStringMaxLength(%q, %d) = %q, 不是合法 UTF-8", value, i, result)
		}
		if result := SubStringLastMaxLength(value, i); !utf8.ValidString(result) {
			t.Errorf("SubStringLastMaxLength(%q, %d) = %q, 不是合法 UTF-8", value, i, result)
		}
	}
}
