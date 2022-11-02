package version

import (
	"fmt"
	"runtime/debug"
)

var (
	version   string
	commit    string
	date      string
	goVersion string
	modified  bool
	checksum  string
)

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		panic("read build info")
	}

	fmt.Printf("info:\n%+v\n\n", info)
	fmt.Printf("main:\n%+v\n\n", info.Main)
	fmt.Printf("setting:\n%+v\n\n", info.Settings)

	version = info.Main.Version
	goVersion = info.GoVersion
	checksum = info.Main.Sum

	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			commit = setting.Value
		case "vcs.time":
			date = setting.Value
		case "vcs.modified":
			modified = setting.Value == "true"
			version = version + "-dirty"
		default:
			continue
		}
	}
}

func Version() string {
	return version
}

func Commit() string {
	return commit
}

func Date() string {
	return date
}

func GoVersion() string {
	return goVersion
}

func Modified() bool {
	return modified
}

func MarkModified(v *string) {
	if Modified() {
		*v = *v + "-dirty"
	}
}

func SetFromCmd(v string) {
	version = v
}

func String() string {
	return fmt.Sprintf(`%s (%s) (%s)
sum: %s
`, version, commit, goVersion, checksum)
}
