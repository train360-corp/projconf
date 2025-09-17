/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package validators

import (
	"regexp"
)

var (
	validIpAddressRegex = regexp.MustCompile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	validHostnameRegex  = regexp.MustCompile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)
)

// IsValidHost checks if s is a valid IPv4 address or hostname (no port allowed).
func IsValidHost(s string) bool {
	return validIpAddressRegex.MatchString(s) || validHostnameRegex.MatchString(s)
}

// IsValidPort returns true if p is a valid TCP/UDP port (1–65535).
func IsValidPort(p uint16) bool {
	// NOTE: max p for uint16 is 65535, so no need to check
	// return p >= 1 && p <= 65535
	return p >= 1
}
