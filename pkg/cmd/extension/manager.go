package extension

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"text/template"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/ghrepo"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/findsh"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/git"

	"github.com/cli/safeexec"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v3"
)

type Manager struct {
	dataDir     func() string
	systemDirs  func() []string
	lookPath    func(string) (string, error)
	findSh      func() (string, error)
	newCommand  func(string, ...string) *exec.Cmd
	platform    func() (string, string) // returns GOOS and GOARCH
	rosetta     func() bool
	codesign    func(string) error
	client      *http.Client
	config      config.Config
	io          *iostreams.IOStreams
	dryRunMode  bool
	allowBinary bool
}

func NewManager(ios *iostreams.IOStreams, cfg *config.Config) *Manager {
	return &Manager{
		dataDir:    cfg.ExtensionsDataDir,
		systemDirs: cfg.ExtensionsSystemDirs,
		lookPath:   safeexec.LookPath,
		findSh:     findsh.Find,
		newCommand: exec.Command,
		dryRunMode: cfg.DryRun(),
		platform: func() (string, string) {
			return runtime.GOOS, runtime.GOARCH
		},
		rosetta:     hasRosetta,
		codesign:    codesignBinary,
		allowBinary: true,
		io:          ios,
	}
}

func (m *Manager) SetConfig(cfg config.Config) {
	m.config = cfg
}

func (m *Manager) SetClient(client *http.Client) {
	m.client = client
}

func (m *Manager) EnableDryRunMode() {
	m.dryRunMode = true
}

// Execute runs an extension executable. Scripts are run via sh on Windows to support shebang lines
func (m *Manager) Execute(opts extensions.ExecOptions) error {
	if opts.Executable == "" {
		return errors.New("extension executable is empty")
	}

	args := opts.Args
	exe := opts.Executable
	if !opts.IsBinary && runtime.GOOS == "windows" {
		shExe, err := m.findSh()
		if err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				return errors.New("the `sh.exe` interpreter is required. Please install Git for Windows and try again")
			}
			return err
		}
		args = append([]string{"-c", `command "$@"`, "--", exe}, args...)
		exe = shExe
	}

	externalCmd := m.newCommand(exe, args...)
	externalCmd.Env = opts.Env
	externalCmd.Stdin = opts.Stdin
	externalCmd.Stdout = opts.Stdout
	externalCmd.Stderr = opts.Stderr
	return externalCmd.Run()
}

func (m *Manager) List() []extensions.Extension {
	exts, _ := m.list(false)
	r := make([]extensions.Extension, len(exts))
	for i, v := range exts {
		val := v
		r[i] = &val
	}
	return r
}

func (m *Manager) list(includeMetadata bool) ([]Extension, error) {
	dir := m.installDir()
	entries, err := os.ReadDir(dir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	var results []Extension
	for _, f := range entries {
		// Ignore hidden folders, e.g. temporary install folders
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}
		if !strings.HasPrefix(f.Name(), ExtPrefix) && !strings.Contains(f.Name(), ExtPrefix) {
			continue
		}
		var ext Extension
		var err error
		if f.IsDir() {
			ext, err = m.parseExtensionDir(f)
		} else {
			ext, err = m.parseExtensionFile(f)
		}
		if err != nil {
			return nil, err
		}
		ext.manifest = &manifestCache{}
		results = append(results, ext)
	}

	results = append(results, m.listSystemExtensions(results)...)

	if includeMetadata {
		m.populateLatestVersions(results)
	}

	return results, nil
}

// listSystemExtensions returns the extensions in the system-wide extension folders. These are installed
// by package managers, and are only read. Extensions in the user's extension folder (existing) and earlier
// system folders take precedence over extensions with the same name.
func (m *Manager) listSystemExtensions(existing []Extension) []Extension {
	if m.systemDirs == nil {
		return nil
	}
	seen := map[string]bool{}
	for _, ext := range existing {
		seen[ext.Name()] = true
	}

	installDir := filepath.Clean(m.installDir())
	results := []Extension{}
	for _, dir := range m.systemDirs() {
		if filepath.Clean(dir) == installDir {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			name := entry.Name()
			if strings.HasPrefix(name, ".") || !strings.HasPrefix(name, ExtPrefix) {
				continue
			}
			path := filepath.Join(dir, name)
			// Follow symlinks, e.g. Homebrew links the folders from the Cellar
			if info, err := os.Stat(path); err != nil || !info.IsDir() {
				continue
			}
			ext := Extension{path: path, isSystem: true, manifest: &manifestCache{}}
			if seen[ext.Name()] {
				continue
			}
			seen[ext.Name()] = true

			// Package managers can ship the manifest of a binary install, which includes the version
			if b, err := os.ReadFile(filepath.Join(path, manifestName)); err == nil {
				var bm binManifest
				if yaml.Unmarshal(b, &bm) == nil {
					ext.currentVersion = bm.Tag
					if bm.Owner != "" && bm.Name != "" {
						ext.url = ghrepo.GenerateRepoURL(ghrepo.NewWithHost(bm.Owner, bm.Name, bm.Host, ""), "")
					}
				}
			}
			results = append(results, ext)
		}
	}
	return results
}

