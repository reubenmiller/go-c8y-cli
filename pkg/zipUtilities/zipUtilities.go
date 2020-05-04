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

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
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

func UnzipFile(src string, dest string, names []string) ([]string, error) {
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Only unzip names files (if given), otherwise unzip everything
		if len(names) > 0 {
			found := true
			for _, iName := range names {
				if strings.ToLower(iName) == strings.ToLower(f.Name) {
					found = false
					break
				}
			}
			if !found {
				// Skip file
				continue
			}
		}

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}
