package flop

import (
	"fmt"
	"slices"

	"github.com/ysomad/flop/expr"
	"github.com/ysomad/flop/filter"
	"github.com/ysomad/flop/internal/ident"
)

// Col declares one trusted storage column exposed for ordering.
type Col[T any] struct {
	name  string
	key   bool
	value func(T) expr.Value
}

// Sortable declares an orderable column.
func Sortable[T any](name string, value func(T) expr.Value) Col[T] {
	return Col[T]{name: name, value: value}
}

// Key declares an orderable column that makes cursor ordering unique.
func Key[T any](name string, value func(T) expr.Value) Col[T] {
	return Col[T]{name: name, key: true, value: value}
}

// Asc creates an ascending order field.
func Asc(path string) Field { return Field{Path: path} }

// Desc creates a descending order field.
func Desc(path string) Field { return Field{Path: path, Desc: true} }

// Paginator validates and resolves one endpoint's pagination policy.
type Paginator[T any] struct {
	cols         map[string]Col[T]
	keys         []string
	filterSchema *filter.Schema
	defaultOrder OrderBy
	defaultSize  int32
	maxSize      int32
}

// NewPaginator validates and copies an endpoint declaration.
func NewPaginator[T any](
	cols []Col[T],
	schema *filter.Schema,
	defaultOrder []Field,
	defaultSize int32,
	maxSize int32,
) (*Paginator[T], error) {
	if len(cols) == 0 {
		return nil, fmt.Errorf("flop: paginator has no columns")
	}
	if !schema.Initialized() {
		return nil, fmt.Errorf("flop: paginator has an uninitialized filter schema")
	}
	if len(defaultOrder) == 0 {
		return nil, fmt.Errorf("flop: paginator has no default order")
	}
	if defaultSize <= 0 || maxSize <= 0 || defaultSize > maxSize {
		return nil, fmt.Errorf("flop: invalid page-size limits")
	}

	pag := &Paginator[T]{
		cols:         make(map[string]Col[T], len(cols)),
		filterSchema: schema,
		defaultSize:  defaultSize,
		maxSize:      maxSize,
	}

	for _, col := range cols {
		if !ident.ValidPath(col.name) {
			return nil, fmt.Errorf("flop: invalid column %q", col.name)
		}
		if col.value == nil {
			return nil, fmt.Errorf("flop: column %q has no value function", col.name)
		}
		if _, exists := pag.cols[col.name]; exists {
			return nil, fmt.Errorf("flop: duplicate column %q", col.name)
		}
		pag.cols[col.name] = col
		if col.key {
			pag.keys = append(pag.keys, col.name)
		}
	}
	if len(pag.keys) == 0 {
		return nil, fmt.Errorf("flop: paginator has no key columns")
	}

	resolvedDefault, err := pag.resolveOrder(OrderBy{Fields: slices.Clone(defaultOrder)})
	if err != nil {
		return nil, fmt.Errorf("flop: invalid default order: %w", err)
	}
	pag.defaultOrder = resolvedDefault

	return pag, nil
}

// Cursor creates validated cursor parameters for a request.
func (p *Paginator[T]) Cursor(cursor string, pageSize int32, ord OrderBy, fil filter.Filter) (CursorParams, error) {
	pageSize, err := p.normalizePageSize(pageSize)
	if err != nil {
		return CursorParams{}, err
	}
	ord, err = p.resolveOrder(ord)
	if err != nil {
		return CursorParams{}, err
	}
	fil, err = p.filterSchema.Validate(fil)
	if err != nil {
		return CursorParams{}, err
	}
	params, err := NewCursorParams(cursor, pageSize, ord, fil)
	if err != nil {
		return CursorParams{}, err
	}
	if err := validateCursorPos(params.After, ord, false); err != nil {
		return CursorParams{}, err
	}
	return params, nil
}

// Offset creates validated page-number parameters for a request.
func (p *Paginator[T]) Offset(page, pageSize int32, ord OrderBy, fil filter.Filter) (Params, error) {
	pageSize, err := p.normalizePageSize(pageSize)
	if err != nil {
		return Params{}, err
	}
	ord, err = p.resolveOrder(ord)
	if err != nil {
		return Params{}, err
	}
	fil, err = p.filterSchema.Validate(fil)
	if err != nil {
		return Params{}, err
	}
	params, err := NewParams(page, pageSize)
	if err != nil {
		return Params{}, err
	}
	params.orderBy = ord.clone()
	params.filter = fil
	return params, nil
}

// CursorPage creates a page and derives its next position from declared columns.
func (p *Paginator[T]) CursorPage(rows []T, params CursorParams) (CursorPage[T], error) {
	orderBy := params.OrderBy()
	if len(orderBy.Fields) == 0 {
		return CursorPage[T]{}, fmt.Errorf("flop: cursor parameters have no order")
	}

	return NewCursorPage(rows, params, func(row T) (CursorPosition, error) {
		pos := make(CursorPosition, len(orderBy.Fields))

		for i, fld := range orderBy.Fields {
			col, ok := p.cols[fld.Path]
			if !ok {
				return nil, fmt.Errorf("flop: cursor column %q is not declared", fld.Path)
			}
			value := col.value(row)
			if value.Kind() == expr.NullKind {
				return nil, fmt.Errorf("flop: cursor column %q is null", fld.Path)
			}
			pos[i] = CursorField{Path: fld.Path, Value: value}
		}
		return pos, nil
	})
}

func (p *Paginator[T]) normalizePageSize(size int32) (int32, error) {
	if size < 0 {
		return 0, fmt.Errorf("flop: page size must not be negative")
	}
	if size == 0 {
		return p.defaultSize, nil
	}
	if size > p.maxSize {
		return p.maxSize, nil
	}
	return size, nil
}

func (p *Paginator[T]) resolveOrder(orderBy OrderBy) (OrderBy, error) {
	fields := orderBy.Fields
	if len(fields) == 0 {
		fields = p.defaultOrder.Fields
	}
	resolved := make([]Field, 0, len(fields)+len(p.keys))
	seen := make(map[string]struct{}, len(fields)+len(p.keys))
	for _, field := range fields {
		if !ident.ValidPath(field.Path) {
			return OrderBy{}, fmt.Errorf("flop: invalid order field %q", field.Path)
		}
		if _, ok := p.cols[field.Path]; !ok {
			return OrderBy{}, fmt.Errorf("flop: unknown order field %q", field.Path)
		}
		if _, duplicate := seen[field.Path]; duplicate {
			return OrderBy{}, fmt.Errorf("flop: duplicate order field %q", field.Path)
		}
		seen[field.Path] = struct{}{}
		resolved = append(resolved, field)
	}
	for _, key := range p.keys {
		if _, exists := seen[key]; exists {
			continue
		}
		resolved = append(resolved, Field{Path: key})
	}
	return OrderBy{Fields: resolved}, nil
}

func validateCursorPos(pos CursorPosition, ord OrderBy, require bool) error {
	if len(pos) == 0 && !require {
		return nil
	}
	if len(pos) != len(ord.Fields) {
		return fmt.Errorf("flop: cursor position does not match the resolved order")
	}
	for i, field := range pos {
		if field.Path != ord.Fields[i].Path {
			return fmt.Errorf("flop: cursor position does not match the resolved order")
		}
		if field.Value.Kind() == expr.NullKind {
			return fmt.Errorf("flop: cursor column %q is null", field.Path)
		}
	}
	return nil
}