func (m *Manager) parseExtensionFile(fi fs.DirEntry) (Extension, error) {
	ext := Extension{isLocal: true}
	id := m.installDir()
	exePath := filepath.Join(id, fi.Name())
	if !isSymlink(fi.Type()) {
		// if this is a regular file, its contents is the local directory of the extension
		p, err := readPathFromFile(filepath.Join(id, fi.Name()))
		if err != nil {
			return ext, err
		}
		exePath = p
	}
	ext.path = exePath
	return ext, nil
}

func (m *Manager) parseExtensionDir(fi fs.DirEntry) (Extension, error) {
	id := m.installDir()
	if _, err := os.Stat(filepath.Join(id, fi.Name(), manifestName)); err == nil {
		return m.parseBinaryExtensionDir(fi)
	}

	return m.parseGitExtensionDir(fi)
}

func (m *Manager) parseBinaryExtensionDir(fi fs.DirEntry) (Extension, error) {
	id := m.installDir()
	ext := Extension{path: filepath.Join(id, fi.Name()), kind: BinaryKind}
	manifestPath := filepath.Join(id, fi.Name(), manifestName)
	manifest, err := os.ReadFile(manifestPath)
	if err != nil {
		return ext, fmt.Errorf("could not open %s for reading: %w", manifestPath, err)
	}
	var bm binManifest
	err = yaml.Unmarshal(manifest, &bm)
	if err != nil {
		return ext, fmt.Errorf("could not parse %s: %w", manifestPath, err)
	}
	repo := ghrepo.NewWithHost(bm.Owner, bm.Name, bm.Host, "")
	remoteURL := ghrepo.GenerateRepoURL(repo, "")
	ext.url = remoteURL
	ext.currentVersion = bm.Tag
	ext.isPinned = bm.IsPinned
	return ext, nil
}

func (m *Manager) parseGitExtensionDir(fi fs.DirEntry) (Extension, error) {
	id := m.installDir()
	exePath := filepath.Join(id, fi.Name())
	remoteUrl := m.getRemoteUrl(fi.Name())
	currentVersion := m.getCurrentVersion(fi.Name())

	var isPinned bool
	pinPath := filepath.Join(id, fi.Name(), fmt.Sprintf(".pin-%s", currentVersion))
	if _, err := os.Stat(pinPath); err == nil {
		isPinned = true
	}

	return Extension{
		path:           exePath,
		url:            remoteUrl,
		isLocal:        false,
		currentVersion: currentVersion,
		kind:           GitKind,
		isPinned:       isPinned,
	}, nil
}

// getCurrentVersion determines the current version for non-local git extensions.
func (m *Manager) getCurrentVersion(extension string) string {
	gitExe, err := m.lookPath("git")
	if err != nil {
		return ""
	}
	dir := m.installDir()
	gitDir := "--git-dir=" + filepath.Join(dir, extension, ".git")
	cmd := m.newCommand(gitExe, gitDir, "rev-parse", "HEAD")

	localSha, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(bytes.TrimSpace(localSha))
}

// getRemoteUrl determines the remote URL for non-local git extensions.
func (m *Manager) getRemoteUrl(extension string) string {
	gitExe, err := m.lookPath("git")
	if err != nil {
		return ""
	}
	dir := m.installDir()
	gitDir := "--git-dir=" + filepath.Join(dir, extension, ".git")
	cmd := m.newCommand(gitExe, gitDir, "config", "remote.origin.url")
	url, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(url))
}

func (m *Manager) populateLatestVersions(exts []Extension) {
	size := len(exts)
	type result struct {
		index   int
		version string
	}
	ch := make(chan result, size)
	var wg sync.WaitGroup
	wg.Add(size)
	for idx, ext := range exts {
		go func(i int, e Extension) {
			defer wg.Done()
			version, _ := m.getLatestVersion(e)
			ch <- result{index: i, version: version}
		}(idx, ext)
	}
	wg.Wait()
	close(ch)
	for r := range ch {
		ext := &exts[r.index]
		ext.latestVersion = r.version
	}
}

func (m *Manager) getLatestVersion(ext Extension) (string, error) {
	if ext.isSystem {
		return "", ErrSystemExtensionUpgrade
	}
	if ext.isLocal {
		return "", ErrLocalExtensionUpgrade
	}
	if ext.IsBinary() {
		repo, err := ghrepo.FromFullName(ext.url)
		if err != nil {
			return "", err
		}
		r, err := fetchLatestRelease(m.client, repo)
		if err != nil {
			return "", err
		}
		return r.Tag, nil
	} else {
		gitExe, err := m.lookPath("git")
		if err != nil {
			return "", err
		}
		extDir := ext.path
		gitDir := "--git-dir=" + filepath.Join(extDir, ".git")
		cmd := m.newCommand(gitExe, gitDir, "ls-remote", "origin", "HEAD")
		lsRemote, err := cmd.Output()
		if err != nil {
			return "", err
		}
		remoteSha := bytes.SplitN(lsRemote, []byte("\t"), 2)[0]
		return string(remoteSha), nil
	}
}

