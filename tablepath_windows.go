// +build windows

package lxr

import (
	"path/filepath"

	"golang.org/x/sys/windows"
)

func GetSystemTablePath() (string, error) {
	pdpath, err := windows.KnownFolderPath(windows.FOLDERID_ProgramData, 0)
	if err != nil {
		return "", err
	}
	path := filepath.Join(pdpath, "LXRHash")
	return path, nil
}
