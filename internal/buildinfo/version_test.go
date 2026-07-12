package buildinfo

import (
	"runtime/debug"
	"testing"
)

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

func TestVersionUsesTaggedModuleMetadata(t *testing.T) {
	if got := versionFromBuildInfo(developmentVersion, &debug.BuildInfo{Main: debug.Module{Version: "v1.2.3"}}, true); got != "1.2.3" {
		t.Fatalf("versionFromBuildInfo() = %q, want tagged module version", got)
	}
}

func TestVersionRejectsDevelopmentModuleMetadata(t *testing.T) {
	if got := versionFromBuildInfo(developmentVersion, &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, true); got != developmentVersion {
		t.Fatalf("versionFromBuildInfo() = %q, want development fallback", got)
	}
}

func TestVersionRejectsUnavailableBuildMetadata(t *testing.T) {
	if got := versionFromBuildInfo(developmentVersion, nil, false); got != developmentVersion {
		t.Fatalf("versionFromBuildInfo() = %q, want development fallback", got)
	}
}
