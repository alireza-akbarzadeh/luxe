package version

import "os"

// Link-time overrides (see makefile build LDFLAGS).
var (
	Version = "0.0.1"
	BuildID = "dev"
)

// Info is exposed on health endpoints for deploy identity.
type Info struct {
	Version string `json:"version"`
	BuildID string `json:"buildId"`
}

// Get returns release version and build id, with APP_VERSION / GIT_COMMIT env overrides for CI.
func Get() Info {
	v := Version
	if env := os.Getenv("APP_VERSION"); env != "" {
		v = env
	}

	b := BuildID
	if env := os.Getenv("GIT_COMMIT"); env != "" {
		b = env
	}

	return Info{Version: v, BuildID: b}
}
