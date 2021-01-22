package algorithm

import (
	"testing"
)

func Test_insertionSort(t *testing.T) {
	type args struct {
		a []int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test",
			args: args{
				a: []int{1, 2, 3, 7, 8, 4, 6, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InsertionSort(tt.args.a)
			t.Log(tt.args.a)
		})
	}
}

func Test_bubbleSort(t *testing.T) {
	type args struct {
		a []int
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "test",
			args: args{
				a: []int{1, 2, 3, 7, 8, 4, 6, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			BubbleSort(tt.args.a)
			t.Log(tt.args.a)
		})
	}
}

func Test_selectionSort(t *testing.T) {
	type args struct {
		a []int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			args: args{
				a: []int{1, 2, 3, 7, 8, 4, 6, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SelectionSort(tt.args.a)
			t.Log(tt.args.a)
		})
	}
}

func TestMergeSort(t *testing.T) {
	type args struct {
		arr []int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test",
			args: args{
				arr: []int{1, 2, 3, 7, 8, 4, 6, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MergeSort(tt.args.arr)
			t.Log(tt.args.arr)
		})
	}
}

func TestQuickSort(t *testing.T) {
	type args struct {
		arr []int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test",
			args: args{
				arr: []int{1, 2, 3, 7, 8, 4, 6, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			QuickSort(tt.args.arr)
			t.Log(tt.args.arr)
		})
	}
}

