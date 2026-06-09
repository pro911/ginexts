package json

import (
	"bytes"
	"testing"
	"unsafe"
)

// ——— Test Structs ———

type hexInt64Struct struct {
	Val int64 `json:"val,hexstring"`
}

type hexPtrInt64Struct struct {
	Val *int64 `json:"val,hexstring"`
}

type emptyObjectStruct struct {
	Val *string `json:"val,emptyobject"`
}

type emptyArrayStruct struct {
	Val []int `json:"val,emptyarray"`
}

type toStringStruct struct {
	Val string `json:"val,tostring"`
}

type toFalseStruct struct {
	Val bool `json:"val,tofalse"`
}

type toTrueStruct struct {
	Val bool `json:"val,totrue"`
}

type stringMaxStruct struct {
	Val string `json:"val,stringmax=5"`
}

type combinedStruct struct {
	HexID     int64   `json:"hex_id,hexstring"`
	PtrID     *int64  `json:"ptr_id,hexstring"`
	EmptyObj  *string `json:"empty_obj,emptyobject"`
	EmptyArr  []int   `json:"empty_arr,emptyarray"`
	ToStr     string  `json:"to_str,tostring"`
	ToFalse   bool    `json:"to_false,tofalse"`
	ToTrue    bool    `json:"to_true,totrue"`
	MaxStr    string  `json:"max_str,stringmax=3"`
}

// ——— hexstring int64 ———

func TestHexInt64Encode(t *testing.T) {
	api := NewAPI()
	s := hexInt64Struct{Val: 255}
	data, err := api.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"val":"ff"}` {
		t.Errorf("expected {\"val\":\"ff\"}, got %s", data)
	}
}

func TestHexInt64EncodeZero(t *testing.T) {
	api := NewAPI()
	s := hexInt64Struct{Val: 0}
	data, err := api.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"val":"0"}` {
		t.Errorf("expected {\"val\":\"0\"}, got %s", data)
	}
}

