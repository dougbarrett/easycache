package easycache

import (
	"fmt"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	type args struct {
		ttl time.Duration
		fn  func(key any) any
	}
	tests := []struct {
		name     string
		args     args
		key      any
		expected string
	}{
		{
			name: "happy path",
			args: args{
				ttl: 1 * time.Second,
				fn: func(key any) any {
					return fmt.Sprintf("happy path %s", key)
				},
			},
			key:      "1",
			expected: "happy path 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.args.ttl, tt.args.fn)

			want, err := c(tt.key)

			if err != nil {
				t.Errorf("New() error = %v, wantErr %v", err, nil)
				return
			}

			if tt.expected != want {
				t.Errorf("New() = %v, want %v", want, tt.expected)
			}

		})
	}
}
