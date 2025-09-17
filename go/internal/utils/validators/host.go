/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package validators

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// Public errors you can match on.
var (
	ErrScheme      = errors.New("scheme must be http or https")
	ErrUserInfo    = errors.New("userinfo not allowed (no user:pass@...)")
	ErrPathQueryFr = errors.New("must not include path, query, or fragment")
	ErrHostEmpty   = errors.New("empty host")
	ErrHostInvalid = errors.New("host must be a valid IP or hostname")
	ErrPortInvalid = errors.New("invalid port")
)

// ValidateHTTPHostURL checks that raw is a URL whose scheme is http/https, with a host
// that is either an IP or a DNS hostname, and optional port. No path/query/fragment.
func ValidateHTTPHostURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrScheme
	}
	if u.User != nil {
		return ErrUserInfo
	}
	if u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return ErrPathQueryFr
	}

	host := u.Hostname() // strips port and [] around IPv6
	if host == "" {
		return ErrHostEmpty
	}
	if !isValidIP(host) && !isValidHostname(host) {
		return ErrHostInvalid
	}

	if p := u.Port(); p != "" {
		if err := validatePort(p); err != nil {
			return err
		}
	}

	return nil
}

// IsValidHTTPHostURL returns true if ValidateHTTPHostURL(raw) == nil.
func IsValidHTTPHostURL(raw string) bool { return ValidateHTTPHostURL(raw) == nil }

func isValidIP(s string) bool {
	return net.ParseIP(s) != nil
}

func validatePort(p string) error {
	n, err := strconv.Atoi(p)
	if err != nil || n < 1 || n > 65535 {
		return ErrPortInvalid
	}
	return nil
}

// isValidHostname validates a DNS hostname per common rules:
// - total length 1..253
// - at least one dot (e.g., "foo.example")  <-- change if you want to allow single-label like "localhost"
// - labels 1..63 chars, alnum or '-', not starting/ending with '-'
func isValidHostname(h string) bool {
	if len(h) == 0 || len(h) > 253 {
		return false
	}
	if strings.HasPrefix(h, ".") || strings.HasSuffix(h, ".") {
		return false
	}
	if !strings.Contains(h, ".") { // require FQDN-like hostnames; relax if desired
		return false
	}
	labels := strings.Split(h, ".")
	for _, l := range labels {
		if !isValidLabel(l) {
			return false
		}
	}
	return true
}

func isValidLabel(l string) bool {
	if len(l) == 0 || len(l) > 63 {
		return false
	}
	if l[0] == '-' || l[len(l)-1] == '-' {
		return false
	}
	for i := 0; i < len(l); i++ {
		b := l[i]
		switch {
		case b >= 'a' && b <= 'z',
			b >= 'A' && b <= 'Z',
			b >= '0' && b <= '9',
			b == '-':
			// ok
		default:
			return false
		}
	}
	return true
}
