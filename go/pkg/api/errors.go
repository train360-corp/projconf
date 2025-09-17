/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package api

import (
	"fmt"
	"reflect"
)

type APIWithError struct {
	JSON400 *Error
	JSON401 *Error
	JSON403 *Error
	JSON500 *Error
	Body    []byte
}

func asAPIWithError(v any) *APIWithError {
	if v == nil {
		return nil
	}
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}

	get := func(name string) reflect.Value {
		f := val.FieldByName(name)
		if f.IsValid() && !f.IsZero() {
			return f
		}
		return reflect.Value{}
	}

	out := &APIWithError{}
	if f := get("JSON400"); f.IsValid() {
		if p, ok := f.Interface().(*Error); ok {
			out.JSON400 = p
		}
	}
	if f := get("JSON401"); f.IsValid() {
		if p, ok := f.Interface().(*Error); ok {
			out.JSON401 = p
		}
	}
	if f := get("JSON403"); f.IsValid() {
		if p, ok := f.Interface().(*Error); ok {
			out.JSON403 = p
		}
	}
	if f := get("JSON500"); f.IsValid() {
		if p, ok := f.Interface().(*Error); ok {
			out.JSON500 = p
		}
	}
	if f := get("Body"); f.IsValid() {
		if p, ok := f.Interface().([]byte); ok {
			out.Body = p
		}
	}

	if out.JSON400 == nil && out.JSON401 == nil && out.JSON403 == nil && out.JSON500 == nil && len(out.Body) == 0 {
		return nil
	}
	return out
}

func getAPIError(resp *APIWithError) string {

	if resp == nil {
		return "unformattable error"
	}

	if resp.JSON401 != nil {
		return fmt.Sprintf("%v: %v", resp.JSON401.Error, resp.JSON401.Description)
	} else if resp.JSON403 != nil {
		return fmt.Sprintf("%v: %v", resp.JSON403.Error, resp.JSON403.Description)
	} else if resp.JSON500 != nil {
		return fmt.Sprintf("%v: %v", resp.JSON500.Error, resp.JSON500.Description)
	}
	return string(resp.Body)
}

func GetAPIError(v any) string {
	return getAPIError(asAPIWithError(v))
}
