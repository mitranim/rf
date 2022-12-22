package rf

import (
	r "reflect"
	u "unsafe"
)

const expectedStructNesting = 8

func isFieldValid(val r.StructField) bool {
	return val.Type != nil && len(val.Index) > 0
}

var typeFieldsCache = Cache{Func: func(typ r.Type) any {
	typ = ValidateTypeStruct(typ)
	out := make([]r.StructField, 0, typ.NumField())
	for ind := range Iter(typ.NumField()) {
		out = append(out, typ.Field(ind))
	}
	return out
}}

var typeDeepFieldsCache = Cache{Func: func(typ r.Type) any {
	typ = ValidateTypeStruct(typ)
	buf := make([]r.StructField, 0, typ.NumField())
	path := make(Path, 0, expectedStructNesting)
	appendStructDeepFields(&buf, &path, r.StructField{Type: typ, Anonymous: true})
	return buf
}}

func appendStructDeepFields(
	buf *[]r.StructField, path *Path, field r.StructField,
) {
	defer path.Add(field.Index).Reset()

	typ := TypeDeref(field.Type)
	if IsEmbed(field) {
		for _, inner := range TypeFields(typ) {
			inner.Offset += field.Offset
			appendStructDeepFields(buf, path, inner)
		}
	} else {
		field.Index = path.Copy()
		*buf = append(*buf, field)
	}
}

var typeOffsetFieldsCache = Cache{Func: func(typ r.Type) any {
	typ = ValidateTypeStruct(typ)
	if typ == nil {
		return map[uintptr][]r.StructField(nil)
	}

	fields := TypeDeepFields(typ)
	out := make(map[uintptr][]r.StructField, len(fields))
	for _, field := range fields {
		out[field.Offset] = append(out[field.Offset], field)
	}
	return out
}}

func cast[Out, Src any](val Src) Out { return *(*Out)(u.Pointer(&val)) }
