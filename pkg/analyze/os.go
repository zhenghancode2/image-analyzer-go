package analyze

import (
	"os"
	"path/filepath"
)

func CheckOSInfo(root string) string {
	path := filepath.Join(root, "etc", "os-release")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}
