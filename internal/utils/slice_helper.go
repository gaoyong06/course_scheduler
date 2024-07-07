package utils

func RemoveElement(slice []int, element int) []int {
	for i, v := range slice {
		if v == element {
			slice = append(slice[:i], slice[i+1:]...)
			break
		}
	}
	return slice
}
