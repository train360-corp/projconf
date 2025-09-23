/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/train360-corp/projconf/go/internal/utils"
	"github.com/train360-corp/projconf/go/pkg/api"
	"github.com/train360-corp/projconf/go/pkg/postgrest"
)

var preferFull = utils.Ptr(postgrest.PostProjectsParamsPrefer("return=representation"))

var success = api.SuccessResponse{Status: "success"}

// parse takes JSON bytes and unmarshals into a new T
func parse[T any](data []byte) (*T, error) {
	var obj T
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

// parseOne takes JSON bytes of an array, and unmarshals into a new T,
// expecting only a single T (fails if len(data) != 1)
func parseOne[T any](data []byte) (*T, error) {
	objs, err := parse[[]T](data)
	if err != nil {
		return nil, err
	}
	if len(*objs) != 1 {
		return nil, fmt.Errorf("expected 1 object, got %d", len(*objs))
	}
	return &(*objs)[0], nil
}
