package json

import (
	"io"

	ginjsoncodec "github.com/gin-gonic/gin/codec/json"
	jsoniter "github.com/json-iterator/go"
)

// API is the shared jsoniter instance with all extensions registered.
// It's used internally by gin's JSON codec and exposed for programmatic use.
var API jsoniter.API

// GetAPI returns the shared jsoniter API with all extensions.
func GetAPI() jsoniter.API {
	return API
}

func init() {
	API = NewAPI()
	ginjsoncodec.API = ginextCodec{}
}

// ginextCodec implements gin's codec/json.Core interface.
type ginextCodec struct{}

func (ginextCodec) Marshal(v any) ([]byte, error)      { return API.Marshal(v) }
func (ginextCodec) Unmarshal(data []byte, v any) error { return API.Unmarshal(data, v) }
func (ginextCodec) MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return API.MarshalIndent(v, prefix, indent)
}
func (ginextCodec) NewEncoder(w io.Writer) ginjsoncodec.Encoder { return API.NewEncoder(w) }
func (ginextCodec) NewDecoder(r io.Reader) ginjsoncodec.Decoder { return API.NewDecoder(r) }
