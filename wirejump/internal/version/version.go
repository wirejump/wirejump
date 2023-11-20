package version

import "fmt"

var (
	BuildCommit string
	BuildDate   string
	Version     string
)

func VersionString() string {
	return fmt.Sprintf("WireJump %s-%s built @%s", Version, BuildCommit, BuildDate)
}
