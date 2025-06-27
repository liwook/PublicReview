package response

import (
	"net/http"
)

// ErrorMeta 错误元信息（包含HTTP状态码和默认消息）
type ErrorMeta struct {
	HTTP    int    `json:"-"`       // HTTP状态码（不返回给客户端）,假如直接返回给客户端使用的话，用了“-”符号，则不会在body中的结构体中返回HTTP状态码,
	Message string `json:"message"` // 默认错误消息
}

func (coder ErrorMeta) Error() string {
	return coder.Message
}

func (coder ErrorMeta) HTTPStatus() int {
	if coder.HTTP == 0 {
		return 500
	}
	return coder.HTTP
}

func register(httpStatus int, message string) *ErrorMeta {
	if _, ok := allowedHTTPStatus[httpStatus]; !ok {
		panic("httstatuscode and code  are not good")
	}

	return &ErrorMeta{
		HTTP:    httpStatus,
		Message: message,
	}
}

var codes = map[int]*ErrorMeta{
	ErrBind:              register(http.StatusBadRequest, "请求参数格式错误"),
	ErrSuccess:           register(http.StatusOK, "SUCCESS"),
	ErrUnknown:           register(http.StatusInternalServerError, "内部服务器错误"),
	ErrValidation:        register(http.StatusBadRequest, "验证失败"),
	ErrNotFound:          register(http.StatusNotFound, "资源不存在"),
	ErrTokenInvalid:      register(http.StatusUnauthorized, "token无效"),
	ErrDatabase:          register(http.StatusInternalServerError, "数据库错误"),
	ErrEncrypt:           register(http.StatusUnauthorized, "用户密码加密时发生错误"),
	ErrSignatureInvalid:  register(http.StatusUnauthorized, "签名无效"),
	ErrExpired:           register(http.StatusUnauthorized, "token过期"),
	ErrInvalidAuthHeader: register(http.StatusUnauthorized, "无效的授权头"),
	ErrMissingHeader:     register(http.StatusUnauthorized, "授权头为空"),
	ErrPasswordIncorrect: register(http.StatusUnauthorized, "密码错误"),
	ErrPermissionDenied:  register(http.StatusForbidden, "权限不足"),
	ErrEncodingFailed:    register(http.StatusInternalServerError, "由于数据错误导致编码失败"),
	ErrDecodingFailed:    register(http.StatusInternalServerError, "由于数据错误导致解码失败"),
	ErrInvalidJSON:       register(http.StatusInternalServerError, "数据不是有效的JSON"),
	ErrEncodingJSON:      register(http.StatusInternalServerError, "JSON数据无法编码"),
	ErrDecodingJSON:      register(http.StatusInternalServerError, "JSON数据无法解码"),
	ErrInvalidYaml:       register(http.StatusInternalServerError, "数据不是有效的YAML"),
	ErrEncodingYaml:      register(http.StatusInternalServerError, "YAML数据无法编码"),
	ErrDecodingYaml:      register(http.StatusInternalServerError, "YAML数据无法解码"),
}
