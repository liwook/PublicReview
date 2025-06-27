package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	BizCode int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"` // 返回数据，omitempty使其在为空时不输出
}

// WriteResponse used to write an error and JSON data into response.
func writeResponse(c *gin.Context, bizCode int, message string, data any) {
	coder, ok := codes[bizCode]
	if !ok {
		coder = codes[ErrUnknown]
	}

	if message != "" {
		coder.Message = message
	}
	if coder.HTTPStatus() != http.StatusOK {
		c.AbortWithStatusJSON(coder.HTTPStatus(), Response{
			BizCode: bizCode,
			Message: coder.Error(),
			Data:    data,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		BizCode: bizCode,
		Message: coder.Error(),
		Data:    data})
}

func Success(c *gin.Context, data any) {
	writeResponse(c, ErrSuccess, "", data)
}

func SuccesswithMsg(c *gin.Context, message string, data any) {
	writeResponse(c, ErrSuccess, message, data)
}

func Error(c *gin.Context, bizCode int, message string) {
	writeResponse(c, bizCode, message, nil)
}

func ErrorWithData(c *gin.Context, bizCode int, message string, data any) {
	writeResponse(c, bizCode, message, data)
}
