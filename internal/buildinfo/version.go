// Package buildinfo resolves the version embedded in a Go binary.
package buildinfo

import (
	"runtime/debug"
	"strings"
)

const developmentVersion = "0.0.0-dev"

var readBuildInfo = debug.ReadBuildInfo

// Version returns an explicit build version when supplied, otherwise it uses
// the module version embedded by `go install module/path/cmd@version`.
func Version(explicit string) string {
	if explicit != "" && explicit != developmentVersion {
		return explicit
	}

	info, ok := readBuildInfo()
	return versionFromBuildInfo(explicit, info, ok)
}

func versionFromBuildInfo(explicit string, info *debug.BuildInfo, ok bool) string {
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return explicit
	}

	return strings.TrimPrefix(info.Main.Version, "v")
}
