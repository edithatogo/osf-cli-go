package zenodoid

import "testing"

func TestParseRecord(t *testing.T) {
	for input, want := range map[string]string{"123": "123", "zenodo:record:123": "123", "https://zenodo.org/records/123/": "123", "https://sandbox.zenodo.org/records/456": "456"} {
		got, err := ParseRecord(input)
		if err != nil || got != want {
			t.Fatalf("ParseRecord(%q)=%q,%v want %q", input, got, err, want)
		}
	}
	for _, input := range []string{"", "abc", "osf:project:123", "zenodo:record:abc", "https://user@zenodo.org/records/1", "https://zenodo.org:444/records/1", "https://zenodo.org/records/1?q=x", "https://example.org/records/1", "https://zenodo.org/deposits/1", "1/2"} {
		if _, err := ParseRecord(input); err == nil {
			t.Fatalf("ParseRecord(%q) succeeded", input)
		}
	}
}
