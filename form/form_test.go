package form

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// ——— head ———

func TestHead(t *testing.T) {
	h, tail := head("name,hexstring", ",")
	if h != "name" {
		t.Errorf("expected head 'name', got '%s'", h)
	}
	if tail != "hexstring" {
		t.Errorf("expected tail 'hexstring', got '%s'", tail)
	}
}

func TestHeadNoSeparator(t *testing.T) {
	h, tail := head("name", ",")
	if h != "name" {
		t.Errorf("expected head 'name', got '%s'", h)
	}
	if tail != "" {
		t.Errorf("expected empty tail, got '%s'", tail)
	}
}

func TestHeadEmpty(t *testing.T) {
	h, tail := head("", ",")
	if h != "" {
		t.Errorf("expected empty head, got '%s'", h)
	}
	if tail != "" {
		t.Errorf("expected empty tail, got '%s'", tail)
	}
}

func TestHeadMultipleSep(t *testing.T) {
	h, tail := head("a,b,c", ",")
	if h != "a" {
		t.Errorf("expected head 'a', got '%s'", h)
	}
	if tail != "b,c" {
		t.Errorf("expected tail 'b,c', got '%s'", tail)
	}
}

// ——— hasHexstring ———

func TestHasHexstring(t *testing.T) {
	if !hasHexstring("hexstring") {
		t.Error("expected true for 'hexstring'")
	}
	if !hasHexstring("name,hexstring") {
		t.Error("expected true for 'name,hexstring'")
	}
	if !hasHexstring("name,omitempty,hexstring") {
		t.Error("expected true for 'name,omitempty,hexstring'")
	}
	if !hasHexstring("hexstring,omitempty") {
		t.Error("expected true for 'hexstring,omitempty'")
	}
}

func TestHasHexstringFalse(t *testing.T) {
	if hasHexstring("") {
		t.Error("expected false for empty string")
	}
	if hasHexstring("name") {
		t.Error("expected false for 'name'")
	}
	if hasHexstring("name,omitempty") {
		t.Error("expected false for 'name,omitempty'")
	}
}

// ——— isIntKind ———

func TestIsIntKind(t *testing.T) {
	kinds := []struct {
		kind     string
		expected bool
	}{
		{"int", true},
		{"int8", true},
		{"int16", true},
		{"int32", true},
		{"int64", true},
		{"uint", false},
		{"string", false},
		{"bool", false},
	}
	for _, k := range kinds {
		t.Run(k.kind, func(t *testing.T) {
			result := false
			switch k.kind {
			case "int":
				result = isIntKind(2) // reflect.Int
			case "int8":
				result = isIntKind(3) // reflect.Int8
			case "int16":
				result = isIntKind(4) // reflect.Int16
			case "int32":
				result = isIntKind(5) // reflect.Int32
			case "int64":
				result = isIntKind(6) // reflect.Int64
			case "uint":
				result = isIntKind(7) // reflect.Uint
			case "string":
				result = isIntKind(24) // reflect.String
			case "bool":
				result = isIntKind(1) // reflect.Bool
			}
			if result != k.expected {
				t.Errorf("isIntKind(%s)=%v, expected %v", k.kind, result, k.expected)
			}
		})
	}
}

// ——— isUintKind ———

func TestIsUintKind(t *testing.T) {
	kinds := []struct {
		kind     string
		expected bool
	}{
		{"uint", true},
		{"uint8", true},
		{"uint16", true},
		{"uint32", true},
		{"uint64", true},
		{"int64", false},
		{"string", false},
	}
	for _, k := range kinds {
		t.Run(k.kind, func(t *testing.T) {
			result := false
			switch k.kind {
			case "uint":
				result = isUintKind(7) // reflect.Uint
			case "uint8":
				result = isUintKind(8) // reflect.Uint8
			case "uint16":
				result = isUintKind(9) // reflect.Uint16
			case "uint32":
				result = isUintKind(10) // reflect.Uint32
			case "uint64":
				result = isUintKind(11) // reflect.Uint64
			case "int64":
				result = isUintKind(6) // reflect.Int64
			case "string":
				result = isUintKind(24) // reflect.String
			}
			if result != k.expected {
				t.Errorf("isUintKind(%s)=%v, expected %v", k.kind, result, k.expected)
			}
		})
	}
}

