package wfc

import (
	"testing"
)

func TestFilterOptions(t *testing.T) {
	tests := []struct {
		name     string
		orig     []string
		options  []string
		expected []string
	}{
		{"empty orig", []string{}, []string{"a", "b"}, []string{}},
		{"empty options", []string{"a", "b"}, []string{}, []string{}},
		{"no matches", []string{"a", "b"}, []string{"c", "d"}, []string{}},
		{"some matches", []string{"a", "b", "c"}, []string{"a", "c"}, []string{"a", "c"}},
		{"all matches", []string{"a", "b"}, []string{"a", "b"}, []string{"a", "b"}},
		{"options with duplicates", []string{"a", "b"}, []string{"a", "a", "b"}, []string{"a", "b"}},
		{"orig with duplicates", []string{"a", "a", "b"}, []string{"a", "b"}, []string{"a", "a", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wfc := &Wfc{}
			actual := wfc.FilterOptions(tt.orig, tt.options)
			if !sliceEqual(actual, tt.expected) {
				t.Errorf("FilterOptions(%v, %v) = %v, want %v", tt.orig, tt.options, actual, tt.expected)
			}
		})
	}
}

func BenchmarkFilterOptions(b *testing.B) {
	wfc := &Wfc{}
	orig := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
	options := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
	for i := 0; i < b.N; i++ {
		wfc.FilterOptions(orig, options)
	}
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