func (m *Manager) InstallLocal(dir string, opts extensions.InstallOptions) (string, error) {
	folder, err := resolveInstallFolder(opts, dir, filepath.Base(dir), false)
	if err != nil {
		return "", err
	}
	name := strings.TrimPrefix(folder, ExtPrefix)
	targetLink := filepath.Join(m.installDir(), folder)

	// Installing the same folder again replaces the link, e.g. after changing the extension
	if target, ok := localLinkTarget(targetLink); ok && sameDir(target, dir) {
		if m.dryRunMode {
			return name, nil
		}
		if err := os.Remove(targetLink); err != nil {
			return "", err
		}
	} else if opts.CheckName != nil {
		if err := opts.CheckName(name); err != nil {
			return "", err
		}
	}

	if m.dryRunMode {
		return name, nil
	}
	if err := os.MkdirAll(filepath.Dir(targetLink), 0755); err != nil {
		return "", err
	}
	return name, makeSymlink(dir, targetLink)
}

// localLinkTarget returns the folder which a local extension link points to: the target of a
// symlink, or the path in the file which is used instead of a symlink on Windows
func localLinkTarget(link string) (string, bool) {
	info, err := os.Lstat(link)
	if err != nil {
		return "", false
	}
	if isSymlink(info.Mode().Type()) {
		target, err := filepath.EvalSymlinks(link)
		return target, err == nil
	}
	if info.Mode().IsRegular() {
		target, err := readPathFromFile(link)
		return target, err == nil
	}
	return "", false
}

// sameDir reports whether both paths are the same folder, after resolving symlinks
func sameDir(a, b string) bool {
	resolve := func(p string) string {
		if r, err := filepath.EvalSymlinks(p); err == nil {
			p = r
		}
		if abs, err := filepath.Abs(p); err == nil {
			p = abs
		}
		return filepath.Clean(p)
	}
	return resolve(a) == resolve(b)
}

// ErrInvalidRepositoryName is returned if the extension name is taken from the repository name, which doesn't have the c8y- prefix
var ErrInvalidRepositoryName = errors.New("invalid extension repository name")

// resolveInstallFolder returns the folder name of an extension in the extension directory. It is
// c8y-<name> for the name set by the user (--name) or declared in the extension.yaml file in dir.
// Otherwise it is the fallback (the repository or folder name). For remote installs the fallback
// must contain the c8y- prefix, and for local installs the prefix is added
func resolveInstallFolder(opts extensions.InstallOptions, dir string, fallback string, requirePrefix bool) (string, error) {
	if opts.Name != "" {
		if strings.HasPrefix(opts.Name, ExtPrefix) {
			return opts.Name, nil
		}
		return ExtPrefix + opts.Name, nil
	}

	declared, err := readDeclaredName(dir)
	if err != nil {
		return "", err
	}
	if declared != "" {
		name, err := extensions.NormalizeName(declared)
		if err != nil {
			return "", fmt.Errorf("%w. Fix the name in %s, or install the extension with --name", err, fileAlias)
		}
		return ExtPrefix + name, nil
	}

	if strings.HasPrefix(fallback, ExtPrefix) || strings.Contains(fallback, ExtPrefix) {
		return fallback, nil
	}
	if requirePrefix {
		return "", fmt.Errorf("%w %q. The repository name must start with %q, or its %s must set name (or install it with --name)", ErrInvalidRepositoryName, fallback, ExtPrefix, fileAlias)
	}
	return ExtPrefix + fallback, nil
}

// checkInstallName runs the name check of the install options
func checkInstallName(opts extensions.InstallOptions, folder string) error {
	if opts.CheckName == nil {
		return nil
	}
	return opts.CheckName(strings.TrimPrefix(folder, ExtPrefix))
}

type binManifest struct {
	Owner    string
	Name     string
	Host     string
	Tag      string
	IsPinned bool
	// TODO I may end up not using this; just thinking ahead to local installs
	Path string

	// Asset is the name of the installed release asset
	Asset string `yaml:"asset,omitempty"`
	// SHA256 is the verified checksum of the release asset (empty if the release did not publish checksums)
	SHA256 string `yaml:"sha256,omitempty"`
}

// isGitHubHost returns true if releases can be checked on the host via the GitHub API
func isGitHubHost(host string) bool {
	return strings.Contains(strings.ToLower(host), "github")
}

// Install installs an extension from repo, and pins to a release tag or commit hash if provided.
// GitHub repositories which publish release assets for the current platform are installed
// as binary extensions, all others are installed using git clone
func (m *Manager) Install(repo ghrepo.Interface, opts extensions.InstallOptions) (string, error) {
	if m.allowBinary && isGitHubHost(repo.RepoHost()) {
		r, err := m.fetchRelease(repo, opts.Pin)
		switch {
		case err == nil:
			if hasPlatformAssets(r) {
				return m.installBinRelease(repo, binInstall{opts: opts}, r, opts.Pin != "")
			}
		case errors.Is(err, ErrReleaseNotFound):
			// No release, so install the repository using git
		default:
			// Don't block git installs (e.g. private repositories or API rate limits)
			fmt.Fprintf(m.io.ErrOut, "could not check the releases of %s, installing with git. %s\n", ghrepo.FullName(repo), err)
		}
	}

	return m.installGit(repo, opts, m.io.Out, m.io.ErrOut)
}

// fetchRelease returns the release with the given tag, or the latest release if tag is empty
func (m *Manager) fetchRelease(repo ghrepo.Interface, tag string) (*release, error) {
	if tag != "" {
		return fetchReleaseFromTag(m.client, repo, tag)
	}
	return fetchLatestRelease(m.client, repo)
}

