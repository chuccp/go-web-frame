package util

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"emperror.dev/errors"

	"github.com/google/uuid"
	"github.com/spf13/cast"
)

func MD5(data []byte) string {
	hash := md5.New()
	hash.Write(data)
	hashValue := hash.Sum(nil)
	md5Hash := hex.EncodeToString(hashValue)
	return md5Hash
}
func MD5Str(data string) string {
	return MD5([]byte(data))
}

// DefaultChunkSize 常用分块大小：4KB，与大多数操作系统页大小一致
const DefaultChunkSize = 4096 // 4KB

// CalculateMD5 计算字节数组的MD5哈希值（自动分块处理）
func CalculateMD5(data []byte) (string, error) {
	return CalculateMD5WithChunkSize(data, DefaultChunkSize)
}

// CalculateMD5WithChunkSize 支持自定义分块大小的MD5计算
// 适合处理超大字节数组，避免一次性加载全部数据
func CalculateMD5WithChunkSize(data []byte, chunkSize int) (string, error) {
	if chunkSize <= 0 {
		return "", fmt.Errorf("无效的分块大小: %d", chunkSize)
	}

	hasher := md5.New()
	length := len(data)

	for i := 0; i < length; i += chunkSize {
		end := i + chunkSize
		if end > length {
			end = length
		}
		// 分块写入数据
		_, err := hasher.Write(data[i:end])
		if err != nil {
			return "", fmt.Errorf("分块写入失败: %w", err)
		}
	}

	// 转换为16进制字符串
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func SplitAndDeduplicate(text, sep string) []string {
	ps := make([]string, 0)
	if len(text) == 0 {
		return ps
	}
	vs := strings.Split(text, sep)
	m := make(map[string]bool)
	for _, i2 := range vs {
		i2 = strings.TrimSpace(i2)
		if len(i2) == 0 || m[i2] {
			continue
		} else {
			m[i2] = true
			ps = append(ps, i2)
		}
	}
	return ps
}

func SplitPath(path string) []string {
	path = strings.ReplaceAll(path, "\\", "/")
	vs := strings.Split(path, "/")
	ps := make([]string, 0)
	for _, i2 := range vs {
		i2 = strings.TrimSpace(i2)
		if len(i2) == 0 {
			continue
		}
		ps = append(ps, i2)
	}
	return ps
}

func DeduplicateIds(ids string) string {
	vs := strings.Split(ids, ",")
	vvs := RemoveDuplicates(vs)

	return strings.Join(vvs, ",")

}

func RemoveDuplicates(nums []string) []string {
	m := make(map[string]bool)
	var result []string
	for _, num := range nums {
		if _, ok := m[num]; !ok {
			m[num] = true
			result = append(result, num)
		}
	}
	return result
}

func StringToUintIds(ids string) []uint {
	uIds := make([]uint, 0)
	receiveEmailIds := strings.Split(ids, ",")
	for _, i2 := range receiveEmailIds {
		atoi, err := strconv.Atoi(i2)
		if err != nil {
			continue
		}
		uIds = append(uIds, uint(atoi))
	}
	return uIds
}

func IsMatchPath(path, smath string) bool {
	math := ReplaceAllRegex(smath, "\\*[a-zA-Z]+", ".*")
	re := regexp.MustCompile("^" + math)
	fa := re.MatchString(path)
	return fa

}
func ReplaceAllRegex(path, regex, math string) string {
	re := regexp.MustCompile(regex)
	return re.ReplaceAllString(path, math)
}

func ContainsAny(s string, strs ...string) bool {
	for _, str := range strs {
		if strings.Contains(s, str) {
			return true
		}
	}
	return false
}

func ContainsAnyIgnoreCase(s string, strs ...string) bool {
	sLower := strings.ToLower(s)
	for _, str := range strs {
		if strings.Contains(sLower, strings.ToLower(str)) {
			return true
		}
	}
	return false
}
func EqualsAnyIgnoreCase(s string, strs ...string) bool {
	sLower := strings.ToLower(s)
	for _, str := range strs {
		if sLower == strings.ToLower(str) {
			return true
		}
	}
	return false
}
func StartsWithAnyIgnoreCase(s string, prefix ...string) bool {
	sLower := strings.ToLower(s)
	for _, prefix := range prefix {
		if strings.HasPrefix(sLower, strings.ToLower(prefix)) {
			return true
		}
	}
	return false
}

func BoolToString(b bool) string {
	if b {
		return "true"
	} else {
		return "false"
	}
}
func IsNumber(value string) bool {
	return IsMatch(value, "^[0-9]+$")
}
func StringToBool(s string) bool {
	if s == "true" {
		return true
	} else {
		return false
	}
}
func StringToUInt(s string) uint {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return uint(i)
}
func StringToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}
func Trim(value string) string {
	return strings.TrimSpace(value)
}
func GetCachePath(rootPath, filename string) string {
	id := uuid.New().String()
	name := MD5([]byte(id))
	ext := path.Ext(filename)
	if len(ext) > 0 {
		name = name + ext
	}
	return path.Join(rootPath, FormatDate(time.Now()), name)
}
func IsBlank(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}
func IsNotBlank(str string) bool {
	return !IsBlank(str)
}

