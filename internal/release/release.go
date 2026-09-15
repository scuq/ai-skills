// Package release finds a release of a GitHub repository, downloads its
// assets, and verifies them against the SHA256SUMS asset.
package release

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SumsName is the name of the asset with the SHA-256 sums of the other assets.
const SumsName = "SHA256SUMS"

const userAgent = "ai-skillsctl"

// A Client reads releases through the GitHub REST API.
type Client struct {
	// APIURL is the base URL of the API, for example https://api.github.com.
	APIURL string
	// Repo is the repository as OWNER/NAME.
	Repo string
	// Token is a GitHub token for the API requests. If Token is empty, the
	// client sends no Authorization header. The client never sends Token
	// with an asset download.
	Token string
	// HTTP is the HTTP client. If HTTP is nil, the client uses a client
	// with a timeout of 5 minutes.
	HTTP *http.Client
}

// A Release is one release of the repository.
type Release struct {
	// Tag is the Git tag of the release, for example "v0.1.0".
	Tag string
	// Assets holds the download URL of each asset, by asset name.
	Assets map[string]string
}

// Find returns the release with the tag tag.
// If tag is empty, Find returns the latest release.
func (c *Client) Find(ctx context.Context, tag string) (Release, error) {
	path := "/releases/latest"
	if tag != "" {
		path = "/releases/tags/" + url.PathEscape(tag)
	}
	endpoint := strings.TrimSuffix(c.APIURL, "/") + "/repos/" + c.Repo + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", userAgent)
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return Release{}, err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound && tag == "":
		return Release{}, fmt.Errorf("repository %s has no release, or it is not public", c.Repo)
	case resp.StatusCode == http.StatusNotFound:
		return Release{}, fmt.Errorf("repository %s has no release with the tag %s", c.Repo, tag)
	case resp.StatusCode != http.StatusOK:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return Release{}, fmt.Errorf("GitHub API returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var data struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 10<<20)).Decode(&data); err != nil {
		return Release{}, fmt.Errorf("GitHub API response is not valid: %w", err)
	}
	if data.TagName == "" {
		return Release{}, fmt.Errorf("GitHub API response has no tag_name")
	}
	r := Release{Tag: data.TagName, Assets: make(map[string]string, len(data.Assets))}
	for _, a := range data.Assets {
		r.Assets[a.Name] = a.URL
	}
	return r, nil
}

// Download returns the content of the asset name of r.
// It returns an error if the content is larger than limit bytes.
func (c *Client) Download(ctx context.Context, r Release, name string, limit int64) ([]byte, error) {
	assetURL, ok := r.Assets[name]
	if !ok {
		return nil, fmt.Errorf("release %s has no asset %s", r.Tag, name)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, assetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download of %s returned %s", name, resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("download of %s: %w", name, err)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("asset %s is larger than %d bytes", name, limit)
	}
	return data, nil
}

// DownloadVerified downloads the asset [SumsName] and the asset name of r.
// It returns the content of name if its SHA-256 sum is the sum in [SumsName].
func (c *Client) DownloadVerified(ctx context.Context, r Release, name string, limit int64) ([]byte, error) {
	sumsData, err := c.Download(ctx, r, SumsName, 1<<20)
	if err != nil {
		return nil, err
	}
	sums, err := ParseSums(sumsData)
	if err != nil {
		return nil, err
	}
	data, err := c.Download(ctx, r, name, limit)
	if err != nil {
		return nil, err
	}
	if err := Verify(sums, name, data); err != nil {
		return nil, err
	}
	return data, nil
}

// ParseSums returns the sums in the output format of sha256sum, by file name.
func ParseSums(data []byte) (map[string]string, error) {
	sums := map[string]string{}
	for i, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		sum, name, ok := strings.Cut(line, " ")
		name = strings.TrimLeft(name, " *")
		if _, err := hex.DecodeString(sum); !ok || err != nil || len(sum) != 64 || name == "" {
			return nil, fmt.Errorf("%s line %d is not in the sha256sum format", SumsName, i+1)
		}
		sums[name] = strings.ToLower(sum)
	}
	return sums, nil
}

// Verify returns an error if the SHA-256 sum of data is not the sum for name
// in sums.
func Verify(sums map[string]string, name string, data []byte) error {
	want, ok := sums[name]
	if !ok {
		return fmt.Errorf("%s has no sum for %s", SumsName, name)
	}
	got := sha256.Sum256(data)
	if hex.EncodeToString(got[:]) != want {
		return fmt.Errorf("the SHA-256 sum of %s is not the sum in %s, the download is possibly damaged or changed", name, SumsName)
	}
	return nil
}

func (c *Client) httpClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 5 * time.Minute}
}
