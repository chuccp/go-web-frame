package util

import (
	"path/filepath"
	"strings"
)

var imageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".webp": true,
	".tiff": true,
	".svg":  true,
	".ico":  true,
}

// IsImageByURL 通过URL路径判断是否为图片
// 参数：rawURL 原始URL字符串（如"https://example.com/photo.jpg?width=100"）
// 返回：bool（是否为图片）、error（URL解析错误）
func IsImagePath(path string) bool {
	// 1. 解析URL，提取完整路径（自动忽略查询参数、锚点）

	// 2. 提取URL路径中的文件扩展名
	pathExt := filepath.Ext(path)

	// 3. 统一转换为小写，兼容大小写扩展名（如.JPG、.Png等）
	lowerExt := strings.ToLower(pathExt)

	// 4. 判断扩展名是否在图片格式集合中
	return imageExtensions[lowerExt]
}
