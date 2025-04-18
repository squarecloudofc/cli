package i18n

import "strings"

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

func toNestedMap(flatMap map[string]any) map[string]any {
	nestedMap := make(map[string]any)

	for key, value := range flatMap {
		parts := strings.Split(key, ".")
		current := nestedMap

		for i, part := range parts {
			if i == len(parts)-1 {
				current[part] = value
			} else {
				if _, exists := current[part]; !exists {
					current[part] = make(map[string]any)
				}
				current = current[part].(map[string]any)
			}
		}
	}
	return nestedMap
}
