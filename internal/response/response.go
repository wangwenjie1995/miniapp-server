// response 包提供公共 JSON 输出入口，避免各业务模块重复设置响应格式。
package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// JSON 将 body 序列化后写入响应，并设置状态码和 JSON Content-Type。
// any 表示可传入不同类型，例如结构体、map 或切片；当前不额外包装响应体。
func JSON(c *gin.Context, status int, body any) {
	c.JSON(status, body)
}

// Error 返回统一错误结构：code 是业务错误标识，message 是给调用方阅读的说明。
// status 是 HTTP 状态码，例如 500；业务错误码和 HTTP 状态码是两个不同字段。
func Error(c *gin.Context, status int, code, message string) {
	// 同时阻止后续处理函数并写出 JSON。该调用不会自动 return 当前函数；
	// 调用方如需停止自身后续代码，仍要使用 return。
	c.AbortWithStatusJSON(status, gin.H{"code": code, "message": message})
}

// Success 是业务接口的成功响应，对齐小程序请求层的 code/message/data 约定。
// code=0 表示成功，data 放实际结果；前端统一请求层会自动解包 data。
func Success(c *gin.Context, data any) {
	JSON(c, http.StatusOK, gin.H{"code": 0, "message": "success", "data": data})
}