// installBin installs the latest release (or the release with the given tag) as a binary extension
// The folder is the folder name of the installed extension
func (m *Manager) installBin(repo ghrepo.Interface, folder string, tag string) error {
	r, err := m.fetchRelease(repo, tag)
	if err != nil {
		return err
	}
	_, err = m.installBinRelease(repo, binInstall{folder: folder}, r, tag != "")
	return err
}

// binInstall is where a binary extension is installed: in a fixed folder (updates), or in the
// folder resolved from the install options and the extension's manifest (installs)
type binInstall struct {
	folder string
	opts   extensions.InstallOptions
}

// installBinRelease installs a release as a binary extension. The release asset is downloaded,
// verified and extracted into a temporary folder in the data directory, and only then moved into
// place, so a failure leaves any previously installed version untouched
func (m *Manager) installBinRelease(repo ghrepo.Interface, target binInstall, r *release, isPinned bool) (string, error) {
	goos, goarch := m.platform()
	asset, kind, trueARMBinary := selectAsset(r.Assets, goos, goarch, m.rosetta())
	if asset == nil {
		return "", fmt.Errorf("%w. repository=%s, release=%s, platform=%s/%s", ErrNoPlatformAsset, ghrepo.FullName(repo), r.Tag, goos, goarch)
	}

	// resolveFolder returns the folder of the extension, using the manifest in dir for new installs
	resolveFolder := func(dir string) (string, error) {
		if target.folder != "" {
			return target.folder, nil
		}
		folder, err := resolveInstallFolder(target.opts, dir, repo.RepoName(), true)
		if err != nil {
			return "", err
		}
		return folder, checkInstallName(target.opts, folder)
	}

	if m.dryRunMode {
		m.printDryRunInstall(fmt.Sprintf("%s %s (asset %s)", ghrepo.FullName(repo), r.Tag, asset.Name), target.folder, target.opts, repo.RepoName())
		return "", nil
	}

	if err := os.MkdirAll(m.installDir(), 0755); err != nil {
		return "", fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Hidden folder (ignored when listing extensions), on the same file system as the target
	// so it can be renamed into place
	tmpDir, err := os.MkdirTemp(m.installDir(), ".tmp-install-")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	downloadPath := filepath.Join(tmpDir, "download")
	if err := downloadAsset(m.client, repo.RepoHost(), *asset, downloadPath); err != nil {
		return "", fmt.Errorf("failed to download asset %s: %w", asset.Name, err)
	}

	checksum, err := fileSHA256(downloadPath)
	if err != nil {
		return "", err
	}
	verifiedChecksum := ""
	if checksumAsset := findChecksumAsset(r.Assets); checksumAsset != nil {
		checksums, err := downloadChecksums(m.client, repo.RepoHost(), *checksumAsset)
		if err != nil {
			return "", fmt.Errorf("failed to download checksums %s: %w", checksumAsset.Name, err)
		}
		expected, ok := checksums[asset.Name]
		if !ok {
			return "", fmt.Errorf("%w. %s does not contain an entry for %s", ErrChecksumMismatch, checksumAsset.Name, asset.Name)
		}
		if !strings.EqualFold(expected, checksum) {
			return "", fmt.Errorf("%w. asset=%s, expected=%s, got=%s", ErrChecksumMismatch, asset.Name, expected, checksum)
		}
		verifiedChecksum = checksum
	} else {
		fmt.Fprintf(m.io.ErrOut, "the release does not publish checksums, so %s was not verified\n", asset.Name)
	}

	contentDir := filepath.Join(tmpDir, "content")
	folder := ""
	if kind.IsArchive() {
		if err := extractArchive(downloadPath, kind, contentDir); err != nil {
			return "", fmt.Errorf("failed to extract asset %s: %w", asset.Name, err)
		}
		root, err := extensionRoot(contentDir)
		if err != nil {
			return "", fmt.Errorf("invalid extension archive %s: %w", asset.Name, err)
		}
		contentDir = root
	} else {
		// Bare binary (gh compatible), so generate a minimal extension manifest
		// The extension has no manifest, so the folder is the name set by the user or the repository name
		folder, err = resolveFolder(contentDir)
		if err != nil {
			return "", err
		}
		binName := folder
		if goos == "windows" {
			binName += ".exe"
		}
		if err := os.MkdirAll(contentDir, 0755); err != nil {
			return "", err
		}
		if err := os.Rename(downloadPath, filepath.Join(contentDir, binName)); err != nil {
			return "", err
		}
		manifestFile, err := yaml.Marshal(&ExtensionFile{
			Manifest: extensions.Manifest{
				Version: "v1",
				Binary:  &extensions.BinarySpec{Path: binName},
			},
		})
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(filepath.Join(contentDir, fileAlias), manifestFile, 0644); err != nil {
			return "", err
		}
	}

	exePath := ""
	if manifest, err := readManifest(contentDir); err != nil {
		return "", err
	} else if manifest.Binary != nil {
		// The default binary path uses the repository name (the extension can be installed under a different name)
		binPath := manifest.Binary.Path
		if binPath == "" {
			binPath = filepath.Join("bin", repo.RepoName())
		}
		exePath = resolveExecutable(contentDir, binPath, goos)
		if _, err := os.Stat(exePath); err != nil {
			return "", fmt.Errorf("the binary declared in %s is missing from asset %s: %w", fileAlias, asset.Name, err)
		}
		if goos != "windows" {
			if err := os.Chmod(exePath, 0755); err != nil {
				return "", err
			}
		}
		if trueARMBinary {
			if err := m.codesign(exePath); err != nil {
				return "", fmt.Errorf("failed to codesign downloaded binary: %w", err)
			}
		}
	}

	manifest := binManifest{
		Name:     repo.RepoName(),
		Owner:    repo.RepoOwner(),
		Host:     repo.RepoHost(),
		Tag:      r.Tag,
		IsPinned: isPinned,
		Asset:    asset.Name,
		SHA256:   verifiedChecksum,
	}
	if exePath != "" {
		if rel, err := filepath.Rel(contentDir, exePath); err == nil {
			manifest.Path = filepath.ToSlash(rel)
		}
	}

	bs, err := yaml.Marshal(manifest)
	if err != nil {
		return "", fmt.Errorf("failed to serialize manifest: %w", err)
	}
	if err := os.WriteFile(filepath.Join(contentDir, manifestName), bs, 0600); err != nil {
		return "", fmt.Errorf("failed write manifest file: %w", err)
	}

	if folder == "" {
		folder, err = resolveFolder(contentDir)
		if err != nil {
			return "", err
		}
	}
	if err := replaceDir(contentDir, filepath.Join(m.installDir(), folder), filepath.Join(tmpDir, "previous")); err != nil {
		return "", err
	}
	return strings.TrimPrefix(folder, ExtPrefix), nil
}

// replaceDir moves src to dst. An existing dst is moved to backup first, and restored if the move fails
func replaceDir(src, dst, backup string) error {
	hasPrevious := false
	if _, err := os.Lstat(dst); err == nil {
		if err := os.Rename(dst, backup); err != nil {
			return fmt.Errorf("failed to replace existing extension: %w", err)
		}
		hasPrevious = true
	}
	if err := os.Rename(src, dst); err != nil {
		if hasPrevious {
			_ = os.Rename(backup, dst)
		}
		return fmt.Errorf("failed to install extension: %w", err)
	}
	return nil
}

func (m *Manager) installGit(repo ghrepo.Interface, opts extensions.InstallOptions, stdout, stderr io.Writer) (string, error) {
	protocol := repo.RepoHost()
	cloneURL := ghrepo.FormatRemoteURL(repo, protocol)
	repoName := strings.TrimSuffix(path.Base(cloneURL), ".git")

	exe, err := m.lookPath("git")
	if err != nil {
		return "", err
	}

	var commitSHA string
	if opts.Pin != "" {
		commitSHA, err = fetchCommitSHA(m.client, repo, opts.Pin)
		if err != nil {
			return "", err
		}
	}

	if m.dryRunMode {
		m.printDryRunInstall(cloneURL, "", opts, repoName)
		return "", nil
	}

	if err := os.MkdirAll(m.installDir(), 0755); err != nil {
		return "", fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Clone into a hidden folder (ignored when listing extensions), on the same file system as the
	// target, so the name can be read from the extension's manifest before it is moved into place
	tmpDir, err := os.MkdirTemp(m.installDir(), ".tmp-install-")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	cloneDir := filepath.Join(tmpDir, "repo")

	externalCmd := m.newCommand(exe, "clone", cloneURL, cloneDir)
	externalCmd.Stdout = stdout
	externalCmd.Stderr = stderr
	if err := externalCmd.Run(); err != nil {
		return "", err
	}

	if commitSHA != "" {
		checkoutCmd := m.newCommand(exe, "-C", cloneDir, "checkout", commitSHA)
		checkoutCmd.Stdout = stdout
		checkoutCmd.Stderr = stderr
		if err := checkoutCmd.Run(); err != nil {
			return "", err
		}

		pinPath := filepath.Join(cloneDir, fmt.Sprintf(".pin-%s", commitSHA))
		f, err := os.OpenFile(pinPath, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return "", fmt.Errorf("failed to create pin file in directory: %w", err)
		}
		if err := f.Close(); err != nil {
			return "", err
		}
	}

	folder, err := resolveInstallFolder(opts, cloneDir, repoName, true)
	if err != nil {
		return "", err
	}
	if err := checkInstallName(opts, folder); err != nil {
		return "", err
	}
	if err := replaceDir(cloneDir, filepath.Join(m.installDir(), folder), filepath.Join(tmpDir, "previous")); err != nil {
		return "", err
	}
	return strings.TrimPrefix(folder, ExtPrefix), nil
}

// printDryRunInstall prints what an install would do. The folder is empty for new installs, where the
// name set by the user, or the name declared in the extension's manifest, or the repository name is used
func (m *Manager) printDryRunInstall(source string, folder string, opts extensions.InstallOptions, repoName string) {
	if folder != "" {
		fmt.Fprintf(m.io.Out, "would install %s to %s\n", source, filepath.Join(m.installDir(), folder))
		return
	}
	if opts.Name != "" {
		name := opts.Name
		if !strings.HasPrefix(name, ExtPrefix) {
			name = ExtPrefix + name
		}
		fmt.Fprintf(m.io.Out, "would install %s to %s\n", source, filepath.Join(m.installDir(), name))
		return
	}
	fmt.Fprintf(m.io.Out, "would install %s to %s, or to %s if its %s declares a name\n",
		source, filepath.Join(m.installDir(), repoName), filepath.Join(m.installDir(), ExtPrefix+"<name>"), fileAlias)
}

var ErrNoPlatformAsset = errors.New("the release does not have an asset for this platform")
var ErrPinnedExtensionUpgrade = errors.New("pinned extensions can not be updated")
var ErrLocalExtensionUpgrade = errors.New("local extensions can not be updated")
var ErrSystemExtensionUpgrade = errors.New("extensions in a system extensions folder can only be updated by the package manager which installed them")
var ErrUpToDate = errors.New("already up to date")
var ErrNoExtensionsInstalled = errors.New("no extensions installed")

func (m *Manager) Upgrade(name string, force bool) error {
	// Fetch metadata during list only when upgrading all extensions.
	// This is a performance improvement so that we don't make a
	// bunch of unnecessary network requests when trying to upgrade a single extension.
	fetchMetadata := name == ""
	exts, _ := m.list(fetchMetadata)
	if len(exts) == 0 {
		return ErrNoExtensionsInstalled
	}
	if name == "" {
		return m.upgradeExtensions(exts, force)
	}
	for _, f := range exts {
		if f.Name() != name {
			continue
		}
		var err error
		// For single extensions manually retrieve latest version since we forgo
		// doing it during list.
		f.latestVersion, err = m.getLatestVersion(f)
		if err != nil {
			return err
		}
		return m.upgradeExtension(f, force)
	}
	return fmt.Errorf("no extension matched %q", name)
}

func (m *Manager) upgradeExtensions(exts []Extension, force bool) error {
	var failed bool
	for _, f := range exts {
		fmt.Fprintf(m.io.Out, "[%s]: ", f.Name())
		err := m.upgradeExtension(f, force)
		if err != nil {
			if !errors.Is(err, ErrLocalExtensionUpgrade) &&
				!errors.Is(err, ErrSystemExtensionUpgrade) &&
				!errors.Is(err, ErrUpToDate) &&
				!errors.Is(err, ErrPinnedExtensionUpgrade) {
				failed = true
			}
			fmt.Fprintf(m.io.Out, "%s\n", err)
			continue
		}
		currentVersion := displayExtensionVersion(&f, f.currentVersion)
		latestVersion := displayExtensionVersion(&f, f.latestVersion)
		if m.dryRunMode {
			fmt.Fprintf(m.io.Out, "would have upgraded from %s to %s\n", currentVersion, latestVersion)
		} else {
			fmt.Fprintf(m.io.Out, "upgraded from %s to %s\n", currentVersion, latestVersion)
		}
	}
	if failed {
		return errors.New("some extensions failed to upgrade")
	}
	return nil
}

func (m *Manager) upgradeExtension(ext Extension, force bool) error {
	if ext.isSystem {
		return ErrSystemExtensionUpgrade
	}
	if ext.isLocal {
		return ErrLocalExtensionUpgrade
	}
	if ext.IsPinned() {
		return ErrPinnedExtensionUpgrade
	}
	if !ext.UpdateAvailable() {
		if version := displayExtensionVersion(&ext, ext.currentVersion); version != "" {
			return fmt.Errorf("%w (%s)", ErrUpToDate, version)
		}
		return ErrUpToDate
	}
	// The installed name is kept. If the update changes the declared name, users are told how to use it
	previousName, _ := readDeclaredName(ext.Path())

	var err error
	if ext.IsBinary() {
		err = m.upgradeBinExtension(ext)
	} else {
		// Check if git extension has changed to a binary extension
		var isBin bool
		repo, repoErr := repoFromPath(ext.Path())
		if repoErr == nil {
			isBin, _ = isBinExtension(m.client, repo)
		}
		if isBin {
			// The git checkout is replaced only after the binary release is installed successfully
			err = m.installBin(repo, filepath.Base(ext.Path()), "")
		} else {
			err = m.upgradeGitExtension(ext, force)
		}
	}
	if err == nil && !m.dryRunMode {
		m.printNameChange(ext, previousName)
	}
	return err
}

// printNameChange prints a notice if the name declared in the extension's manifest changed in an update,
// and differs from the installed name. Updates don't rename extensions
func (m *Manager) printNameChange(ext Extension, previousName string) {
	currentName, err := readDeclaredName(ext.Path())
	if err != nil || currentName == "" {
		return
	}
	previous, _ := extensions.NormalizeName(previousName)
	current, err := extensions.NormalizeName(currentName)
	if err != nil || current == previous || current == ext.Name() {
		return
	}
	source := ext.URL()
	if repo, err := ghrepo.FromFullName(source); err == nil && repo.RepoHost() == "github.com" {
		source = ghrepo.FullName(repo)
	}
	fmt.Fprintf(m.io.ErrOut, "extension %s now declares the name '%s'. To use the new name, run:\n  c8y extensions delete %s\n  c8y extensions install %s\n",
		ext.Name(), current, ext.Name(), source)
}

func (m *Manager) upgradeGitExtension(ext Extension, force bool) error {
	exe, err := m.lookPath("git")
	if err != nil {
		return err
	}
	dir := ext.path
	if m.dryRunMode {
		return nil
	}
	if force {
		if err := m.newCommand(exe, "-C", dir, "fetch", "origin", "HEAD").Run(); err != nil {
			return err
		}
		return m.newCommand(exe, "-C", dir, "reset", "--hard", "origin/HEAD").Run()
	}
	return m.newCommand(exe, "-C", dir, "pull", "--ff-only").Run()
}

func (m *Manager) upgradeBinExtension(ext Extension) error {
	repo, err := ghrepo.FromFullName(ext.url)
	if err != nil {
		return fmt.Errorf("failed to parse URL %s: %w", ext.url, err)
	}
	return m.installBin(repo, filepath.Base(ext.Path()), "")
}

func (m *Manager) Remove(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("extension name is empty")
	}
	targetDirs := []string{
		filepath.Join(m.installDir(), ExtPrefix+name),
		filepath.Join(m.installDir(), name),
	}

	var targetDir string
	var found = false

	for _, targetDir = range targetDirs {
		if _, err := os.Lstat(targetDir); os.IsNotExist(err) {
			continue
		}
		found = true
		break
	}

	if !found {
		for _, ext := range m.listSystemExtensions(nil) {
			if ext.Name() == strings.TrimPrefix(name, ExtPrefix) {
				return fmt.Errorf("%q is installed in a system extensions folder (%s). Remove it with the package manager which installed it", name, ext.Path())
			}
		}
		return fmt.Errorf("no extension found: %q", name)
	}

	if m.dryRunMode {
		return nil
	}
	return os.RemoveAll(targetDir)
}

func (m *Manager) installDir() string {
	return m.dataDir()
}

//go:embed ext_tmpls/script.sh
var scriptTmpl string

//go:embed ext_tmpls/README.md
var readmeTmpl string

//go:embed ext_tmpls/extension.yaml
var exampleExtensionManifest []byte

//go:embed ext_tmpls/customCommand.jsonnet
var exampleJsonnet []byte

//go:embed ext_tmpls/exampleDevice.json
var exampleView []byte

//go:embed ext_tmpls/apiCommandTemplate.yaml
var commandGroupTmpl string

//go:embed all:ext_tmpls/go all:ext_tmpls/other
var precompiledTemplates embed.FS

// defaultExtSDKVersion is the extension Go SDK version used by the Go template if the
// build does not set extSDKVersion, e.g. go install or go build without -ldflags
const defaultExtSDKVersion = "v0.0.1"

// extSDKVersion is set at build time to the newest pkg/extsdk/* tag (without the prefix),
// see the C8Y_EXTSDK_VERSION variable in Taskfile.yml and the ldflags in .goreleaser.yml
var extSDKVersion string

// ExtSDKVersion returns the version of the extension Go SDK used by the Go template
func ExtSDKVersion() string {
	if semver.IsValid(extSDKVersion) {
		return extSDKVersion
	}
	return defaultExtSDKVersion
}

// templateData is the data available in the precompiled extension templates
type templateData struct {
	// Name of the extension, e.g. c8y-foo
	Name string
	// ExtName is the command name of the extension, e.g. foo
	ExtName string
	// ViewType is the type of the example items, which is used by the example view
	ViewType string
	// SDKVersion is the version of the extension Go SDK
	SDKVersion string
}

// renderTemplates renders the templates in the given folder of the embedded templates into dir.
// Templates use [[ ]] delimiters, as the generated files (e.g. GoReleaser and GitHub workflows)
// use {{ }} themselves. A ".tmpl" suffix is removed, and "gitignore" is renamed to ".gitignore"
func renderTemplates(root string, dir string, data templateData) ([]string, error) {
	files := []string{}
	err := fs.WalkDir(precompiledTemplates, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel := strings.TrimSuffix(strings.TrimPrefix(path, root+"/"), ".tmpl")
		if rel == "gitignore" {
			rel = ".gitignore"
		}
		contents, err := precompiledTemplates.ReadFile(path)
		if err != nil {
			return err
		}
		tmpl, err := template.New(rel).Delims("[[", "]]").Parse(string(contents))
		if err != nil {
			return fmt.Errorf("invalid template %s: %w", path, err)
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("invalid template %s: %w", path, err)
		}
		mode := os.FileMode(0644)
		if strings.HasSuffix(rel, ".sh") {
			mode = 0755
		}
		if err := writeFile(filepath.Join(dir, filepath.FromSlash(rel)), buf.Bytes(), mode); err != nil {
			return err
		}
		files = append(files, rel)
		return nil
	})
	return files, err
}

