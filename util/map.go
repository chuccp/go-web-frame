package util

func GetMap(key string, value any) map[string]any {
	return map[string]any{
		key: value,
	}
}
func OfMap(key string, value any) map[string]any {
	return GetMap(key, value)
}
