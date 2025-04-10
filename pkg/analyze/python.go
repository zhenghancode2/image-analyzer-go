package analyze

import (
	"os"
	"path/filepath"
	"strings"
)

func ListPythonPackages(root string) []string {
	var pkgs []string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err == nil && strings.HasSuffix(info.Name(), ".dist-info") {
			pkgs = append(pkgs, filepath.Base(path))
		}
		return nil
	})
	return pkgs
}
