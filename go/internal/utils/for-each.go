/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package utils

func ForEach[M any, T any](items []M, converter func(M) T) []T {
	converted := make([]T, len(items))
	for i, item := range items {
		converted[i] = converter(item)
	}
	return converted
}
