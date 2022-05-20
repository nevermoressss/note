package common

// alignUp rounds n up to a multiple of a. a must be a power of 2.
// 将n向上舍入为a的倍数,a 为2次幂
func AlignUp(n, a int) int {
	return (n + a - 1) &^ (a - 1)
}

// alignDown rounds n down to a multiple of a. a must be a power of 2.
// 将n向下舍入为a的倍数,a 为2次幂
func AlignDown(n, a int) int {
	return n &^ (a - 1)
}
