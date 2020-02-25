// +build freebsd linux netbsd openbsd

package lxr

import (
	"path/filepath"
)

func GetSystemTablePath() (string, error) {
	path := filepath.Join("/var", "lib", "LXRHash")
	return path, nil
}
