package main

func function(a, b int) (int, int) {
	return a + b, a - b
}

//go tool compile -S -N -l 02-function.go
func main() {
	function(11, 22)
}