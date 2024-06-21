package utils

import "strings"

// 删除字符串
// 如果连堂课,或者普通课时间已经被使用,则从可用时间中删掉
func RemoveRelatedItems(slice []string, itemToRemove string) []string {

	parts := strings.Split(itemToRemove, "_")
	part1, part2 := parts[0], ""
	if len(parts) == 2 {
		part2 = parts[1]
	}

	var result []string
	for _, item := range slice {
		itemParts := strings.Split(item, "_")
		if len(itemParts) == 2 {
			if part2 != "" {
				if itemParts[0] != part1 && itemParts[1] != part1 && itemParts[0] != part2 && itemParts[1] != part2 {
					result = append(result, item)
				}
			} else {
				if itemParts[0] != part1 && itemParts[1] != part1 {
					result = append(result, item)
				}
			}
		} else {
			if part2 != "" {
				if item != part1 && item != part2 {
					result = append(result, item)
				}
			} else {
				if item != part1 {
					result = append(result, item)
				}
			}
		}
	}

	return result
}
