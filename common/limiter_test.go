package common

import (
	"testing"
)

func TestNewChannelLimiter(t *testing.T) {
	type args struct {
		limiter int
	}
	tests := []struct {
		name string
		args args
		want limiter
	}{
		{
			name: "test",
			args: args{limiter: 10},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit := NewChannelLimiter(tt.args.limiter)
			for i := 0; i < tt.args.limiter; i++ {
				t.Log("limit notWait", limit.NotWait())
			}
			t.Log("limit notWait", limit.NotWait())
			for i := 0; i < tt.args.limiter; i++ {
				limit.Release()
				t.Log("limit Relaase")
			}
			limit.Close()
		})
	}
}
