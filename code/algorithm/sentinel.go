package algorithm

// 查找
func noSentinel(list []int, key int) int {
	if len(list) == 0 {
		return -1
	}
	for i := 0; i < len(list); i++ {
		if list[i] == key {
			return i
		}
	}
	return -1
}

// 哨兵
func sentinel(list []int, key int) int {
	if len(list) == 0 {
		return -1
	}
	// 吧0号位当做哨兵
	list[0] = key
	i := len(list) - 1
	for list[i] != key {
		i--
	}
	if i == 0 {
		return -1
	}
	return i
}
