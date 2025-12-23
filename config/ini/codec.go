package ini

import (
	"bytes"
	"fmt"

	"github.com/spf13/viper"
	"gopkg.in/ini.v1"
)

// Codec implements the encoding.Encoder and encoding.Decoder interfaces for INI encoding.
// It allows Viper (v1.20+) to read/write .ini files again.
type Codec struct{}

// Encode 将 map[string]any 编码为 INI 格式的 []byte
func (c Codec) Encode(v map[string]any) ([]byte, error) {
	// 创建一个空的 ini.File
	file := ini.Empty()

	// Viper 的 map 结构通常是扁平化的点分隔键（如 "database.host"）
	// 我们需要将其还原为 section + key 的结构
	for flatKey, value := range v {
		// 处理无 section 的键（直接在根部）
		sectionName := "DEFAULT"
		keyName := flatKey

		if dotIdx := bytes.IndexByte([]byte(flatKey), '.'); dotIdx != -1 {
			sectionName = flatKey[:dotIdx]
			keyName = flatKey[dotIdx+1:]
		}

		sec, err := file.NewSection(sectionName)
		if err != nil {
			return nil, err
		}

		// 根据值的类型设置字符串（ini.v1 只支持字符串）
		strVal := fmt.Sprintf("%v", value)
		sec.NewKey(keyName, strVal)
	}

	// 将 ini.File 写入 buffer
	var buf bytes.Buffer
	_, err := file.WriteTo(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Decode 将 INI 格式的 []byte 解码到 map[string]any 中
func (c Codec) Decode(b []byte, v map[string]any) error {
	// 使用 ini.v1 加载字节数据
	file, err := ini.Load(b)
	if err != nil {
		return err
	}
	// 遍历所有 section 和 key，展平为 Viper 喜欢的 "section.key" 格式
	for _, sec := range file.Sections() {
		sectionName := sec.Name()
		for _, key := range sec.Keys() {
			keyName := key.Name()

			var fullKey string
			if sectionName == "DEFAULT" {
				fullKey = keyName // 无 section 的键直接作为顶级键
			} else {
				fullKey = sectionName + "." + keyName
			}
			v[fullKey] = key.Value()
		}
	}

	return nil
}

func NewCodec() viper.Codec {
	return &Codec{}
}
