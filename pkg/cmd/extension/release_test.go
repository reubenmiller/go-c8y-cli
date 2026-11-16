package extension

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assets(names ...string) []releaseAsset {
	items := make([]releaseAsset, 0, len(names))
	for _, name := range names {
		items = append(items, releaseAsset{Name: name, APIURL: "https://example.invalid/" + name})
	}
	return items
}

func TestSelectAsset(t *testing.T) {
	cases := []struct {
		name      string
		assets    []string
		goos      string
		goarch    string
		rosetta   bool
		want      string
		wantType  assetType
		wantNoARM bool
	}{
		{
			name:     "goreleaser archive",
			assets:   []string{"checksums.txt", "c8y-foo_1.2.0_darwin_arm64.tar.gz", "c8y-foo_1.2.0_linux_amd64.tar.gz", "c8y-foo_1.2.0_linux_arm64.tar.gz"},
			goos:     "linux",
			goarch:   "amd64",
			want:     "c8y-foo_1.2.0_linux_amd64.tar.gz",
			wantType: assetTarGz,
		},
		{
			name:     "zip on windows",
			assets:   []string{"c8y-foo_1.2.0_linux_amd64.tar.gz", "c8y-foo_1.2.0_windows_amd64.zip"},
			goos:     "windows",
			goarch:   "amd64",
			want:     "c8y-foo_1.2.0_windows_amd64.zip",
			wantType: assetZip,
		},
		{
			name:     "case insensitive and alias arch",
			assets:   []string{"c8y-foo_Linux_x86_64.tar.gz"},
			goos:     "linux",
			goarch:   "amd64",
			want:     "c8y-foo_Linux_x86_64.tar.gz",
			wantType: assetTarGz,
		},
		{
			name:     "arm does not match arm64",
			assets:   []string{"c8y-foo_linux_arm64.tar.gz", "c8y-foo_linux_arm.tar.gz"},
			goos:     "linux",
			goarch:   "arm",
			want:     "c8y-foo_linux_arm.tar.gz",
			wantType: assetTarGz,
		},
		{
			name:     "archive preferred over bare binary",
			assets:   []string{"c8y-foo-darwin-arm64", "c8y-foo_darwin_arm64.tar.gz"},
			goos:     "darwin",
			goarch:   "arm64",
			want:     "c8y-foo_darwin_arm64.tar.gz",
			wantType: assetTarGz,
		},
		{
			name:     "bare binary",
			assets:   []string{"c8y-foo-linux-amd64", "c8y-foo-windows-amd64.exe"},
			goos:     "windows",
			goarch:   "amd64",
			want:     "c8y-foo-windows-amd64.exe",
			wantType: assetBinary,
		},
		{
			name:     "bare binary on windows requires exe",
			assets:   []string{"c8y-foo-windows-amd64"},
			goos:     "windows",
			goarch:   "amd64",
			wantType: assetNone,
		},
		{
			name:      "rosetta fallback",
			assets:    []string{"c8y-foo_darwin_amd64.tar.gz"},
			goos:      "darwin",
			goarch:    "arm64",
			rosetta:   true,
			want:      "c8y-foo_darwin_amd64.tar.gz",
			wantType:  assetTarGz,
			wantNoARM: true,
		},
		{
			name:     "no rosetta",
			assets:   []string{"c8y-foo_darwin_amd64.tar.gz"},
			goos:     "darwin",
			goarch:   "arm64",
			wantType: assetNone,
		},
		{
			name:     "missing platform",
			assets:   []string{"c8y-foo_linux_amd64.tar.gz", "source.tar.gz"},
			goos:     "freebsd",
			goarch:   "amd64",
			wantType: assetNone,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			asset, kind, trueARM := selectAsset(assets(tc.assets...), tc.goos, tc.goarch, tc.rosetta)
			assert.Equal(t, tc.wantType, kind)
			if tc.want == "" {
				assert.Nil(t, asset)
				return
			}
			require.NotNil(t, asset)
			assert.Equal(t, tc.want, asset.Name)
			if tc.goos == "darwin" && tc.goarch == "arm64" {
				assert.Equal(t, !tc.wantNoARM, trueARM)
			}
		})
	}
}

func TestHasPlatformAssets(t *testing.T) {
	assert.True(t, hasPlatformAssets(&release{Assets: assets("c8y-foo_1.0.0_linux_amd64.tar.gz")}))
	assert.True(t, hasPlatformAssets(&release{Assets: assets("c8y-foo-windows-arm64.exe")}))
	assert.False(t, hasPlatformAssets(&release{Assets: assets("source.zip", "checksums.txt", "notes.md")}))
	assert.False(t, hasPlatformAssets(&release{}))
}

func TestChecksums(t *testing.T) {
	assert.Equal(t, "checksums.txt", findChecksumAsset(assets("a.tar.gz", "checksums.txt")).Name)
	assert.Equal(t, "c8y-foo_1.0.0_checksums.txt", findChecksumAsset(assets("c8y-foo_1.0.0_checksums.txt")).Name)
	assert.Nil(t, findChecksumAsset(assets("a.tar.gz")))

	checksums, err := parseChecksums(strings.NewReader("ABCDEF  c8y-foo_linux_amd64.tar.gz\n123456 *c8y-foo.zip\n\ninvalid\n"))
	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"c8y-foo_linux_amd64.tar.gz": "abcdef",
		"c8y-foo.zip":                "123456",
	}, checksums)
}

type archiveEntry struct {
	Name     string
	Body     string
	Mode     int64
	Linkname string
	Type     byte
}

