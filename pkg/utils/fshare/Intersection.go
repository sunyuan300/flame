package fshare

func Intersect(sliceStringA, sliceStringB []string) []string {
	if len(sliceStringA) == 0 {
		return sliceStringB
	}
	if len(sliceStringB) == 0 {
		return sliceStringA
	}
	m := make(map[string]int)
	n := make([]string, 0)
	for _, v := range sliceStringA {
		m[v]++
	}

	for _, v := range sliceStringB {
		times, _ := m[v]
		if times == 1 {
			n = append(n, v)
		}
	}
	return n
}
