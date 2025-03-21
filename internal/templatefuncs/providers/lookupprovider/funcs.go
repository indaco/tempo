package lookupprovider

import "strings"

// Lookup retrieves a value from a nested map using a dot-separated key.
func Lookup(data map[string]any, key string) any {
	if key == "" {
		return data
	}

	keys := strings.Split(key, ".")
	var value any = data

	for _, k := range keys {
		if m, ok := value.(map[string]any); ok {
			if v, exists := m[k]; exists {
				value = v
			} else {
				return nil
			}
		} else {
			return nil
		}
	}
	return value
}