// ——— hexToDec ———

func TestHexToDec(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"0", "0"},
		{"1", "1"},
		{"a", "10"},
		{"ff", "255"},
		{"0xff", "255"},
		{"0XFF", "255"},
		{"10", "16"},
		{"nothex", "nothex"},
	}
	for _, tc := range tests {
		actual := hexToDec(tc.input)
		if actual != tc.expected {
			t.Errorf("hexToDec(%q) = %q, expected %q", tc.input, actual, tc.expected)
		}
	}
}

func TestHexToDecLargeHex(t *testing.T) {
	// 0x7fffffffffffffff = 9223372036854775807 (max int64)
	result := hexToDec("7fffffffffffffff")
	if result == "7fffffffffffffff" {
		t.Error("expected conversion of large hex")
	}
}

// ——— preprocess / walkStruct ———

type testFormStruct struct {
	ProjectID int64  `form:"project_id,hexstring"`
	UserID    *int64 `form:"user_id,hexstring"`
	GroupIDs  []int64 `form:"group_ids,hexstring"`
	Name      string `form:"name"`
	Count     int    `form:"count"`
}

type embeddedFormStruct struct {
	testFormStruct
	Extra int `form:"extra"`
}

type noHexStruct struct {
	Name  string `form:"name"`
	Value int    `form:"value"`
}

func TestPreprocessHexInt64(t *testing.T) {
	obj := &testFormStruct{}
	form := map[string][]string{
		"project_id": {"ff"},
	}
	preprocess(obj, form, "form")

	if form["project_id"][0] != "255" {
		t.Errorf("expected '255', got '%s'", form["project_id"][0])
	}
}

func TestPreprocessHexPtrInt64(t *testing.T) {
	obj := &testFormStruct{}
	form := map[string][]string{
		"user_id": {"a"},
	}
	preprocess(obj, form, "form")

	if form["user_id"][0] != "10" {
		t.Errorf("expected '10', got '%s'", form["user_id"][0])
	}
}

func TestPreprocessHexSliceInt64(t *testing.T) {
	obj := &testFormStruct{}
	form := map[string][]string{
		"group_ids": {"a", "ff", "10"},
	}
	preprocess(obj, form, "form")

	if len(form["group_ids"]) != 3 {
		t.Fatalf("expected 3 values, got %d", len(form["group_ids"]))
	}
	if form["group_ids"][0] != "10" {
		t.Errorf("expected '10', got '%s'", form["group_ids"][0])
	}
	if form["group_ids"][1] != "255" {
		t.Errorf("expected '255', got '%s'", form["group_ids"][1])
	}
	if form["group_ids"][2] != "16" {
		t.Errorf("expected '16', got '%s'", form["group_ids"][2])
	}
}

func TestPreprocessNonHexField(t *testing.T) {
	obj := &testFormStruct{}
	form := map[string][]string{
		"name":  {"hello"},
		"count": {"10"},
	}
	preprocess(obj, form, "form")

	if form["name"][0] != "hello" {
		t.Errorf("expected 'hello', got '%s'", form["name"][0])
	}
	if form["count"][0] != "10" {
		t.Errorf("expected '10', got '%s'", form["count"][0])
	}
}

func TestPreprocessEmptyForm(t *testing.T) {
	obj := &testFormStruct{}
	form := map[string][]string{}
	preprocess(obj, form, "form")
	// should not panic
}

func TestPreprocessNilObj(t *testing.T) {
	form := map[string][]string{"project_id": {"ff"}}
	preprocess(nil, form, "form")
	// should not panic
}

func TestPreprocessNilForm(t *testing.T) {
	obj := &testFormStruct{}
	preprocess(obj, nil, "form")
	// should not panic
}

