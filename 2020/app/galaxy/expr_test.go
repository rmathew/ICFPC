package galaxy

import (
	"fmt"
	"testing"
)

func TestExtrList(t *testing.T) {
	tests := []struct {
		e    string
		want []string
	}{
		{"nil", []string{}},
		{"ap ap cons 1 nil", []string{"1"}},
		{"ap ap cons 1 ap ap cons 2 nil", []string{"1", "2"}},
		{"ap ap cons 1 ap ap cons 2 ap ap cons 3 nil", []string{"1", "2", "3"}},
		{"ap ap cons 1 ap ap cons ap ap cons 2 ap ap cons 3 nil nil",
			[]string{"1", "[2, 3]"}},
	}
	for _, tc := range tests {
		e, err := strToExpr(tc.e)
		if err != nil {
			t.Errorf("Error converting %q to expr %v.", tc.e, err)
		}
		got, gErr := extrList(e)
		if gErr != nil {
			t.Errorf("For %q, got error %v.", tc.e, gErr)
		}
		if len(got) != len(tc.want) {
			t.Errorf("For %q, expected %d elements, got %d.", tc.e,
				len(tc.want), len(got))
		}
		for i, g := range got {
			gStr := fmt.Sprintf("%v", g)
			if gStr != tc.want[i] {
				t.Errorf("For %q, list-element at %d expected as %q, got %q.",
					tc.e, i, tc.want[i], gStr)

			}
		}
	}
}
