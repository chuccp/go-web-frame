package util

func GetMap(key string, value any) map[string]any {
	return map[string]any{
		key: value,
	}
}
func OfMap(key string, value any) map[string]any {
	return GetMap(key, value)
}
func OfMap2(key string, value any, key2 string, value2 any) map[string]any {
	return map[string]any{
		key:  value,
		key2: value2,
	}
}
