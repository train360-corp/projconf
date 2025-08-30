/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package api

import (
	_ "embed"
	"github.com/getkin/kin-openapi/openapi3"
)

//go:embed openapi.yaml
var openAPISpec []byte

func MustSpec() *openapi3.T {
	ldr := &openapi3.Loader{IsExternalRefsAllowed: false}
	doc, err := ldr.LoadFromData(openAPISpec)
	if err != nil {
		panic(err)
	}
	if err := doc.Validate(ldr.Context); err != nil {
		panic(err)
	}
	return doc
}
