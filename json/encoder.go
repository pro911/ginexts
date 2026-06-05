package json

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
)

// ——— helpers ———

func containsAF(s string) bool {
	for _, c := range s {
		if c >= 'a' && c <= 'f' {
			return true
		}
	}
	return false
}

// ——— HexStringCodec ———

// HexStringCodec converts int64 / *int64 to/from hexadecimal string.
type HexStringCodec struct {
	IsInt64Pointer bool
}

func (e *HexStringCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	if e.IsInt64Pointer {
		pp := (**int64)(ptr)
		if *pp == nil {
			stream.WriteNil()
			return
		}
		stream.WriteString(fmt.Sprintf("%x", **pp))
		return
	}
	v := *(*int64)(ptr)
	if v == 0 {
		stream.WriteString("0")
	} else {
		stream.WriteString(fmt.Sprintf("%x", v))
	}
}

func (e *HexStringCodec) IsEmpty(ptr unsafe.Pointer) bool {
	return ptr == nil || *(*int64)(ptr) == 0
}

func (e *HexStringCodec) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	var i int64
	switch iter.WhatIsNext() {
	case jsoniter.StringValue:
		s := iter.ReadString()
		if len(s) < 17 || containsAF(s) {
			i, _ = strconv.ParseInt(s, 16, 64)
		} else {
			i, _ = strconv.ParseInt(s, 10, 64)
		}
	case jsoniter.NumberValue:
		i = iter.ReadInt64()
	}

	if e.IsInt64Pointer {
		v := new(int64)
		*v = i
		*(*unsafe.Pointer)(ptr) = unsafe.Pointer(v)
	} else {
		*((*int64)(ptr)) = i
	}
}

// ——— EmptyObjectEncoder ———

type EmptyObjectEncoder struct {
	Encoder jsoniter.ValEncoder
}

func (e *EmptyObjectEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	if *(*uintptr)(ptr) == 0 {
		stream.WriteRaw("{}")
		return
	}
	e.Encoder.Encode(ptr, stream)
}

func (e *EmptyObjectEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return e.Encoder.IsEmpty(ptr)
}

// ——— EmptyArrayEncoder (generic) ———

type EmptyArrayEncoder struct {
	Encoder jsoniter.ValEncoder
}

func (e *EmptyArrayEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	if *(*uintptr)(ptr) == 0 {
		stream.WriteRaw("[]")
		return
	}
	e.Encoder.Encode(ptr, stream)
}

func (e *EmptyArrayEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return e.Encoder.IsEmpty(ptr)
}

// ——— EmptyArrayInt64Codec (auto hex + nil->[]) ———

type EmptyArrayInt64Codec struct {
	Encoder jsoniter.ValEncoder
	Decoder jsoniter.ValDecoder
}

var internalAPI = jsoniter.Config{}.Froze()

func (e *EmptyArrayInt64Codec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	if *(*uintptr)(ptr) == 0 {
		stream.WriteRaw("[]")
		return
	}
	slice := (*[]int64)(ptr)
	strs := make([]string, len(*slice))
	for i, v := range *slice {
		if v == 0 {
			strs[i] = "0"
		} else {
			strs[i] = fmt.Sprintf("%x", v)
		}
	}
	if b, err := internalAPI.Marshal(strs); err == nil {
		stream.WriteRaw(string(b))
	} else {
		e.Encoder.Encode(ptr, stream)
	}
}

func (e *EmptyArrayInt64Codec) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	var vals []int64
	for iter.ReadArray() {
		val := iter.Read()
		if iter.Error != nil {
			break
		}
		if s, ok := val.(string); ok {
			if len(s) < 17 || containsAF(s) {
				if v, err := strconv.ParseInt(s, 16, 64); err == nil {
					vals = append(vals, v)
				}
			} else {
				if v, err := strconv.ParseInt(s, 10, 64); err == nil {
					vals = append(vals, v)
				}
			}
		}
	}
	reflect.ValueOf((*[]int64)(ptr)).Elem().Set(reflect.ValueOf(vals))
}

func (e *EmptyArrayInt64Codec) IsEmpty(ptr unsafe.Pointer) bool {
	return e.Encoder.IsEmpty(ptr)
}

// ——— ToStringDecoder ———

type ToStringDecoder struct{}

func (e *ToStringDecoder) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	if iter.WhatIsNext() == jsoniter.StringValue {
		*((*string)(ptr)) = iter.ReadString()
	} else {
		*((*string)(ptr)) = iter.ReadAny().ToString()
	}
}

func (e *ToStringDecoder) IsEmpty(ptr unsafe.Pointer) bool { return false }

// ——— ToBoolDecoder ———

type ToBoolDecoder struct {
	Decoder    jsoniter.ValDecoder
	DefaultVal bool
}

func (e *ToBoolDecoder) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	switch iter.WhatIsNext() {
	case jsoniter.BoolValue:
		e.Decoder.Decode(ptr, iter)
	case jsoniter.NumberValue:
		*((*bool)(ptr)) = iter.ReadInt() > 0
	case jsoniter.StringValue:
		s := strings.ToLower(iter.ReadString())
		switch s {
		case "true", "1":
			*((*bool)(ptr)) = true
		case "false", "0", "-1":
			*((*bool)(ptr)) = false
		default:
			*((*bool)(ptr)) = e.DefaultVal
		}
	default:
		s := strings.ToLower(iter.ReadAny().ToString())
		switch s {
		case "true", "1":
			*((*bool)(ptr)) = true
		case "false", "0", "-1":
			*((*bool)(ptr)) = false
		default:
			*((*bool)(ptr)) = e.DefaultVal
		}
	}
}

// ——— StringMaxDecoder ———
// Truncates a string to the specified maximum rune count on decode.
// Usage: ` + "`json:"field,stringmax=100"`" + `

const stringMaxTagKey = "stringmax="

type StringMaxDecoder struct {
	Decoder    jsoniter.ValDecoder
	jsonTagStr string
}

func (s *StringMaxDecoder) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	oldStr := iter.ReadString()
	newStr := s.truncateString(oldStr)
	log.Printf("StringMax.Decode: s.jsonTagStr: %v;\n newStr: %s;\n oldStr: %s\n", s.jsonTagStr, newStr, oldStr)
	*((*string)(ptr)) = newStr
}

func (s *StringMaxDecoder) truncateString(str string) string {
	startIndex := strings.Index(s.jsonTagStr, stringMaxTagKey)
	if startIndex == -1 {
		log.Printf("json.stringmax jsonTagStr:%v err: %s \n", s.jsonTagStr, "missing stringmax=N parameter")
		return str
	}

	startIndex += len(stringMaxTagKey)
	remaining := s.jsonTagStr[startIndex:]

	endIndex := strings.Index(remaining, ",")
	if endIndex == -1 {
		endIndex = len(remaining)
	}

	numberStr := remaining[:endIndex]
	maxVal, err := strconv.Atoi(numberStr)
	if err != nil {
		log.Printf("json.stringmax jsonTagStr:%v err converting string to int: %s \n", s.jsonTagStr, err.Error())
		return str
	}

	if utf8.RuneCountInString(str) <= maxVal {
		return str
	}

	runeStr := []rune(str)
	if len(runeStr) > maxVal {
		runeStr = runeStr[:maxVal]
	}
	return string(runeStr)
}
