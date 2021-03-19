package pathresolver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolvePaths find matching files within a directory. The filenames ca be filtered by pattern and extension
func ResolvePaths(sourceDir string, pattern string, extension string, ignoreDir string) ([]string, error) {
	files := make([]string, 0)

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if info.IsDir() && info.Name() == ignoreDir {
			fmt.Printf("skipping a dir without errors: %+v \n", info.Name())
			return filepath.SkipDir
		}

		if extension != "" && !strings.HasSuffix(path, extension) {
			return nil
		}

		isMatch := false
		if pattern != "" {
			if matched, _ := filepath.Match(pattern, info.Name()); matched {
				isMatch = true
			}
		}

		if isMatch {
			files = append(files, path)
		}

		return nil
	})
	return files, err
}
