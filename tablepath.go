package lxr

import (
	"os/user"
	"path/filepath"
)

func GetUserTablePath() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	userTablePath := filepath.Join(u.HomeDir, ".lxrhash")
	return userTablePath, nil
}
