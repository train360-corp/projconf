/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package validators

import "github.com/google/uuid"

func IsValidUUID(str string) bool {
	_, err := uuid.Parse(str)
	return err == nil
}
