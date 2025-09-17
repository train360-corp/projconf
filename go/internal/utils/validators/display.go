/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package validators

import (
	"regexp"
	"strings"
)

var displayRegex = regexp.MustCompile(`^[[:alnum:] _]+$`)
var variableRegex = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

func IsValidDisplay(s string) bool {
	return len(strings.Trim(s, " ")) > 0 && displayRegex.MatchString(s)
}

func IsValidVariable(s string) bool {
	return len(strings.Trim(s, " ")) > 0 && variableRegex.MatchString(s)
}
