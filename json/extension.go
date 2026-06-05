// Package json overrides gin's default JSON codec with jsoniter + custom extensions.
//
// Import this package to activate all custom JSON tag features:
//
//	import _ "github.com/pro911/ginexts/json"
//
// Or use the top-level ginexts package:
//
//	import _ "github.com/pro911/ginexts"
//
// Supported JSON tag options:
//   - hexstring:   int64 / *int64 / []int64 <-> hexadecimal string
//   - emptyobject: nil pointer/interface/slice/bytes -> {}
//   - emptyarray:  nil slice -> []
//   - tostring:    any JSON value -> string on decode
//   - stringmax=N: truncate string to N runes on decode
//   - tofalse:     any JSON value -> bool (default: false) on decode
//   - totrue:      any JSON value -> bool (default: true) on decode
//
// For programmatic use outside gin, use GetAPI():
//
//	api := json.GetAPI()
//	api.Marshal(obj)
package json

import (
	"reflect"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

// Extension inspects struct field JSON tags and applies custom encoders/decoders.
type Extension struct {
	jsoniter.DummyExtension
}

// UpdateStructDescriptor replaces default encoder/decoder for fields with custom tag options.
func (ext *Extension) UpdateStructDescriptor(desc *jsoniter.StructDescriptor) {
	for _, binding := range desc.Fields {
		tagStr := binding.Field.Tag().Get("json")

		switch binding.Field.Type().Kind() {
		case reflect.Int64:
			if strings.Contains(tagStr, "hexstring") {
				binding.Encoder = &HexStringCodec{}
				binding.Decoder = &HexStringCodec{}
			}

		case reflect.Ptr:
			if strings.Contains(tagStr, "hexstring") && binding.Field.Type().Type1().Elem().Kind() == reflect.Int64 {
				binding.Encoder = &HexStringCodec{IsInt64Pointer: true}
				binding.Decoder = &HexStringCodec{IsInt64Pointer: true}
			} else if strings.Contains(tagStr, "emptyobject") {
				binding.Encoder = &EmptyObjectEncoder{Encoder: binding.Encoder}
			}

		case reflect.Interface:
			if strings.Contains(tagStr, "emptyobject") {
				binding.Encoder = &EmptyObjectEncoder{Encoder: binding.Encoder}
			}

		case reflect.Slice, reflect.Array:
			elemKind := binding.Field.Type().Type1().Elem().Kind()
			if elemKind == reflect.Int64 {
				codec := &EmptyArrayInt64Codec{Encoder: binding.Encoder, Decoder: binding.Decoder}
				binding.Encoder = codec
				binding.Decoder = codec
			} else if strings.Contains(tagStr, "emptyarray") {
				binding.Encoder = &EmptyArrayEncoder{Encoder: binding.Encoder}
			} else if strings.Contains(tagStr, "emptyobject") {
				binding.Encoder = &EmptyObjectEncoder{Encoder: binding.Encoder}
			}

		case reflect.String:
			if strings.Contains(tagStr, "tostring") {
				binding.Decoder = &ToStringDecoder{}
			}
			if strings.Contains(tagStr, "stringmax=") {
				binding.Decoder = &StringMaxDecoder{Decoder: binding.Decoder, jsonTagStr: tagStr}
			}

		case reflect.Bool:
			if strings.Contains(tagStr, "tofalse") {
				binding.Decoder = &ToBoolDecoder{Decoder: binding.Decoder, DefaultVal: false}
			} else if strings.Contains(tagStr, "totrue") {
				binding.Decoder = &ToBoolDecoder{Decoder: binding.Decoder, DefaultVal: true}
			}
		}
	}
}

// NewAPI creates a new jsoniter.API with all extensions pre-registered.
func NewAPI() jsoniter.API {
	api := jsoniter.Config{}.Froze()
	api.RegisterExtension(&Extension{})
	return api
}
