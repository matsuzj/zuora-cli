// Package build contains build-time metadata injected via ldflags.
package build

import "runtime/debug"

// Variables set at build time via -ldflags.
var (
	Version = "dev"
	Date    = ""
	Commit  = ""
)

func init() {
	if Version != "dev" {
		return
	}
	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		Version = info.Main.Version
	}
}
