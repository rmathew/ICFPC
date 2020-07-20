package galaxy

import (
	"strings"
	"testing"
)

func TestModulate(t *testing.T) {
	tests := []struct {
		i    int64
		want string
	}{
		// Spaces inserted here for understanding to be removed before testing.
		{0, "01 0"},
		{1, "01 10 0001"},
		{-1, "10 10 0001"},
		{16, "01 110 0001 0000"},
		{-16, "10 110 0001 0000"},
		{255, "01 110 1111 1111"},
		{-255, "10 110 1111 1111"},
		{256, "01 1110 0001 0000 0000"},
		{-256, "10 1110 0001 0000 0000"},
	}
	for _, tc := range tests {
		got := modulate(tc.i)
		if got != strings.ReplaceAll(tc.want, " ", "") {
			t.Errorf("For %d, wanted %q, got %q.", tc.i, tc.want, got)
		}
	}
}

func TestDemodulate(t *testing.T) {
	tests := []struct {
		s    string
		want int64
	}{
		// Spaces inserted here for understanding to be removed before testing.
		{"01 0", 0},
		{"01 10 0001", 1},
		{"10 10 0001", -1},
		{"01 110 0001 0000", 16},
		{"10 110 0001 0000", -16},
		{"01 110 1111 1111", 255},
		{"10 110 1111 1111", -255},
		{"01 1110 0001 0000 0000", 256},
		{"10 1110 0001 0000 0000", -256},
	}
	for _, tc := range tests {
		got, err := demodulate([]rune(strings.ReplaceAll(tc.s, " ", "")))
		if err != nil {
			t.Errorf("For %q, got error %v.", tc.s, err)
		}
		if got != tc.want {
			t.Errorf("For %q, wanted %d, got %d.", tc.s, tc.want, got)
		}
	}
}

func TestModulateList(t *testing.T) {
	tests := []struct {
		e    string
		want string
	}{
		// Spaces inserted here for understanding to be removed before testing.
		{"nil", "11"},
		{"ap ap cons nil nil", "11 00 00"},
		{"ap ap cons 0 nil", "11 010 00"},
		{"ap ap cons 1 2", "11 01100001 01100010"},
		{"ap ap cons 1 ap ap cons 2 nil", "11 01100001 11 01100010 00"},
	}
	for _, tc := range tests {
		e, err := strToExpr(tc.e)
		if err != nil {
			t.Errorf("Error converting %q to expr %v.", tc.e, err)
		}
		got, grr := modulateList(e)
		if grr != nil {
			t.Errorf("For %q, got error %v.", tc.e, grr)
		}
		if got != strings.ReplaceAll(tc.want, " ", "") {
			t.Errorf("For %q, wanted %q, got %q.", tc.e, tc.want, got)
		}
	}
}
