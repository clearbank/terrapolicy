package utils

func UniqString(a []string) []string {
	length := len(a)
	seen := make(map[string]struct{}, length)
	j := 0

	for i := 0; i < length; i++ {
		v := a[i]
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		a[j] = v
		j++
	}

	return a[0:j]
}
