package extension

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// assetType is the type of a release asset which can be installed
type assetType int

const (
	assetNone assetType = iota
	assetTarGz
	assetZip
	assetBinary
)

func (a assetType) IsArchive() bool {
	return a == assetTarGz || a == assetZip
}

// Known operating systems and architectures, used to detect if a release contains binary assets
var knownOS = []string{
	"aix", "android", "darwin", "dragonfly", "freebsd", "illumos", "ios", "js", "linux",
	"netbsd", "openbsd", "plan9", "solaris", "windows",
}

var knownArch = []string{
	"386", "amd64", "arm", "arm64", "armv5", "armv6", "armv7", "loong64", "mips", "mips64",
	"mips64le", "mipsle", "ppc64", "ppc64le", "riscv64", "s390x", "wasm",
	// Aliases used by some release tools
	"x86_64", "i386", "aarch64",
}

// archAliases are other names which are commonly used in release asset names
var archAliases = map[string][]string{
	"amd64": {"x86_64"},
	"386":   {"i386"},
	"arm64": {"aarch64"},
}

var archiveSuffixes = `(\.tar\.gz|\.tgz|\.zip)`

var anyPlatformArchive = regexp.MustCompile(
	`(?i)[_-](` + strings.Join(knownOS, "|") + `)[_-](` + strings.Join(knownArch, "|") + `)` + archiveSuffixes + `$`,
)

var anyPlatformBinary = regexp.MustCompile(
	`(?i)-(` + strings.Join(knownOS, "|") + `)-(` + strings.Join(knownArch, "|") + `)(\.exe)?$`,
)

func archiveType(name string) assetType {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return assetTarGz
	case strings.HasSuffix(lower, ".zip"):
		return assetZip
	}
	return assetNone
}

// hasPlatformAssets returns true if the release contains at least one asset for any platform
func hasPlatformAssets(r *release) bool {
	for _, a := range r.Assets {
		if anyPlatformArchive.MatchString(a.Name) || anyPlatformBinary.MatchString(a.Name) {
			return true
		}
	}
	return false
}

func archNames(goarch string) []string {
	return append([]string{goarch}, archAliases[goarch]...)
}

// matchPlatform returns the matching asset for the given os and arch
func matchPlatform(assets []releaseAsset, goos, goarch string) (*releaseAsset, assetType) {
	arches := make([]string, 0)
	for _, arch := range archNames(goarch) {
		arches = append(arches, regexp.QuoteMeta(arch))
	}
	archPattern := strings.Join(arches, "|")
	osPattern := regexp.QuoteMeta(goos)

	archiveRe := regexp.MustCompile(`(?i)[_-]` + osPattern + `[_-](` + archPattern + `)` + archiveSuffixes + `$`)
	for i, a := range assets {
		if archiveRe.MatchString(a.Name) {
			return &assets[i], archiveType(a.Name)
		}
	}

	binSuffix := ""
	if goos == "windows" {
		binSuffix = `\.exe`
	}
	binaryRe := regexp.MustCompile(`(?i)-` + osPattern + `-(` + archPattern + `)` + binSuffix + `$`)
	for i, a := range assets {
		if binaryRe.MatchString(a.Name) {
			return &assets[i], assetBinary
		}
	}
	return nil, assetNone
}

// selectAsset selects the release asset for the platform. Archives are preferred over bare binaries.
// On darwin/arm64, amd64 assets are used if rosetta is available. The returned bool is
// true if the asset is a native darwin/arm64 binary (which needs to be codesigned)
func selectAsset(assets []releaseAsset, goos, goarch string, rosetta bool) (*releaseAsset, assetType, bool) {
	if asset, kind := matchPlatform(assets, goos, goarch); asset != nil {
		return asset, kind, goos == "darwin" && goarch == "arm64"
	}
	if goos == "darwin" && goarch == "arm64" && rosetta {
		if asset, kind := matchPlatform(assets, goos, "amd64"); asset != nil {
			return asset, kind, false
		}
	}
	return nil, assetNone, false
}

// findChecksumAsset returns the checksums file of a release (GoReleaser format), if present
func findChecksumAsset(assets []releaseAsset) *releaseAsset {
	for i, a := range assets {
		lower := strings.ToLower(a.Name)
		if lower == "checksums.txt" || strings.HasSuffix(lower, "_checksums.txt") || strings.HasSuffix(lower, "-checksums.txt") {
			return &assets[i]
		}
	}
	return nil
}

// parseChecksums parses a checksums file in the format "<sha256>  <file name>"
func parseChecksums(r io.Reader) (map[string]string, error) {
	checksums := make(map[string]string)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 {
			continue
		}
		// sha256sum binary mode prefixes the file name with "*"
		checksums[strings.TrimPrefix(fields[1], "*")] = strings.ToLower(fields[0])
	}
	return checksums, scanner.Err()
}

// fileSHA256 returns the hex encoded sha256 checksum of a file
func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ErrChecksumMismatch is returned when the downloaded asset does not match the published checksum
var ErrChecksumMismatch = errors.New("checksum mismatch")

