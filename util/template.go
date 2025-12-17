package util

import (
	"bytes"
	"text/template"
)

func ParseTemplate(templateStr string, data map[string]interface{}) (string, error) {
	parse, err := template.New("template").Parse(templateStr)
	if err == nil {
		buffer := new(bytes.Buffer)
		err = parse.Execute(buffer, data)
		if err == nil {
			return buffer.String(), nil
		}
	}
	return "", err
}
