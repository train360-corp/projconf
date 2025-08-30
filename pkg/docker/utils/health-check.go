/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package utils

import (
	"context"
	"net/http"
	"time"
)

func HttpHealthCheck(c context.Context, endpoint string) (bool, int) {
	ctx, cancel := context.WithTimeout(c, 3*time.Second)
	defer cancel()
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		panic(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, -1
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, resp.StatusCode
}