// maxExtractedSize is the maximum total size of all files extracted from an archive
var maxExtractedSize int64 = 1 << 30 // 1 GiB

// ErrArchiveTooLarge is returned if the extracted files exceed maxExtractedSize
var ErrArchiveTooLarge = errors.New("archive exceeds the maximum extracted size")

// safeJoin joins an archive entry name to the destination folder. Absolute paths and paths
// which would be outside of the destination folder are rejected
func safeJoin(dest, name string) (string, error) {
	name = strings.ReplaceAll(name, `\`, "/")
	if strings.HasPrefix(name, "/") || filepath.IsAbs(name) || filepath.VolumeName(name) != "" {
		return "", fmt.Errorf("archive entry has an absolute path: %s", name)
	}
	for _, part := range strings.Split(name, "/") {
		if part == ".." {
			return "", fmt.Errorf("archive entry is outside of the target folder: %s", name)
		}
	}
	return filepath.Join(dest, filepath.FromSlash(name)), nil
}

// checkLinkTarget checks that a symlink at path (inside dest) pointing to target stays inside dest
func checkLinkTarget(dest, path, target string) error {
	if filepath.IsAbs(target) || strings.HasPrefix(target, "/") {
		return fmt.Errorf("archive entry is a symlink with an absolute target: %s -> %s", path, target)
	}
	resolved := filepath.Join(filepath.Dir(path), filepath.FromSlash(target))
	rel, err := filepath.Rel(dest, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("archive entry is a symlink outside of the target folder: %s -> %s", path, target)
	}
	return nil
}

type sizeLimiter struct {
	remaining int64
}

func (s *sizeLimiter) copy(dst io.Writer, src io.Reader) error {
	n, err := io.Copy(dst, io.LimitReader(src, s.remaining+1))
	s.remaining -= n
	if err != nil {
		return err
	}
	if s.remaining < 0 {
		return ErrArchiveTooLarge
	}
	return nil
}

func writeExtractedFile(path string, mode os.FileMode, r io.Reader, limiter *sizeLimiter) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	// Only keep the permission bits, and make sure the owner can read and write the file
	perm := (mode.Perm() & 0755) | 0600
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if err := limiter.copy(f, r); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// extractArchive extracts a .tar.gz or .zip archive to dest
func extractArchive(src string, kind assetType, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	switch kind {
	case assetTarGz:
		return extractTarGz(src, dest)
	case assetZip:
		return extractZip(src, dest)
	}
	return fmt.Errorf("unsupported archive: %s", src)
}

func extractTarGz(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	limiter := &sizeLimiter{remaining: maxExtractedSize}
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		target, err := safeJoin(dest, hdr.Name)
		if err != nil {
			return err
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := writeExtractedFile(target, hdr.FileInfo().Mode(), tr, limiter); err != nil {
				return err
			}
		case tar.TypeSymlink:
			if err := checkLinkTarget(dest, target, hdr.Linkname); err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			if err := os.Symlink(hdr.Linkname, target); err != nil {
				return err
			}
		case tar.TypeXGlobalHeader, tar.TypeXHeader:
			// metadata only
		default:
			return fmt.Errorf("archive entry has an unsupported type: %s", hdr.Name)
		}
	}
}

func extractZip(src, dest string) error {
	zr, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer zr.Close()

	limiter := &sizeLimiter{remaining: maxExtractedSize}
	for _, file := range zr.File {
		target, err := safeJoin(dest, file.Name)
		if err != nil {
			return err
		}
		mode := file.Mode()
		switch {
		case mode.IsDir():
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case mode&os.ModeSymlink != 0:
			rc, err := file.Open()
			if err != nil {
				return err
			}
			link, err := io.ReadAll(io.LimitReader(rc, 4096))
			rc.Close()
			if err != nil {
				return err
			}
			if err := checkLinkTarget(dest, target, string(link)); err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			if err := os.Symlink(string(link), target); err != nil {
				return err
			}
		case mode.IsRegular():
			rc, err := file.Open()
			if err != nil {
				return err
			}
			// zip files created on Windows have no unix permissions, so default to executable
			// as the binary would otherwise not be runnable
			if mode.Perm() == 0 {
				mode = 0755
			}
			err = writeExtractedFile(target, mode, rc, limiter)
			rc.Close()
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("archive entry has an unsupported type: %s", file.Name)
		}
	}
	return nil
}

// extensionRoot returns the folder containing extension.yaml. If the folder only contains
// a single folder, then that folder is used (archives often contain a top-level folder)
func extensionRoot(dir string) (string, error) {
	if _, err := os.Stat(filepath.Join(dir, fileAlias)); err == nil {
		return dir, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	if len(entries) == 1 && entries[0].IsDir() {
		sub := filepath.Join(dir, entries[0].Name())
		if _, err := os.Stat(filepath.Join(sub, fileAlias)); err == nil {
			return sub, nil
		}
	}
	return "", fmt.Errorf("archive does not contain %s", fileAlias)
}
