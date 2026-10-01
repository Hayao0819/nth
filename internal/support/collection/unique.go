package collection

func AppendUniqueBy[T any, K comparable](current, incoming []T, key func(T) K) []T {
	seen := make(map[K]struct{}, len(current)+len(incoming))
	result := make([]T, 0, len(current)+len(incoming))
	for _, items := range [][]T{current, incoming} {
		for _, item := range items {
			value := key(item)
			if _, exists := seen[value]; exists {
				continue
			}
			seen[value] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}
