package extension

import (
	"bytes"
	_ "embed"
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

	"github.com/reubenmiller/go-c8y-cli/v2/internal/ghrepo"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/findsh"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/git"

	"github.com/cli/safeexec"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"gopkg.in/yaml.v3"
)

type Manager struct {
	dataDir     func() string
	lookPath    func(string) (string, error)
	findSh      func() (string, error)
	newCommand  func(string, ...string) *exec.Cmd
	platform    func() (string, string)
	client      *http.Client
	config      config.Config
	io          *iostreams.IOStreams
	dryRunMode  bool
	allowBinary bool
}

func NewManager(ios *iostreams.IOStreams, cfg *config.Config) *Manager {
	return &Manager{
		dataDir:    cfg.ExtensionsDataDir,
		lookPath:   safeexec.LookPath,
		findSh:     findsh.Find,
		newCommand: exec.Command,
		dryRunMode: cfg.DryRun(),
		platform: func() (string, string) {
			ext := ""
			if runtime.GOOS == "windows" {
				ext = ".exe"
			}
			return fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH), ext
		},
		io: ios,
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

func (m *Manager) Dispatch(args []string, stdin io.Reader, stdout, stderr io.Writer) (bool, error) {
	if len(args) == 0 {
		return false, errors.New("too few arguments in list")
	}

	var exe string
	extName := args[0]

	forwardArgs := []string{}

	cArgs := strings.Join(args[1:], " ")

	exts, _ := m.list(false)
	var ext Extension
	found := false
	for _, e := range exts {
		if e.Name() == extName {

			if commands, cmdErr := e.Commands(); cmdErr == nil {
				for _, c := range commands {
					if strings.HasPrefix(cArgs, c.Name()) {
						ext = e

						if ext.IsBinary() {
							exe = filepath.Join(ext.Path())
						} else {
							exe = filepath.Join(ext.Path(), commandsName)
						}

						exe = filepath.Join(append([]string{exe}, strings.Split(c.Name(), " ")...)...)

						consumerArgCount := strings.Count(c.Name(), " ") + 1
						if consumerArgCount < len(args)-1 {
							forwardArgs = args[consumerArgCount+1:]
						}
						found = true
						break
					}
				}
			}
		}

		if found {
			break
		}
		// if e.Name() == extName {
		// 	ext = e
		// 	if ext.IsBinary() {
		// 		exe = filepath.Join(ext.Path())
		// 	} else {
		// 		exe = filepath.Join(ext.Path(), commandsName, subCommand)
		// 	}
		// 	break
		// }
	}
	if exe == "" {
		return false, nil
	}

	var externalCmd *exec.Cmd

	if ext.IsBinary() || runtime.GOOS != "windows" {
		externalCmd = m.newCommand(exe, forwardArgs...)
	} else if runtime.GOOS == "windows" {
		// Dispatch all extension calls through the `sh` interpreter to support executable files with a
		// shebang line on Windows.
		shExe, err := m.findSh()
		if err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				return true, errors.New("the `sh.exe` interpreter is required. Please install Git for Windows and try again")
			}
			return true, err
		}
		forwardArgs = append([]string{"-c", `command "$@"`, "--", exe}, forwardArgs...)
		externalCmd = m.newCommand(shExe, forwardArgs...)
	}
	externalCmd.Stdin = stdin
	externalCmd.Stdout = stdout
	externalCmd.Stderr = stderr
	return true, externalCmd.Run()
}

func (m *Manager) Execute(exe string, args []string, isBinary bool, stdin io.Reader, stdout, stderr io.Writer) (bool, error) {
	forwardArgs := args[:]

	if exe == "" {
		return false, nil
	}

	var externalCmd *exec.Cmd

	if isBinary || runtime.GOOS != "windows" {
		externalCmd = m.newCommand(exe, forwardArgs...)
	} else if runtime.GOOS == "windows" {
		// Dispatch all extension calls through the `sh` interpreter to support executable files with a
		// shebang line on Windows.
		shExe, err := m.findSh()
		if err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				return true, errors.New("the `sh.exe` interpreter is required. Please install Git for Windows and try again")
			}
			return true, err
		}
		forwardArgs = append([]string{"-c", `command "$@"`, "--", exe}, forwardArgs...)
		externalCmd = m.newCommand(shExe, forwardArgs...)
	}
	externalCmd.Stdin = stdin
	externalCmd.Stdout = stdout
	externalCmd.Stderr = stderr
	return true, externalCmd.Run()
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
	if err != nil {
		return nil, err
	}

	var results []Extension
	for _, f := range entries {
		if !strings.HasPrefix(f.Name(), ExtPrefix) && !strings.Contains(f.Name(), ExtPrefix) {
			continue
		}
		var ext Extension
		var err error
		if f.IsDir() {
			ext, err = m.parseExtensionDir(f)
			if err != nil {
				return nil, err
			}
			results = append(results, ext)
		} else {
			ext, err = m.parseExtensionFile(f)
			if err != nil {
				return nil, err
			}
			results = append(results, ext)
		}
	}

	if includeMetadata {
		m.populateLatestVersions(results)
	}

	return results, nil
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
	exePath := filepath.Join(id, fi.Name(), fi.Name())
	ext := Extension{path: exePath, kind: BinaryKind}
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