// DecodeBase64 解码base64字符串为字节数组
func DecodeFileBase64(base64Str string) ([]byte, error) {
	if strings.HasPrefix(base64Str, "data:") {
		base64Str = strings.SplitN(base64Str, ",", 2)[1]
	}
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, errors.New("base64解码失败: " + err.Error())
	}
	return data, nil
}

// 预设字符集
const (
	// Alphanumeric 包含大小写字母和数字
	Alphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// Lowercase 仅包含小写字母
	Lowercase = "abcdefghijklmnopqrstuvwxyz"
	// Uppercase 仅包含大写字母
	Uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// Numeric 仅包含数字
	Numeric = "0123456789"
	// WithSpecialChars 包含字母、数字和常见特殊字符
	WithSpecialChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"
)

// GenerateRandomString 生成指定长度的随机字符串
// length: 字符串长度
// charset: 字符集，如果为空则使用默认的字母数字字符集
func GenerateRandomString(length int, charset string) string {
	// 如果未指定字符集，使用默认的字母数字字符集
	if charset == "" {
		charset = Alphanumeric
	}
	// 初始化随机数生成器
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 创建结果切片
	result := make([]byte, length)

	// 生成随机字符串
	for i := range result {
		result[i] = charset[r.Intn(len(charset))]
	}

	return string(result)
}
func GenerateRandomNum(length int) string {
	return GenerateRandomString(length, Numeric)
}
func GenerateRandomStringByAlphanumeric(length int) string {
	return GenerateRandomString(length, Alphanumeric)
}
func Number2String(number int64, charset string) string {
	// 处理空字符集，直接使用默认常量（避免字符串拷贝）
	if charset == "" {
		charset = Alphanumeric
	}
	base := int64(len(charset))
	if base == 0 { // 防御性处理（虽然常量集不会为空，但避免传入非法charset）
		return "0"
	}
	// 特殊情况：数字为0直接返回"0"
	if number == 0 {
		return "0"
	}
	// 估算结果长度并预分配切片（减少扩容次数）
	// 原理：number < base^n → n > log_base(number) → 取对数估算长度
	length := 0
	n := number
	for n > 0 {
		n /= base
		length++
	}
	result := make([]byte, 0, length)
	// 反向收集字符（尾部追加，O(1)操作）
	for number > 0 {
		remainder := number % base
		result = append(result, charset[remainder]) // 尾部追加，无拷贝
		number = number / base
	}
	// 反转切片（得到正确顺序）
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}

func Index2Str(index int, charset string) string {
	if charset == "" {
		charset = Alphanumeric
	}
	return string(charset[index])
}
func JoinValues(values ...any) string {
	b := new(bytes.Buffer)
	for _, v := range values {
		b.WriteString(cast.ToString(v))
	}
	return b.String()
}

func ArrayIntContains(arr []int, str int) bool {
	for _, item := range arr {
		if item == str {
			return true
		}
	}
	return false
}

// SubStringAndPadSpace 截取字符串到指定长度，若原字符串长度不足则右侧补空格
// 参数:
//
//	value - 原始字符串
//	length - 目标字符串长度
//
// 返回值:
//
//	截取/填充后的指定长度字符串
func SubStringAndPadSpace(value string, length int) string {
	value = Trim(value)
	strLen := len(value)
	if strLen >= length {
		return value[:length]
	}
	padding := strings.Repeat(" ", length-strLen)
	return value + padding
}
