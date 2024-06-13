package utils

// 删除字符串
func RemoveStr(slice []string, s string) []string {
	for i, item := range slice {
		if item == s {
			slice = append(slice[:i], slice[i+1:]...)
			return slice
		}
	}
	return slice
}
