package algorithm

import "testing"

func BenchmarkNoSentinel(b *testing.B) {
	var list []int
	for i := 1; i < 10000; i++ {
		list = append(list, i)
	}
	key:=5000
	for i := 0; i < b.N; i++ {
		news := list
		noSentinel(news, key)
	}
}

func BenchmarkSentinel(b *testing.B) {
	var list []int
	for i := 1; i < 10000; i++ {
		list = append(list, i)
	}
	key:=5000
	for i := 0; i < b.N; i++ {
		news := list
		sentinel(news, key)
	}
}
