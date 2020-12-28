package utils

import (
	"fmt"
	"testing"
)

func TestRandStringBytesMaskImprSrcUnsafe(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		want int
	}{
		{
			name: "a",
			want: 8,
		},
		{
			name: "b",
			want: 16,
		},
		{
			name: "c",
			want: 32,
		},
		{
			name: "c",
			want: 64,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := randStringBytesMaskImprSrcUnsafe(tt.want); len(got) != tt.want {
				t.Errorf("RandStringBytesMaskImprSrcUnsafe() = %v, want %v", got, tt.want)
			} else {
				fmt.Println(got)
			}
		})
	}
}
