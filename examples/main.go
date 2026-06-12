// Package main 演示 ginexts 所有扩展功能的示例程序。
//
// 启动方式：
//
//	go run examples/main.go
//
// 访问 http://localhost:8080 查看 API 列表。
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	// 一行空白导入，激活所有 ginexts 功能
	_ "github.com/pro911/ginexts"

	// 导入 json 子包，用于程序化操作
	ginextjson "github.com/pro911/ginexts/json"
)

// ==========================================================================
// 演示用的数据结构
// ==========================================================================

// HexDemo 演示 hexstring 标签：int64 / *int64 / []int64 与十六进制字符串互转。
type HexDemo struct {
	PlainID  int64   `json:"plain_id,hexstring"`
	PtrID    *int64  `json:"ptr_id,hexstring"`
	SliceIDs []int64 `json:"slice_ids,hexstring"`
}

// EmptyDemo 演示 emptyobject / emptyarray：nil 值输出 {} 或 []。
type EmptyDemo struct {
	Obj    *string `json:"obj,emptyobject"`
	Arr    []int   `json:"arr,emptyarray"`
	Normal string  `json:"normal"`
}

// ConvertDemo 演示 tostring / tofalse / totrue：值类型自动转换。
type ConvertDemo struct {
	AsStr    string `json:"as_str,tostring"`
	AsFalse  bool   `json:"as_false,tofalse"`
	AsTrue   bool   `json:"as_true,totrue"`
	ShortStr string `json:"short_str,stringmax=5"`
}

// AllInOne 综合演示所有标签。
type AllInOne struct {
	ID       int64   `json:"id,hexstring"`
	ParentID *int64  `json:"parent_id,hexstring"`
	TagIDs   []int64 `json:"tag_ids"` // []int64 自动 hex 编码 + nil→[]
	Meta     *string `json:"meta,emptyobject"`
	Items    []int   `json:"items,emptyarray"`
	Raw      string  `json:"raw,tostring"`
	Flag     bool    `json:"flag,tofalse"`
	Title    string  `json:"title,stringmax=10"`
}

// FormDemo 演示 form 标签的 hexstring 支持。
type FormDemo struct {
	ProjectID int64   `form:"project_id,hexstring"`
	OwnerID   *int64  `form:"owner_id,hexstring"`
	GroupIDs  []int64 `form:"group_ids,hexstring"`
	Name      string  `form:"name"`
}

// ==========================================================================
// 路由处理器
// ==========================================================================

// POST /api/hex/encode — 展示 hexstring 编码
func handleHexEncode(c *gin.Context) {
	v := int64(255)
	c.JSON(http.StatusOK, HexDemo{
		PlainID:  255,
		PtrID:    &v,
		SliceIDs: []int64{255, 0, 16},
	})
	// 输出: {"plain_id":"ff","ptr_id":"ff","slice_ids":["ff","0","10"]}
}

// POST /api/hex/decode — 展示 hexstring 解码
func handleHexDecode(c *gin.Context) {
	var obj HexDemo
	// 接收十六进制字符串，自动转为十进制 int64
	if err := c.ShouldBindJSON(&obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"plain_id":  obj.PlainID,
		"ptr_id":    obj.PtrID,
		"slice_ids": obj.SliceIDs,
	})
}

// POST /api/empty — 展示 emptyobject / emptyarray
func handleEmpty(c *gin.Context) {
	c.JSON(http.StatusOK, EmptyDemo{
		Obj:    nil, // → {}
		Arr:    nil, // → []
		Normal: "hello",
	})
}

// POST /api/convert — 展示 tostring / tofalse / totrue / stringmax
func handleConvert(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"note": "发送任意 JSON 值到 /api/convert/decode 查看转换效果",
	})
}

// POST /api/convert/decode — 演示解码时的类型转换
func handleConvertDecode(c *gin.Context) {
	var obj ConvertDemo
	if err := c.ShouldBindJSON(&obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"as_str":    obj.AsStr,
		"as_false":  obj.AsFalse,
		"as_true":   obj.AsTrue,
		"short_str": obj.ShortStr,
	})
}

