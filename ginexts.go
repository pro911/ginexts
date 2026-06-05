// Package ginexts provides unified gin extensions.
//
// Import this package to activate all custom gin features with a single blank import:
//
//	import _ "github.com/pro911/ginexts"
//
// This automatically:
//   - Overrides gin's JSON codec with jsoniter + custom tag extensions (hexstring, emptyobject, etc.)
//   - Wraps gin's Form/Query/FormPost/FormMultipart bindings for hexstring form tag support
//
// Sub-packages:
//   - ginexts/json: JSON codec override + programmatic API access
//   - ginexts/form: form/query hexstring binding
//
// For programmatic JSON operations (outside gin), use ginexts/json.GetAPI():
//
//	import ginextjson "github.com/pro911/ginexts/json"
//	api := ginextjson.GetAPI()
//	api.Unmarshal(data, &obj)
package ginexts

import (
	_ "github.com/pro911/ginexts/form" // register hexstring form bindings
	_ "github.com/pro911/ginexts/json" // register JSON codec override
)
