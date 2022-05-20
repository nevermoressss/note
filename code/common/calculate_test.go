package common

import (
	"testing"
)

func TestAlignUp(t *testing.T) {
	type args struct {
		n int
		a int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test",
			args: args{
				824633917248,
				1024,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AlignUp(tt.args.n, tt.args.a); got != tt.want {
				t.Errorf("AlignUp() = %v, want %v", got, tt.want)
			}
		})
	}
}
