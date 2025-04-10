package analyze

import (
	"os"
	"path/filepath"
)

func CheckCommonTools(root string) map[string]bool {
	tools := []string{"sshd", "python3", "curl", "wget", "nvcc"}
	result := make(map[string]bool)
	for _, t := range tools {
		found := false
		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err == nil && info.Name() == t {
				found = true
			}
			return nil
		})
		result[t] = found
	}
	return result
}
