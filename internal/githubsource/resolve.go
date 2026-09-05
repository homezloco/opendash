package githubsource

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/opendash-project/opendash/internal/catalog"
	"github.com/opendash-project/opendash/internal/models"
)

const (
	defaultUserAgent = "OpenDash/0.1.0 (manifest-source; +https://opendash-project.github.io)"
	defaultMaxBytes  = 2 * 1024 * 1024 // 2 MiB
	defaultTimeout   = 30 * time.Second
	defaultRawBase   = "https://raw.githubusercontent.com"
	defaultAPIBase   = "https://api.github.com"
	githubTokenEnv   = "OPENDASH_GITHUB_TOKEN"
	redirectMaxDepth = 3
)

// Resolver fetches GitHub repository manifests safely. It can be pointed at a
// mock server in tests by overriding APIBase and RawBase.
type Resolver struct {
	Client    *http.Client
	Token     string
	UserAgent string
	APIBase   string
	RawBase   string
	MaxBytes  int64
	userAgent string
}

// NewResolver returns a Resolver with safe defaults. The optional token is read
// from the OPENDASH_GITHUB_TOKEN environment variable only and is never
// exposed elsewhere.
func NewResolver() *Resolver {
	return &Resolver{
		Client:    safeHTTPClient(),
		Token:     os.Getenv(githubTokenEnv),
		UserAgent: defaultUserAgent,
		APIBase:   defaultAPIBase,
		RawBase:   defaultRawBase,
		MaxBytes:  defaultMaxBytes,
	}
}

func safeHTTPClient() *http.Client {
	return &http.Client{
		Timeout: defaultTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= redirectMaxDepth {
				return errors.New("too many redirects")
			}
			if req.URL == nil {
				return errors.New("redirect has no URL")
			}
			if err := rejectUnsafeHost(req.URL.Host); err != nil {
				return fmt.Errorf("redirect rejected: %w", err)
			}
			return nil
		},
	}
}

// resolveRef converts a branch or tag name to a commit SHA using the GitHub API.
func (r *Resolver) ResolveRef(ctx context.Context, owner, repo, ref string) (string, error) {
	return r.resolveWithPath(ctx, owner, repo, "commits", ref)
}

func (r *Resolver) resolveWithPath(ctx context.Context, owner, repo, kind, ref string) (string, error) {
	u, err := url.Parse(r.APIBase)
	if err != nil {
		return "", fmt.Errorf("invalid API base URL: %w", err)
	}
	u.Path = path.Join(u.Path, "repos", owner, repo, kind, ref)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	r.setHeaders(req)

	resp, err := r.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("GitHub API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := r.readBody(resp)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("ref %q not found in %s/%s", ref, owner, repo)
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
	}
	var commit struct {
		SHA string `json:"sha"`
	}
	if err := json.Unmarshal(body, &commit); err != nil {
		return "", fmt.Errorf("parse commit response: %w", err)
	}
	if commit.SHA == "" {
		return "", fmt.Errorf("GitHub API did not return a commit SHA")
	}
	return commit.SHA, nil
}

// FetchManifest retrieves the manifest bytes for a resolved commit and path.
func (r *Resolver) FetchManifest(ctx context.Context, owner, repo, commit, manifestPath string) ([]byte, error) {
	u, err := url.Parse(r.RawBase)
	if err != nil {
		return nil, fmt.Errorf("invalid raw base URL: %w", err)
	}
	u.Path = path.Join(u.Path, owner, repo, commit, manifestPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	r.setHeaders(req)

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("raw content request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := r.readBody(resp)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("manifest not found at %s", u.Path)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("raw content returned %d", resp.StatusCode)
	}
	return body, nil
}

// Resolve fetches and validates the manifest for the supplied URL. It returns
// the resolved commit, the manifest, and a SHA-256 checksum of the raw bytes.
func (r *Resolver) Resolve(ctx context.Context, u *URL) (commit, checksum string, manifest *models.Manifest, err error) {
	commit, err = r.ResolveRef(ctx, u.Owner, u.Repo, u.Ref)
	if err != nil {
		return "", "", nil, err
	}
	data, err := r.FetchManifest(ctx, u.Owner, u.Repo, commit, u.Path)
	if err != nil {
		return "", "", nil, err
	}
	manifest, checksum, err = ParseManifest(data)
	if err != nil {
		return "", "", nil, err
	}
	return commit, checksum, manifest, nil
}

// ParseManifest validates the manifest bytes and returns a checksum.
func ParseManifest(data []byte) (*models.Manifest, string, error) {
	checksum := sha256.Sum256(data)
	var m models.Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, "", fmt.Errorf("parse manifest: %w", err)
	}
	if err := catalog.ValidateManifest(&m); err != nil {
		return nil, "", fmt.Errorf("invalid manifest: %w", err)
	}
	return &m, hex.EncodeToString(checksum[:]), nil
}

func (r *Resolver) setHeaders(req *http.Request) {
	ua := r.UserAgent
	if ua == "" {
		ua = defaultUserAgent
	}
	req.Header.Set("User-Agent", ua)
	if r.Token != "" {
		req.Header.Set("Authorization", "Bearer "+r.Token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
}

func (r *Resolver) readBody(resp *http.Response) ([]byte, error) {
	if resp.ContentLength > r.MaxBytes {
		return nil, errors.New("response exceeds size limit")
	}
	if resp.ContentLength < 0 && r.MaxBytes > 0 {
		// Unknown length; the limit reader will enforce the bound.
	}
	reader := io.LimitReader(resp.Body, r.MaxBytes+1)
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if int64(len(data)) > r.MaxBytes {
		return nil, errors.New("response exceeds size limit")
	}
	return data, nil
}

func rejectUnsafeHost(host string) error {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return errors.New("empty host")
	}
	// Strip port for host checks.
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	if host == "localhost" || host == "localhost." {
		return errors.New("localhost host is not allowed")
	}
	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return errors.New("private or loopback IP is not allowed")
		}
		return nil
	}
	// Reject known internal-looking domains. raw.githubusercontent.com and
	// api.github.com are the only allowed hosts for redirects.
	if host == "github.com" || host == "api.github.com" || host == "raw.githubusercontent.com" {
		return nil
	}
	// If it is not one of the known GitHub hosts, resolve and block private.
	addrs, err := net.LookupIP(host)
	if err != nil {
		return nil // cannot resolve; let the HTTP layer fail later
	}
	for _, a := range addrs {
		if a.IsLoopback() || a.IsPrivate() || a.IsLinkLocalUnicast() || a.IsLinkLocalMulticast() {
			return errors.New("resolved to a private or loopback address")
		}
	}
	return nil
}
