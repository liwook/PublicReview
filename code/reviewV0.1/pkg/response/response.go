package response

import (
	"errors"
	"log/slog"
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

// HandleBusinessError 统一处理业务错误
func HandleBusinessError(c *gin.Context, err error) {
	var bizErr *BusinessError
	if errors.As(err, &bizErr) {
		//记录业务错误日志
		if bizErr.Err != nil {
			slog.Error("business error", "originalErr", bizErr.Err, "code", bizErr.Code, "message", bizErr.Message)
		}
		Error(c, bizErr.Code, bizErr.Message)
	} else {
		slog.Error("unknown error", "err", err)
		Error(c, ErrUnknown, "")
	}
}

// HandleBusinessErrorWithData 统一处理业务错误（可附带数据）
func HandleBusinessErrorWithData(c *gin.Context, err error, data any) {
	var bizErr *BusinessError
	if errors.As(err, &bizErr) {
		if bizErr.Err != nil {
			slog.Error("business error with data", "originalErr", bizErr.Err, "code", bizErr.Code, "message", bizErr.Message)
		}
		ErrorWithData(c, bizErr.Code, bizErr.Message, data)
	} else {
		slog.Error("unknown error with data", "err", err)
		ErrorWithData(c, ErrUnknown, "", data)
	}
}

// HandleBusinessResult 统一处理业务结果（错误或成功）
func HandleBusinessResult(c *gin.Context, err error, data any) {
	if err != nil {
		HandleBusinessErrorWithData(c, err, data)
	} else {
		Success(c, data)
	}
}
