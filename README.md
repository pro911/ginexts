# ginexts — Gin Framework Extensions

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.22-blue)](https://go.dev/)
[![Gin Version](https://img.shields.io/badge/Gin-v1.12.0-green)](https://github.com/gin-gonic/gin)

**ginexts** is a non-invasive extension pack for the [Gin web framework](https://github.com/gin-gonic/gin) that adds custom tag support for JSON serialization and Form/Query binding.

## ✨ Features

### JSON Tag Extensions

| Tag Option      | Description                                        | Example                          |
|-----------------|----------------------------------------------------|----------------------------------|
| `hexstring`     | int64 ↔ hex string (supports pointer & slice)      | `json:"id,hexstring"`            |
| `emptyobject`   | nil → `{}`                                         | `json:"data,emptyobject"`        |
| `emptyarray`    | nil slice → `[]`                                   | `json:"tags,emptyarray"`         |
| `tostring`      | Any JSON value → string on decode                  | `json:"val,tostring"`            |
| `stringmax=N`   | Truncate string to N runes on decode               | `json:"name,stringmax=100"`      |
| `tofalse`       | Any JSON value → bool (default: false)             | `json:"admin,tofalse"`           |
| `totrue`        | Any JSON value → bool (default: true)              | `json:"admin,totrue"`            |

### Form/Query Binding Extension

| Tag Option      | Description                                        | Example                          |
|-----------------|----------------------------------------------------|----------------------------------|
| `hexstring`     | Hex string → decimal on form binding               | `form:"id,hexstring"`            |

## 📦 Installation

```bash
go get github.com/pro911/ginexts
```

## 🚀 Quick Start

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
    _ "github.com/pro911/ginexts" // single import activates everything
)

type User struct {
    // JSON: encode int64 as hex, decode hex string back to int64
    ProjectID int64  `json:"project_id,hexstring" form:"project_id,hexstring"`
    // Form: hex string "0xFF" → int64 255 on binding
    UserID    *int64 `json:"user_id,hexstring" form:"user_id,hexstring"`
    Tags      []int64 `json:"tags,hexstring" form:"tags,hexstring"`
    // nil → {}
    Metadata  any    `json:"metadata,emptyobject"`
    // nil → []
    Items     []string `json:"items,emptyarray"`
}

func main() {
    r := gin.Default()
    r.GET("/user", func(c *gin.Context) {
        var u User
        // form binding automatically converts hex strings
        if err := c.ShouldBindQuery(&u); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        // JSON response automatically encodes int64 as hex
        c.JSON(200, u)
    })
    r.Run()
}
```

## 📁 Package Structure

```
ginexts/
├── ginexts.go          # Top-level entry (blank import activates all)
├── json/              # JSON codec overrides
│   ├── codec.go       # Replaces gin's JSON codec
│   ├── encoder.go     # All encoder/decoder implementations
│   └── extension.go   # jsoniter extension registration
└── form/              # Form/Query binding extensions
    └── form.go        # Hex-aware form binding wrapper
```

## 🔧 Programmatic JSON API

For custom JSON operations outside gin's HTTP handlers:

```go
import ginextjson "github.com/pro911/ginexts/json"

api := ginextjson.GetAPI()
data, _ := api.Marshal(myStruct)
```

## 🧠 How It Works

- **JSON**: Overrides gin's `codec/json.API` with a jsoniter instance that has custom `UpdateStructDescriptor` extensions. Zero modification to gin source.
- **Form**: Wraps gin's `binding.Form/Query/FormPost/FormMultipart` (which are `Binding` interface variables) with hex-aware preprocessors. Completely non-invasive.

## 📄 License

MIT
