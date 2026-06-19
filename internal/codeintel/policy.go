package codeintel

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strings"
)

var defaultGitHosts = []string{"github.com", "gitlab.com"}

type RepositoryPolicy struct {
	AllowLocalPaths bool
	AllowedGitHosts []string
}

func NewRepositoryPolicy(allowLocalPaths bool, allowedGitHosts []string) RepositoryPolicy {
	hosts := normalizeHosts(allowedGitHosts)
	if len(hosts) == 0 {
		hosts = defaultGitHosts
	}
	return RepositoryPolicy{AllowLocalPaths: allowLocalPaths, AllowedGitHosts: hosts}
}

func (p RepositoryPolicy) ValidateRegistration(remoteURL, localPath string) error {
	remoteURL = strings.TrimSpace(remoteURL)
	localPath = strings.TrimSpace(localPath)
	if remoteURL == "" && localPath == "" {
		return errors.New("local_path or remote_url is required")
	}
	if localPath != "" && !p.AllowLocalPaths {
		return errors.New("local_path repository registration is disabled")
	}
	if remoteURL != "" {
		if err := p.ValidateRemoteURL(remoteURL); err != nil {
			return err
		}
	}
	return nil
}

func (p RepositoryPolicy) ValidateRemoteURL(rawURL string) error {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return fmt.Errorf("invalid remote_url: %w", err)
	}
	if parsed.Scheme != "https" {
		return errors.New("remote_url must use https")
	}
	if parsed.User != nil {
		return errors.New("remote_url must not contain userinfo")
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return errors.New("remote_url host is required")
	}
	if isBlockedHost(host) {
		return fmt.Errorf("remote_url host %q is not allowed", host)
	}
	if !hostAllowed(host, p.allowedHosts()) {
		return fmt.Errorf("remote_url host %q is not in the allowed Git host list", host)
	}
	return nil
}

func (p RepositoryPolicy) allowedHosts() []string {
	hosts := normalizeHosts(p.AllowedGitHosts)
	if len(hosts) == 0 {
		return defaultGitHosts
	}
	return hosts
}

func normalizeHosts(values []string) []string {
	seen := map[string]struct{}{}
	var hosts []string
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if strings.Contains(value, "://") {
			if parsed, err := url.Parse(value); err == nil {
				value = parsed.Hostname()
			}
		}
		host, _, err := net.SplitHostPort(value)
		if err == nil {
			value = host
		}
		value = strings.Trim(value, "[]")
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		hosts = append(hosts, value)
	}
	return hosts
}

func hostAllowed(host string, allowed []string) bool {
	for _, candidate := range allowed {
		if host == candidate {
			return true
		}
	}
	return false
}

func isBlockedHost(host string) bool {
	switch host {
	case "localhost", "metadata.google.internal", "metadata", "169.254.169.254":
		return true
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	return addr.IsLoopback() ||
		addr.IsPrivate() ||
		addr.IsLinkLocalUnicast() ||
		addr.IsLinkLocalMulticast() ||
		addr.IsMulticast() ||
		addr.IsUnspecified()
}
