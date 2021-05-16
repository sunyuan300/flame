package fshare

func SliceDeduplication(s []string) []string {
	var res []string
	tempMap := map[string]struct{}{}
	for _, e := range s {
		l := len(tempMap)
		tempMap[e] = struct{}{}
		if len(tempMap) != l {
			res = append(res, e)
		}
	}
	return res
}