// POST /api/all-in-one — 综合演示
func handleAllInOne(c *gin.Context) {
	parentID := int64(0xABCD)
	meta := "用户备注"
	c.JSON(http.StatusOK, AllInOne{
		ID:       255,
		ParentID: &parentID,
		TagIDs:   []int64{10, 20, 30},
		Meta:     &meta,
		Items:    []int{1, 2, 3},
		Raw:      "not-converted",
		Flag:     true,
		Title:    "一个很长的标题会被截断",
	})
}

// GET /api/form?project_id=ff&owner_id=a&group_ids=1,2,3&name=test — 演示 form hexstring
func handleFormBind(c *gin.Context) {
	var obj FormDemo
	if err := c.ShouldBindQuery(&obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"project_id": obj.ProjectID,
		"owner_id":   obj.OwnerID,
		"group_ids":  obj.GroupIDs,
		"name":       obj.Name,
	})
}

// GET /api/programmatic — 演示脱离 Gin 的程序化 JSON 操作
func handleProgrammatic(c *gin.Context) {
	api := ginextjson.GetAPI()

	type Item struct {
		ID int64 `json:"id,hexstring"`
	}

	// 编码
	encoded, _ := api.Marshal(Item{ID: 255})
	// {"id":"ff"}

	// 解码
	var decoded Item
	_ = api.Unmarshal([]byte(`{"id":"ff"}`), &decoded)
	// decoded.ID == 255

	c.JSON(http.StatusOK, gin.H{
		"marshal_result":   string(encoded),
		"unmarshal_result": decoded.ID,
		"note":             "See console for more details",
	})

	// 同时在控制台输出更多例子
	fmt.Println("=== 程序化 API 演示 ===")
	fmt.Printf("Marshal(Item{ID:255}) = %s\n", encoded)
	fmt.Printf("Unmarshal(`{\"id\":\"ff\"}`) => ID=%d\n", decoded.ID)

	// 编码一个复合结构
	type Complex struct {
		HexID    int64   `json:"hex_id,hexstring"`
		EmptyObj *int    `json:"empty_obj,emptyobject"`
		MaxStr   string  `json:"max_str,stringmax=4"`
	}
	data, _ := api.Marshal(Complex{HexID: 4095, EmptyObj: nil, MaxStr: "hello world"})
	fmt.Printf("Marshal(Complex) = %s\n", data)

	// 解码一个复合结构
	var c2 Complex
	_ = api.Unmarshal([]byte(`{"hex_id":"fff","empty_obj":null,"max_str":"简短文字"}`), &c2)
	fmt.Printf("Unmarshal => HexID=%d, EmptyObj=%v, MaxStr=%q\n", c2.HexID, c2.EmptyObj, c2.MaxStr)
}

// ==========================================================================
// 主函数
// ==========================================================================

func main() {
	r := gin.Default()

	// 首页：API 导航
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "ginexts 示例服务",
			"endpoints": []string{
				"POST /api/hex/encode        — hexstring 编码演示",
				"POST /api/hex/decode        — hexstring 解码演示",
				"POST /api/empty             — emptyobject / emptyarray 演示",
				"POST /api/convert           — 类型转换说明",
				"POST /api/convert/decode    — tostring/tofalse/totrue/stringmax 解码演示",
				"POST /api/all-in-one        — 综合演示（所有标签）",
				"GET  /api/form?project_id=ff&owner_id=a&group_ids=1,2,3&name=test",
				"                             — form hexstring 演示",
				"GET  /api/programmatic       — 程序化 API 演示",
			},
		})
	})

	api := r.Group("/api")
	{
		api.POST("/hex/encode", handleHexEncode)
		api.POST("/hex/decode", handleHexDecode)
		api.POST("/empty", handleEmpty)
		api.POST("/convert", handleConvert)
		api.POST("/convert/decode", handleConvertDecode)
		api.POST("/all-in-one", handleAllInOne)
		api.GET("/form", handleFormBind)
		api.GET("/programmatic", handleProgrammatic)
	}

	log.Println("ginexts 示例服务启动于 http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
