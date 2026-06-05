// Package form extends gin's form/query binding with hexstring tag support.
//
// It replaces gin's default Form, Query, FormPost, and FormMultipart bindings
// with hex-aware wrappers that convert hexadecimal string values to decimal
// before the actual binding occurs.
//
// Usage: import ginexts (the parent package) to activate:
//
//	import _ "github.com/pro911/ginexts"
//
// Then use hexstring in form tags:
//
//	type User struct {
//	    ProjectId int64  ` + "`form:"project_id,hexstring"`" + `
//	    UserId    *int64 ` + "`form:"user_id,hexstring"`" + `
//	    Ids       []int64 ` + "`form:"ids,hexstring"`" + `
//	}
//
// Supported types: int/int8/16/32/64, uint/uint8/16/32/64, *int64/*uint64, []int64/[]uint64.
// The hexstring option supports "0x"/"0X" prefix.
//
// This package is completely non-invasive — zero modification to gin or business code.
package form

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin/binding"
)

func init() {
	binding.Form = &hexBinding{name: "form-hex", delegate: binding.Form, tag: "form"}
	binding.Query = &hexBinding{name: "query-hex", delegate: binding.Query, tag: "form"}
	binding.FormPost = &hexBinding{name: "form-post-hex", delegate: binding.FormPost, tag: "form"}
	binding.FormMultipart = &hexBinding{name: "form-multipart-hex", delegate: binding.FormMultipart, tag: "form"}
}

// ——— hexBinding ———

type hexBinding struct {
	name     string
	delegate binding.Binding
	tag      string
}

func (h *hexBinding) Name() string { return h.name }

func (h *hexBinding) Bind(req *http.Request, obj any) error {
	_ = req.ParseForm()
	_ = req.ParseMultipartForm(defaultMaxMemory)

	preprocess(obj, req.Form, h.tag)
	if req.PostForm != nil {
		preprocess(obj, req.PostForm, h.tag)
	}
	if req.MultipartForm != nil {
		preprocess(obj, req.MultipartForm.Value, h.tag)
	}
	if req.URL != nil {
		queryVals := req.URL.Query()
		preprocess(obj, queryVals, h.tag)
		req.URL.RawQuery = queryVals.Encode()
	}

	return h.delegate.Bind(req, obj)
}

const defaultMaxMemory = 32 << 20

// ——— preprocess ———

func preprocess(obj any, form map[string][]string, tag string) {
	if obj == nil || form == nil {
		return
	}
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}
	walkStruct(v, form, tag)
}

func walkStruct(v reflect.Value, form map[string][]string, tag string) {
	t := v.Type()
	for i := range v.NumField() {
		sf := t.Field(i)
		fieldVal := v.Field(i)

		if sf.Anonymous && fieldVal.Kind() == reflect.Struct {
			walkStruct(fieldVal, form, tag)
			continue
		}
		if sf.PkgPath != "" {
			continue
		}

		tagValue := sf.Tag.Get(tag)
		if tagValue == "" || tagValue == "-" {
			continue
		}

		fieldName, opts := head(tagValue, ",")
		if fieldName == "" {
			fieldName = sf.Name
		}

		if !hasHexstring(opts) {
			continue
		}

		if vals, ok := form[fieldName]; ok {
			form[fieldName] = convertHexValues(vals, fieldVal)
		}
	}
}

func hasHexstring(opts string) bool {
	for len(opts) > 0 {
		var opt string
		opt, opts = head(opts, ",")
		if opt == "hexstring" {
			return true
		}
	}
	return false
}

func convertHexValues(vals []string, field reflect.Value) []string {
	elemKind := field.Kind()
	if elemKind == reflect.Ptr {
		elemKind = field.Type().Elem().Kind()
	}
	if elemKind == reflect.Slice || elemKind == reflect.Array {
		elemKind = field.Type().Elem().Kind()
	}

	if !isIntKind(elemKind) && !isUintKind(elemKind) {
		return vals
	}

	result := make([]string, len(vals))
	for i, v := range vals {
		result[i] = hexToDec(v)
	}
	return result
}

func hexToDec(s string) string {
	if s == "" {
		return s
	}
	cleaned := strings.TrimPrefix(strings.TrimPrefix(s, "0x"), "0X")
	n, err := strconv.ParseInt(cleaned, 16, 64)
	if err != nil {
		return s
	}
	return strconv.FormatInt(n, 10)
}

func isIntKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	}
	return false
}

func isUintKind(k reflect.Kind) bool {
	switch k {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	}
	return false
}

func head(s, sep string) (head, tail string) {
	head, tail, _ = strings.Cut(s, sep)
	return head, tail
}