// createPrecompiled creates a precompiled (binary) extension from the templates
func (m *Manager) createPrecompiled(gitExe string, name string, tmplType extensions.ExtTemplateType) error {
	cmdName := strings.TrimPrefix(name, ExtPrefix)
	data := templateData{
		Name:       name,
		ExtName:    cmdName,
		ViewType:   strings.ReplaceAll(name, "-", "") + "_service",
		SDKVersion: ExtSDKVersion(),
	}

	root := "ext_tmpls/other"
	if tmplType == extensions.GoBinTemplateType {
		root = "ext_tmpls/go"
	}
	files, err := renderTemplates(root, name, data)
	if err != nil {
		return err
	}

	if tmplType == extensions.GoBinTemplateType {
		// Not fatal, e.g. when offline. The user can run the commands later
		if goExe, err := m.lookPath("go"); err != nil {
			fmt.Fprintf(m.io.ErrOut, "go was not found, so the dependencies were not downloaded. Run 'go mod tidy' in %s\n", name)
		} else {
			tidy := m.newCommand(goExe, "mod", "tidy")
			tidy.Dir = name
			tidy.Stderr = m.io.ErrOut
			if err := tidy.Run(); err != nil {
				fmt.Fprintf(m.io.ErrOut, "could not download the Go dependencies. Run 'go mod tidy' in %s. %s\n", name, err)
			} else {
				files = append(files, "go.sum")
				build := m.newCommand(goExe, "build", "-o", filepath.Join("bin", name), ".")
				build.Dir = name
				build.Stderr = m.io.ErrOut
				if err := build.Run(); err != nil {
					fmt.Fprintf(m.io.ErrOut, "could not build the binary. %s\n", err)
				}
			}
		}
	}

	addArgs := []string{"-C", name, "add", "--"}
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(name, filepath.FromSlash(f))); err == nil {
			addArgs = append(addArgs, filepath.FromSlash(f))
		}
	}
	if err := m.newCommand(gitExe, addArgs...).Run(); err != nil {
		return err
	}
	if tmplType == extensions.OtherBinTemplateType {
		return m.newCommand(gitExe, "-C", name, "update-index", "--chmod=+x", filepath.Join("script", "build.sh")).Run()
	}
	return nil
}

