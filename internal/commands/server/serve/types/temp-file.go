/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package types

type TempFile struct {
	Name          string
	ContainerPath string
	Data          []byte
}
