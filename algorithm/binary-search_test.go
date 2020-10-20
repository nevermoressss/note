package algorithm

import "testing"

func Test_binarySearch(t *testing.T) {
	type args struct {
		list []int
		key  int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name:"left",
			args:args{
				list:[]int{1,2,3,4,5,6,7,8,9,10},
				key:1,
			},
			want:0,
		},
		{
			name:"right",
			args:args{
				list:[]int{1,2,3,4,5,6,7,8,9,10},
				key:10,
			},
			want:9,
		},
		{
			name:"no1",
			args:args{
				list:[]int{1,2,3,4,5,6,7,8,9,10},
				key:-10,
			},
			want:-1,
		},
		{
			name:"no2",
			args:args{
				list:[]int{1,2,3,4,5,6,7,8,9,10},
				key:20,
			},
			want:-1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := binarySearch(tt.args.list, tt.args.key); got != tt.want {
				t.Errorf("binarySearch() = %v, want %v", got, tt.want)
			}
		})
	}
}
