package util

import "regexp"

func IsMatch(value, regx string) bool {
	// 1. 编译正则表达式（处理正则语法错误）
	re, err := regexp.Compile(regx)
	if err != nil {
		// 正则表达式无效时，直接返回 false（也可根据需求改为返回错误，这里优先保证函数可用性）
		return false
	}

	// 2. 检查完全匹配（若需部分匹配，可改为 re.MatchString(value)）
	// 完全匹配：value 必须整个字符串符合正则，不能是部分包含
	return re.MatchString(value)
}