func TestPreprocessNonStruct(t *testing.T) {
	var obj int = 42
	form := map[string][]string{"x": {"ff"}}
	preprocess(&obj, form, "form")
	// should not panic, just skip
}

func TestPreprocessEmbeddedStruct(t *testing.T) {
	obj := &embeddedFormStruct{}
	form := map[string][]string{
		"project_id": {"ff"},
		"extra":      {"42"},
	}
	preprocess(obj, form, "form")

	if form["project_id"][0] != "255" {
		t.Errorf("expected '255', got '%s'", form["project_id"][0])
	}
	if form["extra"][0] != "42" {
		t.Errorf("expected '42', got '%s'", form["extra"][0])
	}
}

func TestPreprocessNoHexFields(t *testing.T) {
	obj := &noHexStruct{}
	form := map[string][]string{
		"name":  {"test"},
		"value": {"123"},
	}
	preprocess(obj, form, "form")

	if form["name"][0] != "test" {
		t.Errorf("expected 'test', got '%s'", form["name"][0])
	}
	if form["value"][0] != "123" {
		t.Errorf("expected '123', got '%s'", form["value"][0])
	}
}

func TestPreprocess0xPrefix(t *testing.T) {
	obj := &testFormStruct{}
	form := map[string][]string{
		"project_id": {"0xff"},
	}
	preprocess(obj, form, "form")

	if form["project_id"][0] != "255" {
		t.Errorf("expected '255', got '%s'", form["project_id"][0])
	}
}

func TestPreprocess0XUppercase(t *testing.T) {
	obj := &testFormStruct{}
	form := map[string][]string{
		"project_id": {"0XFF"},
	}
	preprocess(obj, form, "form")

	if form["project_id"][0] != "255" {
		t.Errorf("expected '255', got '%s'", form["project_id"][0])
	}
}

func TestPreprocessMultipleValues(t *testing.T) {
	obj := &testFormStruct{}
	form := map[string][]string{
		"project_id": {"ff", "a"},
	}
	preprocess(obj, form, "form")

	if len(form["project_id"]) != 2 {
		t.Fatalf("expected 2 values, got %d", len(form["project_id"]))
	}
	if form["project_id"][0] != "255" {
		t.Errorf("expected '255', got '%s'", form["project_id"][0])
	}
	if form["project_id"][1] != "10" {
		t.Errorf("expected '10', got '%s'", form["project_id"][1])
	}
}

// ——— Form tag with omitempty ———

type formWithOmitempty struct {
	ID int64 `form:"id,omitempty,hexstring"`
}

func TestPreprocessWithOmitempty(t *testing.T) {
	obj := &formWithOmitempty{}
	form := map[string][]string{
		"id": {"ff"},
	}
	preprocess(obj, form, "form")

	if form["id"][0] != "255" {
		t.Errorf("expected '255', got '%s'", form["id"][0])
	}
}

// ——— hexBinding.Name ———

func TestHexBindingName(t *testing.T) {
	h := &hexBinding{name: "form-hex"}
	if h.Name() != "form-hex" {
		t.Errorf("expected 'form-hex', got '%s'", h.Name())
	}
}

// ——— hexBinding.Bind ———

func TestHexBindingBindForm(t *testing.T) {
	h := &hexBinding{name: "form-hex", delegate: &mockBinding{}, tag: "form"}

	formData := url.Values{
		"project_id": {"ff"},
		"user_id":    {"a"},
	}
	req, err := http.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	obj := &testFormStruct{}
	err = h.Bind(req, obj)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHexBindingBindQuery(t *testing.T) {
	h := &hexBinding{name: "query-hex", delegate: &mockBinding{}, tag: "form"}

	req, err := http.NewRequest("GET", "/test?project_id=ff&user_id=a", nil)
	if err != nil {
		t.Fatal(err)
	}

	obj := &testFormStruct{}
	err = h.Bind(req, obj)
	if err != nil {
		t.Fatal(err)
	}
}

// ——— mockBinding ———

type mockBinding struct{}

func (m *mockBinding) Name() string                           { return "mock" }
func (m *mockBinding) Bind(req *http.Request, obj any) error { return nil }
