package ghrepo

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/ghinstance"
)

// Interface describes an object that represents a GitHub repository
type Interface interface {
	RepoName() string
	RepoOwner() string
	RepoHost() string
	RepoURL() string
}

func NewRepoFromHost(u string, defaultHost string) (*Respository, error) {
	repo := Respository{
		host: defaultHost,
	}

	parts := strings.Split(u, "/")
	repoURL, err := url.Parse(u)

	if err == nil {
		if repoURL.Host != "" {
			repo.host = repoURL.Host
		}
		if repoURL.Scheme != "" {
			repo.rawURL = u
		}
		parts = strings.Split(repoURL.Path, "/")
	}

	if len(parts) >= 2 {
		repo.owner = strings.Join(parts[0:len(parts)-1], "/")
		repo.name = parts[len(parts)-1]
	}
	return &repo, nil
}

type Respository struct {
	name   string
	owner  string
	host   string
	rawURL string
}

func (r *Respository) Name() string {
	return r.name
}
func (r *Respository) Owner() string {
	return r.owner
}
func (r *Respository) Host() string {
	return r.host
}

func (r *Respository) URL() string {
	return r.rawURL
}

// New instantiates a GitHub repository from owner and name arguments
func New(owner, repo string) Interface {
	return NewWithHost(owner, repo, ghinstance.Default(), "")
}

// NewWithHost is like New with an explicit host name
func NewWithHost(owner, repo, hostname string, fullURL string) Interface {
	return &ghRepo{
		owner:    owner,
		name:     repo,
		hostname: normalizeHostname(hostname),
		fullURL:  fullURL,
	}
}

// FullName serializes a GitHub repository into an "OWNER/REPO" string
func FullName(r Interface) string {
	return fmt.Sprintf("%s/%s", r.RepoOwner(), r.RepoName())
}

func defaultHost() string {
	return "github.com"
}

// FromFullName extracts the GitHub repository information from the following
// formats: "OWNER/REPO", "HOST/OWNER/REPO", and a full URL.
func FromFullName(nwo string) (Interface, error) {
	return FromFullNameWithHost(nwo, defaultHost())
}

// FromFullNameWithHost is like FromFullName that defaults to a specific host for values that don't
// explicitly include a hostname.
func FromFullNameWithHost(nwo, fallbackHost string) (Interface, error) {
	// repo, err := repository.ParseWithHost(nwo, fallbackHost)
	repo, err := NewRepoFromHost(nwo, fallbackHost)
	if err != nil {
		return nil, err
	}
	return NewWithHost(repo.Owner(), repo.Name(), repo.Host(), repo.URL()), nil
}

// FromURL extracts the GitHub repository information from a git remote URL
func FromURL(u *url.URL) (Interface, error) {
	if u.Hostname() == "" {
		return nil, fmt.Errorf("no hostname detected")
	}

	parts := strings.SplitN(strings.Trim(u.Path, "/"), "/", 3)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid path: %s", u.Path)
	}

	return NewWithHost(parts[0], strings.TrimSuffix(parts[1], ".git"), u.Hostname(), u.String()), nil
}

func normalizeHostname(h string) string {
	return strings.ToLower(strings.TrimPrefix(h, "www."))
}

// IsSame compares two GitHub repositories
func IsSame(a, b Interface) bool {
	return strings.EqualFold(a.RepoOwner(), b.RepoOwner()) &&
		strings.EqualFold(a.RepoName(), b.RepoName()) &&
		normalizeHostname(a.RepoHost()) == normalizeHostname(b.RepoHost())
}

func GenerateRepoURL(repo Interface, p string, args ...interface{}) string {
	baseURL := fmt.Sprintf("%s%s/%s", ghinstance.HostPrefix(repo.RepoHost()), repo.RepoOwner(), repo.RepoName())
	if p != "" {
		if path := fmt.Sprintf(p, args...); path != "" {
			return baseURL + "/" + path
		}
	}
	return baseURL
}

// TODO there is a parallel implementation for non-isolated commands
func FormatRemoteURL(repo Interface, protocol string) string {
	if repo.RepoURL() != "" {
		return repo.RepoURL()
	}
	if protocol == "ssh" {
		return fmt.Sprintf("git@%s:%s/%s.git", repo.RepoHost(), repo.RepoOwner(), repo.RepoName())
	}

	return fmt.Sprintf("%s%s/%s.git", ghinstance.HostPrefix(repo.RepoHost()), repo.RepoOwner(), repo.RepoName())
}

type ghRepo struct {
	owner    string
	name     string
	hostname string
	fullURL  string
}

func (r ghRepo) RepoOwner() string {
	return r.owner
}

func (r ghRepo) RepoName() string {
	return r.name
}

func (r ghRepo) RepoHost() string {
	return r.hostname
}

func (r ghRepo) RepoURL() string {
	return r.fullURL
}
