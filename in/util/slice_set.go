package util

func SetSliceList(idList []string) []string {
	newIdList := make([]string, 0)
	setMap := make(map[string]struct{}, 0)
	for _, id := range idList {
		_, ok := setMap[id]
		if !ok {
			newIdList = append(newIdList, id)
			setMap[id] = struct{}{}
		}
	}
	return newIdList
}
