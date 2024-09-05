package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
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

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", fmt.Errorf("ReadBuildInfo failed")
	}

	vcs_modified := ""
	vcs_rev := ""
	vcs_time := ""
	for _, element := range info.Settings {
		if element.Key == "vcs.revision" {
			vcs_rev = safeSubstr(element.Value, 8)
		} else if element.Key == "vcs.time" {
			vcs_time = safeSubstr(element.Value, 10)
		} else if element.Key == "vcs.modified" {
			if element.Value == "true" {
				vcs_modified = "(LOCALLY MODIFIED)"
			} else {
				vcs_modified = ""
			}
		}
	}
	return fmt.Sprintf("%s date=%s commit=%s %s", programName, vcs_time, vcs_rev, vcs_modified), nil
}