func TestHexInt64DecodeHexString(t *testing.T) {
	api := NewAPI()
	var s hexInt64Struct
	err := api.Unmarshal([]byte(`{"val":"ff"}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Val != 255 {
		t.Errorf("expected 255, got %d", s.Val)
	}
}

func TestHexInt64DecodeShortStringAsHex(t *testing.T) {
	// Short strings (< 17 chars) without a-f are treated as hex.
	api := NewAPI()
	var s hexInt64Struct
	err := api.Unmarshal([]byte(`{"val":"255"}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	// 0x255 = 597
	if s.Val != 597 {
		t.Errorf("expected 597, got %d", s.Val)
	}
}

func TestHexInt64DecodeNumber(t *testing.T) {
	api := NewAPI()
	var s hexInt64Struct
	err := api.Unmarshal([]byte(`{"val":255}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Val != 255 {
		t.Errorf("expected 255, got %d", s.Val)
	}
}

func TestHexInt64DecodeLongDecimalString(t *testing.T) {
	// A 17+ digit string without a-f chars should be treated as decimal.
	api := NewAPI()
	var s hexInt64Struct
	err := api.Unmarshal([]byte(`{"val":"12345678901234567"}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Val != 12345678901234567 {
		t.Errorf("expected 12345678901234567, got %d", s.Val)
	}
}

// ——— hexstring *int64 ———

func TestHexPtrInt64Encode(t *testing.T) {
	api := NewAPI()
	v := int64(255)
	s := hexPtrInt64Struct{Val: &v}
	data, err := api.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"val":"ff"}` {
		t.Errorf("expected {\"val\":\"ff\"}, got %s", data)
	}
}

func TestHexPtrInt64EncodeNil(t *testing.T) {
	api := NewAPI()
	s := hexPtrInt64Struct{Val: nil}
	data, err := api.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"val":null}` {
		t.Errorf("expected {\"val\":null}, got %s", data)
	}
}

func TestHexPtrInt64Decode(t *testing.T) {
	api := NewAPI()
	var s hexPtrInt64Struct
	err := api.Unmarshal([]byte(`{"val":"ff"}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Val == nil || *s.Val != 255 {
		t.Errorf("expected pointer to 255, got %v", s.Val)
	}
}

// ——— emptyobject ———

func TestEmptyObjectEncodeNil(t *testing.T) {
	api := NewAPI()
	s := emptyObjectStruct{Val: nil}
	data, err := api.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"val":{}}` {
		t.Errorf("expected {\"val\":{}}, got %s", data)
	}
}

func TestEmptyObjectEncodeNonNil(t *testing.T) {
	api := NewAPI()
	str := "hello"
	s := emptyObjectStruct{Val: &str}
	data, err := api.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"val":"hello"}` {
		t.Errorf("expected {\"val\":\"hello\"}, got %s", data)
	}
}

// ——— emptyarray ———

func TestEmptyArrayEncodeNil(t *testing.T) {
	api := NewAPI()
	s := emptyArrayStruct{Val: nil}
	data, err := api.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"val":[]}` {
		t.Errorf("expected {\"val\":[]}, got %s", data)
	}
}

func TestEmptyArrayEncodeNonNil(t *testing.T) {
	api := NewAPI()
	s := emptyArrayStruct{Val: []int{1, 2, 3}}
	data, err := api.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"val":[1,2,3]}` {
		t.Errorf("expected {\"val\":[1,2,3]}, got %s", data)
	}
}

// ——— tostring ———

func TestToStringDecodeString(t *testing.T) {
	api := NewAPI()
	var s toStringStruct
	err := api.Unmarshal([]byte(`{"val":"hello"}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Val != "hello" {
		t.Errorf("expected 'hello', got '%s'", s.Val)
	}
}

func TestToStringDecodeNumber(t *testing.T) {
	api := NewAPI()
	var s toStringStruct
	err := api.Unmarshal([]byte(`{"val":123}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Val != "123" {
		t.Errorf("expected '123', got '%s'", s.Val)
	}
}

func TestToStringDecodeBool(t *testing.T) {
	api := NewAPI()
	var s toStringStruct
	err := api.Unmarshal([]byte(`{"val":true}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Val != "true" {
		t.Errorf("expected 'true', got '%s'", s.Val)
	}
}

// ——— tofalse ———

func TestToFalseDecodeTrue(t *testing.T) {
	api := NewAPI()
	var s toFalseStruct
	err := api.Unmarshal([]byte(`{"val":true}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	if !s.Val {
		t.Errorf("expected true, got false")
	}
}

func TestToFalseDecodeStringTrue(t *testing.T) {
	api := NewAPI()
	var s toFalseStruct
	err := api.Unmarshal([]byte(`{"val":"true"}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	if !s.Val {
		t.Errorf("expected true, got false")
	}
}

func TestToFalseDecodeString1(t *testing.T) {
	api := NewAPI()
	var s toFalseStruct
	err := api.Unmarshal([]byte(`{"val":"1"}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	if !s.Val {
		t.Errorf("expected true, got false")
	}
}

func TestToFalseDecodeStringFalse(t *testing.T) {
	api := NewAPI()
	var s toFalseStruct
	err := api.Unmarshal([]byte(`{"val":"false"}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Val {
		t.Errorf("expected false, got true")
	}
}

func TestToFalseDecodeUnknownStringDefaultsFalse(t *testing.T) {
	api := NewAPI()
	var s toFalseStruct
	err := api.Unmarshal([]byte(`{"val":"unknown"}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Val {
		t.Errorf("expected false (default for tofalse), got true")
	}
}

func TestToFalseDecodeNumber(t *testing.T) {
	api := NewAPI()
	var s toFalseStruct
	err := api.Unmarshal([]byte(`{"val":1}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	if !s.Val {
		t.Errorf("expected true, got false")
	}
}

func TestToFalseDecodeZero(t *testing.T) {
	api := NewAPI()
	var s toFalseStruct
	err := api.Unmarshal([]byte(`{"val":0}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Val {
		t.Errorf("expected false, got true")
	}
}

// ——— totrue ———

func TestToTrueDecodeFalse(t *testing.T) {
	api := NewAPI()
	var s toTrueStruct
	err := api.Unmarshal([]byte(`{"val":false}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Val {
		t.Errorf("expected false, got true")
	}
}

func TestToTrueDecodeUnknownStringDefaultsTrue(t *testing.T) {
	api := NewAPI()
	var s toTrueStruct
	err := api.Unmarshal([]byte(`{"val":"unknown"}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	if !s.Val {
		t.Errorf("expected true (default for totrue), got false")
	}
}

// ——— stringmax ———

func TestStringMaxWithinLimit(t *testing.T) {
	api := NewAPI()
	var s stringMaxStruct
	err := api.Unmarshal([]byte(`{"val":"abc"}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Val != "abc" {
		t.Errorf("expected 'abc', got '%s'", s.Val)
	}
}

func TestStringMaxExceedsLimit(t *testing.T) {
	api := NewAPI()
	var s stringMaxStruct
	err := api.Unmarshal([]byte(`{"val":"abcdefgh"}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Val != "abcde" {
		t.Errorf("expected 'abcde', got '%s'", s.Val)
	}
}

func TestStringMaxExactLimit(t *testing.T) {
	api := NewAPI()
	var s stringMaxStruct
	err := api.Unmarshal([]byte(`{"val":"abcde"}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Val != "abcde" {
		t.Errorf("expected 'abcde', got '%s'", s.Val)
	}
}

func TestStringMaxUnicode(t *testing.T) {
	api := NewAPI()
	var s stringMaxStruct
	err := api.Unmarshal([]byte(`{"val":"你好世界啊"}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Val != "你好世界啊" {
		t.Errorf("expected first 5 runes, got '%s'", s.Val)
	}
	if len([]rune(s.Val)) != 5 {
		t.Errorf("expected 5 runes, got %d", len([]rune(s.Val)))
	}
}

// ——— Combined Struct ———

func TestCombinedStructMarshal(t *testing.T) {
	api := NewAPI()
	ptrID := int64(15)
	strVal := "hello"
	s := combinedStruct{
		HexID:    255,
		PtrID:    &ptrID,
		EmptyObj: &strVal,
		EmptyArr: []int{1, 2},
		ToStr:    "hello",
		ToFalse:  true,
		ToTrue:   false,
		MaxStr:   "test",
	}
	data, err := api.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	// Verify key fields in the output
	jsonStr := string(data)
	if !bytes.Contains([]byte(jsonStr), []byte(`"hex_id":"ff"`)) {
		t.Errorf("missing hex_id:ff in %s", jsonStr)
	}
	if !bytes.Contains([]byte(jsonStr), []byte(`"empty_obj":"hello"`)) {
		t.Errorf("missing empty_obj:hello in %s", jsonStr)
	}

	// Round-trip: marshal then unmarshal
	var out combinedStruct
	err = api.Unmarshal(data, &out)
	if err != nil {
		t.Fatal(err)
	}
	if out.HexID != 255 {
		t.Errorf("HexID: expected 255, got %d", out.HexID)
	}
	if out.PtrID == nil || *out.PtrID != 15 {
		t.Errorf("PtrID: expected 15, got %v", out.PtrID)
	}
	if out.ToStr != "hello" {
		t.Errorf("ToStr: expected 'hello', got '%s'", out.ToStr)
	}
}

// ——— GetAPI ———

func TestGetAPI(t *testing.T) {
	api := GetAPI()
	if api == nil {
		t.Fatal("GetAPI returned nil")
	}
	data, err := api.Marshal(map[string]int{"a": 1})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"a":1}` {
		t.Errorf("unexpected output: %s", data)
	}
}

// ——— ginextCodec ———

func TestGinextCodecMarshal(t *testing.T) {
	codec := ginextCodec{}
	data, err := codec.Marshal(map[string]string{"key": "value"})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"key":"value"}` {
		t.Errorf("unexpected: %s", data)
	}
}

func TestGinextCodecUnmarshal(t *testing.T) {
	codec := ginextCodec{}
	var m map[string]string
	err := codec.Unmarshal([]byte(`{"key":"value"}`), &m)
	if err != nil {
		t.Fatal(err)
	}
	if m["key"] != "value" {
		t.Errorf("expected 'value', got '%s'", m["key"])
	}
}

func TestGinextCodecMarshalIndent(t *testing.T) {
	codec := ginextCodec{}
	data, err := codec.MarshalIndent(map[string]string{"key": "value"}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	expected := "{\n  \"key\": \"value\"\n}"
	if string(data) != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, data)
	}
}

func TestGinextCodecEncoder(t *testing.T) {
	codec := ginextCodec{}
	var buf bytes.Buffer
	encoder := codec.NewEncoder(&buf)
	err := encoder.Encode(map[string]string{"a": "b"})
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "{\"a\":\"b\"}\n" {
		t.Errorf("unexpected: %s", buf.String())
	}
}

func TestGinextCodecDecoder(t *testing.T) {
	codec := ginextCodec{}
	reader := bytes.NewReader([]byte(`{"key":"val"}`))
	decoder := codec.NewDecoder(reader)
	var m map[string]string
	err := decoder.Decode(&m)
	if err != nil {
		t.Fatal(err)
	}
	if m["key"] != "val" {
		t.Errorf("expected 'val', got '%s'", m["key"])
	}
}

// ——— HexStringCodec IsEmpty ———

func TestHexStringCodec_PtrIsEmpty(t *testing.T) {
	codec := &HexStringCodec{IsInt64Pointer: true}

	// nil *int64 — ptr is pointer to nil pointer
	var p *int64 = nil
	empty := codec.IsEmpty(unsafe.Pointer(&p))
	if !empty {
		t.Error("expected empty for nil *int64")
	}

	// non-nil *int64
	v := int64(5)
	p = &v
	empty = codec.IsEmpty(unsafe.Pointer(&p))
	if empty {
		t.Error("expected non-empty for non-nil *int64")
	}
}

func TestHexStringCodec_NonPtrIsEmpty(t *testing.T) {
	codec := &HexStringCodec{IsInt64Pointer: false}

	v := int64(0)
	empty := codec.IsEmpty(unsafe.Pointer(&v))
	if !empty {
		t.Error("expected empty for zero int64")
	}

	v = int64(10)
	empty = codec.IsEmpty(unsafe.Pointer(&v))
	if empty {
		t.Error("expected non-empty for non-zero int64")
	}
}

// ——— EmptyArrayInt64Codec ———

func TestEmptyArrayInt64CodecNilSlice(t *testing.T) {
	api := NewAPI()
	type testStruct struct {
		IDs []int64 `json:"ids"`
	}
	// nil slice with embedded EmptyArrayInt64Codec should encode as []
	s := testStruct{IDs: nil}
	data, err := api.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"ids":[]}` {
		t.Errorf("expected {\"ids\":[]} for nil slice, got %s", data)
	}
}

func TestEmptyArrayInt64CodecHexEncoding(t *testing.T) {
	api := NewAPI()
	type testStruct struct {
		IDs []int64 `json:"ids"`
	}
	s := testStruct{IDs: []int64{255, 0, 16}}
	data, err := api.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"ids":["ff","0","10"]}` {
		t.Errorf("expected {\"ids\":[\"ff\",\"0\",\"10\"]}, got %s", data)
	}
}

func TestEmptyArrayInt64CodecDecode(t *testing.T) {
	api := NewAPI()
	type testStruct struct {
		IDs []int64 `json:"ids"`
	}
	var s testStruct
	err := api.Unmarshal([]byte(`{"ids":["a","10","ff"]}`), &s)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.IDs) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(s.IDs))
	}
	if s.IDs[0] != 10 {
		t.Errorf("expected 10, got %d", s.IDs[0])
	}
	if s.IDs[1] != 16 {
		t.Errorf("expected 16, got %d", s.IDs[1])
	}
	if s.IDs[2] != 255 {
		t.Errorf("expected 255, got %d", s.IDs[2])
	}
}