func (m *Manager) InstallLocal(dir string, name string) error {
	if name == "" {
		name = filepath.Base(dir)
	}
	targetLink := filepath.Join(m.installDir(), name)
	if err := os.MkdirAll(filepath.Dir(targetLink), 0755); err != nil {
		return err
	}
	return makeSymlink(dir, targetLink)
}

type binManifest struct {
	Owner    string
	Name     string
	Host     string
	Tag      string
	IsPinned bool
	// TODO I may end up not using this; just thinking ahead to local installs
	Path string
}

// Install installs an extension from repo, and pins to commitish if provided
func (m *Manager) Install(repo ghrepo.Interface, name string, target string) error {
	// FUTURE: Disable binary extensions for now
	if m.allowBinary && strings.Contains(repo.RepoHost(), "github") {
		isBin, err := isBinExtension(m.client, repo)
		if err != nil {
			if errors.Is(err, ErrReleaseNotFound) {
				if ok, err := repoExists(m.client, repo); err != nil {
					return err
				} else if !ok {
					return ErrRepositoryNotFound
				}
			} else {
				return fmt.Errorf("could not check for binary extension: %w", err)
			}
		}
		if isBin {
			return m.installBin(repo, target)
		}

		hb, err := hasBundle(m.client, repo)
		if err != nil {
			return err
		}

		if !hb {
			return errors.New("extension is not installable: missing executable")
		}
	}

	return m.installGit(repo, name, target, m.io.Out, m.io.ErrOut)
}

func (m *Manager) installBin(repo ghrepo.Interface, target string) error {
	var r *release
	var err error
	isPinned := target != ""
	if isPinned {
		r, err = fetchReleaseFromTag(m.client, repo, target)
	} else {
		r, err = fetchLatestRelease(m.client, repo)
	}
	if err != nil {
		return err
	}

	platform, ext := m.platform()
	isMacARM := platform == "darwin-arm64"
	trueARMBinary := false

	var asset *releaseAsset
	for _, a := range r.Assets {
		if strings.HasSuffix(a.Name, platform+ext) {
			asset = &a
			trueARMBinary = isMacARM
			break
		}
	}

	// if an arm64 binary is unavailable, fall back to amd64 if it can be executed through Rosetta 2
	if asset == nil && isMacARM && hasRosetta() {
		for _, a := range r.Assets {
			if strings.HasSuffix(a.Name, "darwin-amd64") {
				asset = &a
				break
			}
		}
	}

	if asset == nil {
		return fmt.Errorf(
			"%[1]s unsupported for %[2]s. Open an issue: `gh issue create -R %[3]s/%[1]s -t'Support %[2]s'`",
			repo.RepoName(), platform, repo.RepoOwner())
	}

	name := repo.RepoName()
	targetDir := filepath.Join(m.installDir(), name)

	// TODO clean this up if function errs?
	if !m.dryRunMode {
		err = os.MkdirAll(targetDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create installation directory: %w", err)
		}
	}

	binPath := filepath.Join(targetDir, name)
	binPath += ext

	if !m.dryRunMode {
		err = downloadAsset(m.client, *asset, binPath)
		if err != nil {
			return fmt.Errorf("failed to download asset %s: %w", asset.Name, err)
		}
		if trueARMBinary {
			if err := codesignBinary(binPath); err != nil {
				return fmt.Errorf("failed to codesign downloaded binary: %w", err)
			}
		}
	}

	manifest := binManifest{
		Name:     name,
		Owner:    repo.RepoOwner(),
		Host:     repo.RepoHost(),
		Path:     binPath,
		Tag:      r.Tag,
		IsPinned: isPinned,
	}

	bs, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to serialize manifest: %w", err)
	}

	if !m.dryRunMode {
		manifestPath := filepath.Join(targetDir, manifestName)

		f, err := os.OpenFile(manifestPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return fmt.Errorf("failed to open manifest for writing: %w", err)
		}
		defer f.Close()

		_, err = f.Write(bs)
		if err != nil {
			return fmt.Errorf("failed write manifest file: %w", err)
		}
	}

	return nil
}

