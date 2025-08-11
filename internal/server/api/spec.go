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
