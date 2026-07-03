package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDummy(t *testing.T) {
	assert.True(t, true)
}

func TestFMax(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want []*int
	}{
		{
			name: "empty slice",
			in:   []int{},
			want: []*int{nil, nil},
		},
		{
			name: "one element",
			in:   []int{5},
			want: []*int{new(5), new(5)},
		},
		{
			name: "two elements ascending",
			in:   []int{1, 2},
			want: []*int{new(2), new(1)},
		},
		{
			name: "two elements descending",
			in:   []int{2, 1},
			want: []*int{new(2), new(1)},
		},
		{
			name: "max in middle",
			in:   []int{1, 5, 3},
			want: []*int{new(5), new(3)},
		},
		{
			name: "all negative",
			in:   []int{-10, -3, -7, -1},
			want: []*int{new(-1), new(-3)},
		},
		{
			name: "with zero",
			in:   []int{-1, 0, -5},
			want: []*int{new(0), new(-1)},
		},
		{
			name: "duplicates max",
			in:   []int{5, 1, 5, 3},
			want: []*int{new(5), new(3)},
		},
		{
			name: "all equal",
			in:   []int{7, 7, 7},
			want: []*int{new(7), new(7)},
		},
		{
			name: "second max duplicated",
			in:   []int{10, 8, 8, 1},
			want: []*int{new(10), new(8)},
		},
		{
			name: "min int values",
			in:   []int{-2147483648, -2147483647},
			want: []*int{new(-2147483647), new(-2147483648)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fMax(tt.in)

			if len(got) != len(tt.want) {
				t.Fatalf("len got = %d, want %d", len(got), len(tt.want))
			}

			for i := range tt.want {
				if tt.want[i] == nil {
					if got[i] != nil {
						t.Fatalf("got[%d] = %v, want nil", i, *got[i])
					}
					continue
				}

				if got[i] == nil {
					t.Fatalf("got[%d] = nil, want %d", i, *tt.want[i])
				}

				if *got[i] != *tt.want[i] {
					t.Fatalf("got[%d] = %d, want %d", i, *got[i], *tt.want[i])
				}
			}
		})
	}
}
