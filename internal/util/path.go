package util

import (
	"path/filepath"
	"strings"
)

var commonDirs = map[string]bool{
	"src":      true,
	"internal": true,
	"cmd":      true,
	"pkg":      true,
	"api":      true,
	"docs":     true,
	"test":     true,
	"tests":    true,
	"examples": true,
	"vendor":   true,
	"tmp":      true,
	"temp":     true,
}

func ResolvePath(cwd string) string {
	cwd = filepath.Clean(cwd)

	if cwd == "" || cwd == "/" || cwd == "." {
		return "default"
	}

	parts := strings.Split(cwd, string(filepath.Separator))

	for i := len(parts) - 1; i >= 0; i-- {
		dir := parts[i]

		if dir == "" || dir == "." || dir == ".." {
			continue
		}

		if !commonDirs[dir] {
			return dir
		}
	}

	return "default"
}

func AddCommonDir(dir string) {
	commonDirs[dir] = true
}

func RemoveCommonDir(dir string) {
	delete(commonDirs, dir)
}

func IsCommonDir(dir string) bool {
	return commonDirs[dir]
}
