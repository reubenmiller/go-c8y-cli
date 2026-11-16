package extension

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/ghrepo"
)

// localhost is the domain name of a local GitHub instance
const localhost = "github.localhost"

func HandleHTTPError(resp *http.Response) error {
	return fmt.Errorf("http error. url=%s, status=%s", resp.Request.URL, resp.Status)
}

// githubToken returns the token to use for the given GitHub host, from the same
// environment variables as the GitHub CLI
func githubToken(host string) string {
	envs := []string{"GH_TOKEN", "GITHUB_TOKEN"}
	if isEnterprise(host) {
		envs = []string{"GH_ENTERPRISE_TOKEN", "GITHUB_ENTERPRISE_TOKEN"}
	}
	for _, name := range envs {
		if v := os.Getenv(name); v != "" {
			return v
		}
	}
	return ""
}

// newRequest creates a GitHub API request. A token is added if one is set for the host
func newRequest(host string, method string, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	if token := githubToken(host); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	return req, nil
}

func repoExists(httpClient *http.Client, repo ghrepo.Interface) (bool, error) {
	url := fmt.Sprintf("%srepos/%s/%s", RESTPrefix(repo.RepoHost()), repo.RepoOwner(), repo.RepoName())
	req, err := newRequest(repo.RepoHost(), "GET", url)
	if err != nil {
		return false, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return true, nil
	case 404:
		return false, nil
	default:
		return false, HandleHTTPError(resp)
	}
}

func hasBundle(httpClient *http.Client, repo ghrepo.Interface) (bool, error) {
	path := fmt.Sprintf("repos/%s/%s/contents",
		repo.RepoOwner(), repo.RepoName())
	url := RESTPrefix(repo.RepoHost()) + path
	req, err := newRequest(repo.RepoHost(), "GET", url)
	if err != nil {
		return false, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return false, nil
	}

	if resp.StatusCode > 299 {
		err = HandleHTTPError(resp)
		return false, err
	}

	return true, nil
}

type releaseAsset struct {
	Name   string
	APIURL string `json:"url"`
}

type release struct {
	Tag    string `json:"tag_name"`
	Assets []releaseAsset
}

// downloadAsset downloads a single asset to the given file path.
func downloadAsset(httpClient *http.Client, host string, asset releaseAsset, destPath string) error {
	req, err := newRequest(host, "GET", asset.APIURL)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/octet-stream")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return HandleHTTPError(resp)
	}

	f, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// downloadChecksums downloads and parses a checksums asset
func downloadChecksums(httpClient *http.Client, host string, asset releaseAsset) (map[string]string, error) {
	req, err := newRequest(host, "GET", asset.APIURL)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return nil, HandleHTTPError(resp)
	}
	// Checksum files are small, limit the size to protect against unexpected content
	return parseChecksums(io.LimitReader(resp.Body, 1<<20))
}

var ErrCommitNotFound = errors.New("commit not found")
var ErrReleaseNotFound = errors.New("release not found")
var ErrRepositoryNotFound = errors.New("repository not found")

// fetchLatestRelease finds the latest published release for a repository.
func fetchLatestRelease(httpClient *http.Client, baseRepo ghrepo.Interface) (*release, error) {
	path := fmt.Sprintf("repos/%s/%s/releases/latest", baseRepo.RepoOwner(), baseRepo.RepoName())
	url := RESTPrefix(baseRepo.RepoHost()) + path
	req, err := newRequest(baseRepo.RepoHost(), "GET", url)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, ErrReleaseNotFound
	}
	if resp.StatusCode > 299 {
		return nil, HandleHTTPError(resp)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r release
	err = json.Unmarshal(b, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// fetchReleaseFromTag finds release by tag name for a repository
func fetchReleaseFromTag(httpClient *http.Client, baseRepo ghrepo.Interface, tagName string) (*release, error) {
	fullRepoName := fmt.Sprintf("%s/%s", baseRepo.RepoOwner(), baseRepo.RepoName())
	path := fmt.Sprintf("repos/%s/releases/tags/%s", fullRepoName, tagName)
	url := RESTPrefix(baseRepo.RepoHost()) + path
	req, err := newRequest(baseRepo.RepoHost(), "GET", url)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return nil, ErrReleaseNotFound
	}
	if resp.StatusCode > 299 {
		return nil, HandleHTTPError(resp)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r release
	err = json.Unmarshal(b, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// isEnterprise returns true if the host is a GitHub Enterprise Server instance
func isEnterprise(hostname string) bool {
	return !strings.EqualFold(hostname, "github.com") && !strings.EqualFold(hostname, localhost)
}

func RESTPrefix(hostname string) string {
	if isEnterprise(hostname) {
		return fmt.Sprintf("https://%s/api/v3/", hostname)
	}
	if strings.EqualFold(hostname, localhost) {
		return fmt.Sprintf("http://api.%s/", hostname)
	}
	return fmt.Sprintf("https://api.%s/", hostname)
}

// fetchCommitSHA finds full commit SHA from a target ref in a repo
func fetchCommitSHA(httpClient *http.Client, baseRepo ghrepo.Interface, targetRef string) (string, error) {
	path := fmt.Sprintf("repos/%s/%s/commits/%s", baseRepo.RepoOwner(), baseRepo.RepoName(), targetRef)
	url := RESTPrefix(baseRepo.RepoHost()) + path
	req, err := newRequest(baseRepo.RepoHost(), "GET", url)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.github.v3.sha")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode == 422 {
		return "", ErrCommitNotFound
	}
	if resp.StatusCode > 299 {
		return "", HandleHTTPError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