func (m *Manager) installGit(repo ghrepo.Interface, name, target string, stdout, stderr io.Writer) error {
	protocol := repo.RepoHost()
	cloneURL := ghrepo.FormatRemoteURL(repo, protocol)

	exe, err := m.lookPath("git")
	if err != nil {
		return err
	}

	var commitSHA string
	if target != "" {
		commitSHA, err = fetchCommitSHA(m.client, repo, target)
		if err != nil {
			return err
		}
	}

	if name == "" {
		name = strings.TrimSuffix(path.Base(cloneURL), ".git")
	}
	targetDir := filepath.Join(m.installDir(), name)

	externalCmd := m.newCommand(exe, "clone", cloneURL, targetDir)
	externalCmd.Stdout = stdout
	externalCmd.Stderr = stderr
	if err := externalCmd.Run(); err != nil {
		return err
	}
	if commitSHA == "" {
		return nil
	}

	checkoutCmd := m.newCommand(exe, "-C", targetDir, "checkout", commitSHA)
	checkoutCmd.Stdout = stdout
	checkoutCmd.Stderr = stderr
	if err := checkoutCmd.Run(); err != nil {
		return err
	}

	pinPath := filepath.Join(targetDir, fmt.Sprintf(".pin-%s", commitSHA))
	f, err := os.OpenFile(pinPath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to create pin file in directory: %w", err)
	}
	return f.Close()
}

var ErrPinnedExtensionUpgrade = errors.New("pinned extensions can not be updated")
var ErrLocalExtensionUpgrade = errors.New("local extensions can not be updated")
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
	if ext.isLocal {
		return ErrLocalExtensionUpgrade
	}
	if ext.IsPinned() {
		return ErrPinnedExtensionUpgrade
	}
	if !ext.UpdateAvailable() {
		return ErrUpToDate
	}
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
			if err := m.Remove(ext.Name()); err != nil {
				return fmt.Errorf("failed to migrate to new precompiled extension format: %w", err)
			}
			return m.installBin(repo, "")
		}
		err = m.upgradeGitExtension(ext, force)
	}
	return err
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
	return m.installBin(repo, "")
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

func (m *Manager) Create(name string, tmplType extensions.ExtTemplateType) error {
	exe, err := m.lookPath("git")
	if err != nil {
		return err
	}

	cmdName := strings.TrimPrefix(name, ExtPrefix)

	if err := m.newCommand(exe, "init", "--quiet", name).Run(); err != nil {
		return err
	}

	if err := writeFile(filepath.Join(name, "extension.yaml"), exampleExtensionManifest, 0755); err != nil {
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

	if tmplType == extensions.GoBinTemplateType {
		return nil
	} else if tmplType == extensions.OtherBinTemplateType {
		return nil
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

func isBinExtension(client *http.Client, repo ghrepo.Interface) (isBin bool, err error) {
	var r *release
	r, err = fetchLatestRelease(client, repo)
	if err != nil {
		return
	}

	for _, a := range r.Assets {
		dists := possibleDists()
		for _, d := range dists {
			suffix := d
			if strings.HasPrefix(d, "windows") {
				suffix += ".exe"
			}
			if strings.HasSuffix(a.Name, suffix) {
				isBin = true
				break
			}
		}
	}

	return
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

func possibleDists() []string {
	return []string{
		"aix-ppc64",
		"android-386",
		"android-amd64",
		"android-arm",
		"android-arm64",
		"darwin-amd64",
		"darwin-arm64",
		"dragonfly-amd64",
		"freebsd-386",
		"freebsd-amd64",
		"freebsd-arm",
		"freebsd-arm64",
		"illumos-amd64",
		"ios-amd64",
		"ios-arm64",
		"js-wasm",
		"linux-386",
		"linux-amd64",
		"linux-arm",
		"linux-arm64",
		"linux-mips",
		"linux-mips64",
		"linux-mips64le",
		"linux-mipsle",
		"linux-ppc64",
		"linux-ppc64le",
		"linux-riscv64",
		"linux-s390x",
		"netbsd-386",
		"netbsd-amd64",
		"netbsd-arm",
		"netbsd-arm64",
		"openbsd-386",
		"openbsd-amd64",
		"openbsd-arm",
		"openbsd-arm64",
		"openbsd-mips64",
		"plan9-386",
		"plan9-amd64",
		"plan9-arm",
		"solaris-amd64",
		"windows-386",
		"windows-amd64",
		"windows-arm",
		"windows-arm64",
	}
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
