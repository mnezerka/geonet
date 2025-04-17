package utils

func MapsEqual[K comparable, V comparable](a, b map[K]V) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || bv != v {
			return false
		}
	}
	return true
}

func MapKeys[K comparable, V any](m map[K]V) []K {
	result := []K{}
	for k := range m {
		result = append(result, k)
	}
	return result
}

func MapValues[K comparable, V any](m map[K]V) []V {
	result := []V{}
	for _, v := range m {
		result = append(result, v)
	}
	return result
}

func MapCopy[K comparable, V any](original map[K]V) map[K]V {
	copy := make(map[K]V, len(original))
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

func MapsMerge[K comparable, V any](dst, src map[K]V) {
	for k, v := range src {
		dst[k] = v // overwrites existing keys
	}
}
