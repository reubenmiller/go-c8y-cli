package zipUtilities

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ZipDirectoryContents zips the contents of a source folder. The root folder will be excluded from the created zip file
func ZipDirectoryContents(source, target string) error {
	return zipit(source, target, true)
}

// ZipDirectory zips the source folder. The root folder will be included in the created zip file
func ZipDirectory(source, target string) error {
	return zipit(source, target, false)
}

// ZipFile zips the source file
func ZipFile(source, target string) error {
	return zipit(source, target, false)
}

// Zip zips the source and writes it to the given target.
// If the source is a folder, then the folder will be zipped, but the excludeRoot option can be used
// to specify if the root folder should be included within the zipped file.
func zipit(source, target string, excludeRoot bool) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)

		// has trailing path
		if strings.HasSuffix(baseDir, "\\") || strings.HasSuffix(baseDir, "/") {
			baseDir = fmt.Sprintf("%s.%c", baseDir, os.PathSeparator)
		}
	}

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if baseDir != "" && excludeRoot && source == path {
			// Skip the root folder
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			if excludeRoot {
				header.Name = strings.TrimPrefix(path, source)
			} else {
				header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
			}
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}
