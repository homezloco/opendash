package githubsource

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
)

var (
	segmentPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
)

// URL is a normalized GitHub repository/blob or tree URL that points to a
// manifest path. It deliberately does not include credentials, fragments, or
// other qualifiers.
type URL struct {
	Raw    string
	Owner  string
	Repo   string
	Ref    string
	Path   string
	IsTree bool
}

// ParseURL parses a GitHub repository URL that points to a manifest.
//
// Supported forms:
//
//	https://github.com/owner/repo/blob/ref/path/to/manifest.json
//	https://github.com/owner/repo/tree/ref/path/to/manifest.json
//
// It rejects credentials, fragments, non-https schemes, non-GitHub hosts,
// ambiguous paths, paths that do not end in .json, and paths that include
// traversal or hidden segments.
func ParseURL(raw string) (*URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "https" {
		return nil, fmt.Errorf("only https URLs are supported")
	}
	host := strings.ToLower(u.Host)
	if host != "github.com" && !strings.HasPrefix(host, "github.com:") {
		return nil, fmt.Errorf("only github.com URLs are supported")
	}
	if u.User != nil {
		return nil, fmt.Errorf("URLs containing credentials are not allowed")
	}
	if u.Fragment != "" {
		return nil, fmt.Errorf("URLs containing fragments are not allowed")
	}
	if u.RawQuery != "" {
		return nil, fmt.Errorf("URLs containing query strings are not allowed")
	}

	// Reject traversal and hidden segments in the raw path before normalization.
	rawPath := strings.TrimPrefix(u.EscapedPath(), "/")
	for _, seg := range strings.Split(rawPath, "/") {
		decoded, err := url.PathUnescape(seg)
		if err != nil {
			return nil, fmt.Errorf("manifest path is not valid: %w", err)
		}
		if decoded == "" {
			continue
		}
		if decoded == "." || decoded == ".." || strings.HasPrefix(decoded, ".") {
			return nil, fmt.Errorf("URL path contains an invalid segment %q", seg)
		}
	}

	cleaned := path.Clean(u.EscapedPath())
	if cleaned == "." {
		return nil, fmt.Errorf("URL path is empty")
	}
	parts := strings.Split(strings.TrimPrefix(cleaned, "/"), "/")
	if len(parts) < 5 {
		return nil, fmt.Errorf("URL does not point to a repository blob or tree path")
	}
	if parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return nil, fmt.Errorf("URL is missing owner, repo, or view type")
	}
	if parts[2] != "blob" && parts[2] != "tree" {
		return nil, fmt.Errorf("URL must be a /blob or /tree path")
	}

	owner, repo, view := parts[0], parts[1], parts[2]
	if !segmentPattern.MatchString(owner) || !segmentPattern.MatchString(repo) {
		return nil, fmt.Errorf("invalid owner or repository name")
	}
	if strings.Contains(owner, "..") || strings.Contains(repo, "..") {
		return nil, fmt.Errorf("invalid owner or repository name")
	}

	if len(parts) < 5 || parts[3] == "" {
		return nil, fmt.Errorf("URL is missing a ref and manifest path")
	}
	ref := parts[3]

	// Validate and decode each path segment before joining.
	var decodedParts []string
	for _, seg := range parts[4:] {
		if seg == "." || seg == ".." || strings.HasPrefix(seg, ".") {
			return nil, fmt.Errorf("manifest path contains an invalid segment %q", seg)
		}
		decoded, err := url.PathUnescape(seg)
		if err != nil {
			return nil, fmt.Errorf("manifest path is not valid: %w", err)
		}
		if decoded == "." || decoded == ".." || strings.HasPrefix(decoded, ".") ||
			strings.Contains(decoded, "/") || strings.Contains(decoded, "..") {
			return nil, fmt.Errorf("manifest path contains an invalid segment %q", seg)
		}
		decodedParts = append(decodedParts, decoded)
	}
	manifestPath := path.Join(decodedParts...)
	if manifestPath == "" || manifestPath == "." {
		return nil, fmt.Errorf("URL does not point to a manifest file")
	}
	if !strings.HasSuffix(manifestPath, ".json") {
		return nil, fmt.Errorf("manifest path must end with .json")
	}
	// Reject any remaining traversal or double slash in the cleaned path.
	if strings.Contains(manifestPath, "..") || strings.Contains(manifestPath, "//") {
		return nil, fmt.Errorf("manifest path is ambiguous")
	}

	return &URL{
		Raw:    raw,
		Owner:  owner,
		Repo:   repo,
		Ref:    ref,
		Path:   "/" + manifestPath,
		IsTree: view == "tree",
	}, nil
}
