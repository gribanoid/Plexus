package safehttp

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ValidateWebhookURL rejects non-HTTPS URLs and destinations that resolve to private/link-local addresses.
func ValidateWebhookURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL")
	}
	if u.Scheme != "https" {
		return fmt.Errorf("webhook URL must use https")
	}
	if u.Host == "" || u.User != nil {
		return fmt.Errorf("invalid webhook host")
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("invalid webhook host")
	}
	if strings.EqualFold(host, "localhost") || strings.HasSuffix(strings.ToLower(host), ".localhost") {
		return fmt.Errorf("webhook URL must not target localhost")
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("webhook host lookup failed")
	}
	if len(ips) == 0 {
		return fmt.Errorf("webhook host has no addresses")
	}
	for _, ip := range ips {
		if isBlockedIP(ip) {
			return fmt.Errorf("webhook URL must not target private or link-local addresses")
		}
	}
	return nil
}

func isBlockedIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return true
	}
	// Extra cloud metadata ranges
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 169 && ip4[1] == 254 {
			return true
		}
	}
	return false
}

// NewWebhookClient returns an HTTP client with timeouts suitable for outbound webhooks.
func NewWebhookClient() *http.Client {
	return &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("too many redirects")
			}
			return ValidateWebhookURL(req.URL.String())
		},
	}
}
