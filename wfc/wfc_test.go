package wfc

import (
	"testing"
)

func TestIntInSlice(t *testing.T) {
	cases := []struct {
		inSlice []int
		inInt   int
		want    bool
	}{
		{[]int{0, 1, 2, 3}, 0, true},
		{[]int{0, 1, 2, 3}, 1, true},
		{[]int{0, 1, 2, 3}, 2, true},
		{[]int{0, 1, 2, 3}, 3, true},
		{[]int{0, 1, 2, 3}, 4, false},
		{[]int{}, 0, false},
		{[]int{}, 1, false},
		{[]int{}, 2, false},
		{[]int{}, 3, false},
		{[]int{}, 4, false},
	}
	for _, c := range cases {
		// init tiles
		got := IntInSlice(c.inInt, c.inSlice)
		if got != c.want {
			t.Errorf("IntInSlice(%d, %v) == %v, want %v", c.inInt, c.inSlice, got, c.want)
		}
	}
}

func TestStringInSlice(t *testing.T) {
	cases := []struct {
		inSlice  []string
		inString string
		want     bool
	}{
		{[]string{"aaa", "bbbb", "cccc"}, "aaa", true},
		{[]string{"aaa", "bbbb", "cccc"}, "bbbb", true},
		{[]string{"aaa", "bbbb", "cccc"}, "bbbbb", false},
		{[]string{}, "bbbb", false},
	}
	for _, c := range cases {
		// init tiles
		got := StringInSlice(c.inString, c.inSlice)
		if got != c.want {
			t.Errorf("IntInSlice(%s, %v) == %v, want %v", c.inString, c.inSlice, got, c.want)
		}
	}
}
