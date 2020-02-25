// +build darwin

package lxr

import (
	"path/filepath"
)

const BundleIdentifier = "org.pegnet.LXRHash"

func GetSystemTablePath() (string, error) {
	path := filepath.Join("/Library", "Application Support", BundleIdentifier)
	return path, nil
}
