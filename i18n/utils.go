package i18n

func toFlatMap(nestedMap map[string]any) map[string]any {
	flatMap := make(map[string]any)
	flatten("", nestedMap, flatMap)
	return flatMap
}

func flatten(prefix string, nestedMap map[string]any, flatMap map[string]any) {
	for key, value := range nestedMap {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		if subMap, ok := value.(map[string]any); ok {
			flatten(fullKey, subMap, flatMap)
		} else {
			flatMap[fullKey] = value
		}
	}
}
