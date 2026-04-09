package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
)

// If built by goreleaser, these variables are set by goreleaser using -ldflags.
var (
	version = "dev"
	commit  = "unknown"
	date    = ""
	builtBy = ""
)

func safeSubstr(str string, length int) string {
	if len(str) >= length {
		return str[:length]
	} else {
		return str
	}
}

func readBuildInfo() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	programName := filepath.Base(execPath)

	if builtBy != "goreleaser" {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			return "", fmt.Errorf("ReadBuildInfo failed")
		}

		vcsModified := false
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				commit = safeSubstr(setting.Value, 8)
			case "vcs.time":
				date = safeSubstr(setting.Value, 10)
			case "vcs.modified":
				vcsModified = setting.Value == "true"
			}
		}
		if vcsModified {
			commit = "locally_modified"
		}
	}

	return fmt.Sprintf("%s version=%s date=%s commit=%s", programName, version, date, commit), nil
}
