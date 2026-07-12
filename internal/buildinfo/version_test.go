package buildinfo

import "testing"

func TestVersionUsesExplicitReleaseVersion(t *testing.T) {
	if got := Version("1.2.3"); got != "1.2.3" {
		t.Fatalf("Version() = %q, want explicit release version", got)
	}
}

func TestVersionPreservesDevelopmentFallback(t *testing.T) {
	if got := Version(developmentVersion); got != developmentVersion {
		t.Fatalf("Version() = %q, want development fallback", got)
	}
}
