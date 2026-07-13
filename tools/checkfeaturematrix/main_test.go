package main

import "testing"

func TestIssueReferencePattern(t *testing.T) {
	t.Parallel()
	cases := []struct {
		value string
		want  bool
	}{
		{value: "#1", want: true},
		{value: "#80", want: true},
		{value: "80", want: false},
		{value: "#0", want: false},
		{value: "#abc", want: false},
	}
	for _, tc := range cases {
		if got := issueReferencePattern.MatchString(tc.value); got != tc.want {
			t.Errorf("issueReferencePattern.MatchString(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}
}
