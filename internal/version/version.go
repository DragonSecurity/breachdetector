package version

import (
	"fmt"
	"github.com/carlmjohnson/versioninfo"
)

func Get() string {
	fmt.Println("Version:", versioninfo.Version)
	fmt.Println("Revision:", versioninfo.Revision)
	fmt.Println("DirtyBuild:", versioninfo.DirtyBuild)
	fmt.Println("LastCommit:", versioninfo.LastCommit)
	return fmt.Sprintf("%s", versioninfo.Short())
}
