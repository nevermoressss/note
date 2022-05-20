package algorithm

// 二分
func binarySearch(list []int, key int) int {
	low, high := 0, len(list)-1
	for low <= high {
		mid := (low + high) / 2
		if key < list[mid] {
			high = mid - 1
			continue
		}
		if key > list[mid] {
			low = mid + 1
			continue
		}
		return mid
	}
	return -1
}
