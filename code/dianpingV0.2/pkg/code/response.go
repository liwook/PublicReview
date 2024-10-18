package code

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response defines project response format
// 使用json标签中的omitempty选项来实现当字段为空值时不返回该字段
type Response struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// WriteResponse used to write an error and JSON data into response.
func WriteResponse(c *gin.Context, code int, data interface{}) {

	coder := ParseCoder(code)
	if coder.HTTPStatus() != http.StatusOK {
		c.JSON(coder.HTTPStatus(), Response{
			Code:    coder.Code(),
			Message: coder.String(),
			Data:    data,
		})
		return
	}

	c.JSON(http.StatusOK, Response{Data: data})
}