func (m *Manager) Create(name string, tmplType extensions.ExtTemplateType) error {
	exe, err := m.lookPath("git")
	if err != nil {
		return err
	}

	cmdName := strings.TrimPrefix(name, ExtPrefix)

	if err := m.newCommand(exe, "init", "--quiet", name).Run(); err != nil {
		return err
	}

	if tmplType == extensions.GoBinTemplateType || tmplType == extensions.OtherBinTemplateType {
		return m.createPrecompiled(exe, name, tmplType)
	}

	manifest := &bytes.Buffer{}
	manifestTmpl, err := template.New(fileAlias).Delims("[[", "]]").Parse(string(exampleExtensionManifest))
	if err != nil {
		return err
	}
	if err := manifestTmpl.Execute(manifest, templateData{Name: name, ExtName: cmdName}); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(name, "extension.yaml"), manifest.Bytes(), 0755); err != nil {
		return err
	}

	readme := fmt.Sprintf(readmeTmpl, name)
	if err := writeFile(filepath.Join(name, "README.md"), []byte(readme), 0644); err != nil {
		return err
	}

	commandsDir := filepath.Join(name, commandsName)
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		return err
	}
	subCommandsDir := filepath.Join(name, commandsName, "services")
	if err := os.MkdirAll(subCommandsDir, 0755); err != nil {
		return err
	}

	apiDir := filepath.Join(name, apiName)
	if err := os.MkdirAll(apiDir, 0755); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(apiDir, "devices.yaml"), []byte(commandGroupTmpl), 0644); err != nil {
		return err
	}

	templatesDir := filepath.Join(name, templateName)
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(templatesDir, "customCommand.jsonnet"), exampleJsonnet, 0644); err != nil {
		return err
	}

	viewsDir := filepath.Join(name, viewsName)
	if err := os.MkdirAll(viewsDir, 0755); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(viewsDir, "exampleDevice.json"), exampleView, 0644); err != nil {
		return err
	}

	script := fmt.Sprintf(scriptTmpl, cmdName, "services list")
	if err := writeFile(filepath.Join(subCommandsDir, "list"), []byte(script), 0755); err != nil {
		return err
	}
	if err := m.newCommand(exe, "-C", name, "add", filepath.Join(commandsName, "services", "list"), "--chmod=+x").Run(); err != nil {
		return err
	}

	// stage remaining files
	return m.newCommand(exe, "-C", name, "add", "**").Run()
}