func makeTarGz(t *testing.T, entries []archiveEntry) string {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		typ := e.Type
		if typ == 0 {
			typ = tar.TypeReg
		}
		mode := e.Mode
		if mode == 0 {
			mode = 0644
		}
		hdr := &tar.Header{Name: e.Name, Mode: mode, Size: int64(len(e.Body)), Typeflag: typ, Linkname: e.Linkname}
		if typ != tar.TypeReg {
			hdr.Size = 0
		}
		require.NoError(t, tw.WriteHeader(hdr))
		if typ == tar.TypeReg {
			_, err := tw.Write([]byte(e.Body))
			require.NoError(t, err)
		}
	}
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	path := filepath.Join(t.TempDir(), "asset.tar.gz")
	require.NoError(t, os.WriteFile(path, buf.Bytes(), 0644))
	return path
}

func makeZip(t *testing.T, entries []archiveEntry) string {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range entries {
		hdr := &zip.FileHeader{Name: e.Name, Method: zip.Deflate}
		mode := os.FileMode(e.Mode)
		if mode == 0 {
			mode = 0644
		}
		if e.Type == tar.TypeSymlink {
			mode = os.ModeSymlink | 0777
		}
		hdr.SetMode(mode)
		w, err := zw.CreateHeader(hdr)
		require.NoError(t, err)
		body := e.Body
		if e.Type == tar.TypeSymlink {
			body = e.Linkname
		}
		_, err = w.Write([]byte(body))
		require.NoError(t, err)
	}
	require.NoError(t, zw.Close())
	path := filepath.Join(t.TempDir(), "asset.zip")
	require.NoError(t, os.WriteFile(path, buf.Bytes(), 0644))
	return path
}

func TestExtractArchive(t *testing.T) {
	entries := []archiveEntry{
		{Name: "c8y-foo/", Type: tar.TypeDir},
		{Name: "c8y-foo/extension.yaml", Body: "binary: {}\n"},
		{Name: "c8y-foo/bin/c8y-foo", Body: "#!/bin/sh\necho hi\n", Mode: 0755},
		{Name: "c8y-foo/views/foo.json", Body: "{}"},
		{Name: "c8y-foo/views/link.json", Type: tar.TypeSymlink, Linkname: "foo.json"},
	}

	for name, archive := range map[string]func() (string, assetType){
		"tar.gz": func() (string, assetType) { return makeTarGz(t, entries), assetTarGz },
		"zip":    func() (string, assetType) { return makeZip(t, entries), assetZip },
	} {
		t.Run(name, func(t *testing.T) {
			src, kind := archive()
			dest := filepath.Join(t.TempDir(), "content")
			require.NoError(t, extractArchive(src, kind, dest))

			root, err := extensionRoot(dest)
			require.NoError(t, err)
			assert.Equal(t, filepath.Join(dest, "c8y-foo"), root)

			b, err := os.ReadFile(filepath.Join(root, "views", "link.json"))
			require.NoError(t, err)
			assert.Equal(t, "{}", string(b))

			info, err := os.Stat(filepath.Join(root, "bin", "c8y-foo"))
			require.NoError(t, err)
			if os.PathSeparator == '/' {
				assert.NotZero(t, info.Mode().Perm()&0100, "binary should be executable")
			}
		})
	}
}

func TestExtractArchiveRejectsUnsafeEntries(t *testing.T) {
	cases := map[string][]archiveEntry{
		"path traversal":   {{Name: "../../.bashrc", Body: "echo pwned"}},
		"nested traversal": {{Name: "views/../../outside", Body: "x"}},
		"absolute path":    {{Name: "/etc/passwd", Body: "x"}},
		"symlink outside":  {{Name: "link", Type: tar.TypeSymlink, Linkname: "../../outside"}},
		"absolute symlink": {{Name: "link", Type: tar.TypeSymlink, Linkname: "/etc/passwd"}},
		"hard link":        {{Name: "link", Type: tar.TypeLink, Linkname: "extension.yaml"}},
	}
	for name, entries := range cases {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			dest := filepath.Join(parent, "a", "b", "content")
			err := extractArchive(makeTarGz(t, entries), assetTarGz, dest)
			require.Error(t, err)
			assert.Contains(t, err.Error(), entries[0].Name)

			// nothing written outside of the destination
			files := []string{}
			_ = filepath.Walk(parent, func(path string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() {
					files = append(files, path)
				}
				return nil
			})
			assert.Empty(t, files)
		})
	}
}

func TestExtractArchiveSizeLimit(t *testing.T) {
	previous := maxExtractedSize
	maxExtractedSize = 10
	defer func() { maxExtractedSize = previous }()

	src := makeTarGz(t, []archiveEntry{{Name: "big", Body: strings.Repeat("x", 11)}})
	err := extractArchive(src, assetTarGz, t.TempDir())
	assert.True(t, errors.Is(err, ErrArchiveTooLarge), "got %v", err)

	src = makeTarGz(t, []archiveEntry{{Name: "ok", Body: strings.Repeat("x", 10)}})
	assert.NoError(t, extractArchive(src, assetTarGz, t.TempDir()))
}

func TestExtensionRootRequiresManifest(t *testing.T) {
	dest := t.TempDir()
	require.NoError(t, extractArchive(makeTarGz(t, []archiveEntry{{Name: "bin/c8y-foo", Body: "x"}}), assetTarGz, dest))
	_, err := extensionRoot(dest)
	assert.ErrorContains(t, err, "extension.yaml")
}
