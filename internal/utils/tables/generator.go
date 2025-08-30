/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package tables

import (
	"fmt"
	"io"
	"reflect"

	"github.com/jedib0t/go-pretty/v6/table"
)

// Column defines one column for rows of type T.
type Column[T any] struct {
	Header string      // header text
	Cell   func(T) any // maps a row to a cell value
	Width  *int        // optional: min width (0 = auto)
}

// Option configures a render.
type Option func(*config)

type config struct {
	title        string
	style        *table.Style
	rowSeparator bool
}

func WithTitle(title string) Option   { return func(c *config) { c.title = title } }
func WithStyle(s table.Style) Option  { return func(c *config) { c.style = &s } }
func WithRowSeparator(on bool) Option { return func(c *config) { c.rowSeparator = on } }

// Render writes a table to w for the given columns & rows.
func Render[T any](w io.Writer, rows []T, cols []Column[T], opts ...Option) {
	cfg := config{
		style:        &table.StyleRounded,
		rowSeparator: false,
	}
	for _, o := range opts {
		o(&cfg)
	}

	tw := table.NewWriter()
	tw.SetOutputMirror(w)
	if cfg.style != nil {
		tw.SetStyle(*cfg.style)
	}
	if cfg.title != "" {
		tw.SetTitle(cfg.title)
	}

	// Headers
	hdr := make(table.Row, len(cols))
	for i, c := range cols {
		hdr[i] = c.Header
	}
	tw.AppendHeader(hdr)

	// Column configs
	var colCfgs []table.ColumnConfig
	for idx, c := range cols {
		cc := table.ColumnConfig{Number: idx + 1}
		if c.Width != nil && *c.Width > 0 {
			cc.WidthMin = *c.Width
		}
		colCfgs = append(colCfgs, cc)
	}
	if len(colCfgs) > 0 {
		tw.SetColumnConfigs(colCfgs)
	}

	// Rows
	for _, r := range rows {
		row := make(table.Row, len(cols))
		for i, c := range cols {
			row[i] = c.Cell(r)
		}
		tw.AppendRow(row)
		if cfg.rowSeparator {
			tw.AppendSeparator()
		}
	}

	tw.Render()
}

// Build returns the rendered table as a string.
func Build[T any](rows []T, cols []Column[T], opts ...Option) string {
	var b stringsBuilder
	Render(&b, rows, cols, opts...)
	return b.String()
}

// ColumnsByFieldNames builds columns from struct field names on T.
// Each field is fetched via reflection; Header defaults to the field name.
func ColumnsByFieldNames[T any](fieldNames ...string) []Column[T] {
	cols := make([]Column[T], 0, len(fieldNames))
	for _, name := range fieldNames {
		n := name // capture
		cols = append(cols, Column[T]{
			Header: n,
			Cell: func(v T) any {
				rv := reflect.ValueOf(v)
				for rv.Kind() == reflect.Pointer {
					if rv.IsNil() {
						return nil
					}
					rv = rv.Elem()
				}
				if rv.Kind() != reflect.Struct {
					return fmt.Sprintf("<%T is not a struct>", v)
				}
				f := rv.FieldByName(n)
				if !f.IsValid() {
					return fmt.Sprintf("<no field %q>", n)
				}
				return f.Interface()
			},
		})
	}
	return cols
}

// Helper to avoid importing bytes just to collect strings.
type stringsBuilder struct{ b []byte }

func (s *stringsBuilder) Write(p []byte) (int, error) { s.b = append(s.b, p...); return len(p), nil }
func (s *stringsBuilder) String() string              { return string(s.b) }