func isSymlink(m os.FileMode) bool {
	return m&os.ModeSymlink != 0
}

func writeFile(p string, contents []byte, mode os.FileMode) error {
	if dir := filepath.Dir(p); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return os.WriteFile(p, contents, mode)
}

// reads the product of makeSymlink on Windows
func readPathFromFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	b := make([]byte, 1024)
	n, err := f.Read(b)
	return strings.TrimSpace(string(b[:n])), err
}

func isBinExtension(client *http.Client, repo ghrepo.Interface) (bool, error) {
	r, err := fetchLatestRelease(client, repo)
	if err != nil {
		return false, err
	}
	return hasPlatformAssets(r), nil
}

func repoFromPath(path string) (ghrepo.Interface, error) {
	remotes, err := git.RemotesForPath(path)
	if err != nil {
		return nil, err
	}

	if len(remotes) == 0 {
		return nil, fmt.Errorf("no remotes configured for %s", path)
	}

	var remote *git.Remote

	for _, r := range remotes {
		if r.Name == "origin" {
			remote = r
			break
		}
	}

	if remote == nil {
		remote = remotes[0]
	}

	return ghrepo.FromURL(remote.FetchURL)
}

func hasRosetta() bool {
	_, err := os.Stat("/Library/Apple/usr/libexec/oah/libRosettaRuntime")
	return err == nil
}

func codesignBinary(binPath string) error {
	codesignExe, err := safeexec.LookPath("codesign")
	if err != nil {
		return err
	}
	cmd := exec.Command(codesignExe, "--sign", "-", "--force", "--preserve-metadata=entitlements,requirements,flags,runtime", binPath)
	return cmd.Run()
}

type AliasCollection struct {
	Name    string
	Aliases []extensions.Alias
}
