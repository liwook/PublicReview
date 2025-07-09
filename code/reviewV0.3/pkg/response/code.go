package response

import (
	"fmt"
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

// Register 对外开放的错误码注册函数
// 用于允许其他包注册自定义错误码
func Register(code int, httpStatus int, message string) error {
	// 验证错误码是否已存在
	if _, exists := codes[code]; exists {
		return fmt.Errorf("错误码 %d 已存在", code)
	}

	// 验证HTTP状态码是否允许
	if _, ok := allowedHTTPStatus[httpStatus]; !ok {
		return fmt.Errorf("不支持的HTTP状态码: %d", httpStatus)
	}

	// 可以添加错误码范围限制，比如用户自定义错误码必须在特定范围内
	if code < 10000 {
		return fmt.Errorf("自定义错误码必须大于等于10000")
	}

	codes[code] = &ErrorMeta{
		HTTP:    httpStatus,
		Message: message,
	}

	return nil
}

// 为了让外部包能够使用注册的错误码，提供查询功能。
func GetErrorMeta(code int) (*ErrorMeta, bool) {
	meta, exists := codes[code]
	if !exists {
		return nil, false
	}
	// 返回副本，防止外部修改
	return &ErrorMeta{HTTP: meta.HTTP, Message: meta.Message}, true
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
	ErrPasswordIncorrect: register(http.StatusUnauthorized, "密码错误"), // 用于已认证用户的密码验证场景
	ErrPermissionDenied:  register(http.StatusForbidden, "权限不足"),
	ErrEncodingFailed:    register(http.StatusInternalServerError, "由于数据错误导致编码失败"),
	ErrDecodingFailed:    register(http.StatusInternalServerError, "由于数据错误导致解码失败"),
	ErrInvalidJSON:       register(http.StatusInternalServerError, "数据不是有效的JSON"),
	ErrEncodingJSON:      register(http.StatusInternalServerError, "JSON数据无法编码"),
	ErrDecodingJSON:      register(http.StatusInternalServerError, "JSON数据无法解码"),
	ErrInvalidYaml:       register(http.StatusInternalServerError, "数据不是有效的YAML"),
	ErrEncodingYaml:      register(http.StatusInternalServerError, "YAML数据无法编码"),
	ErrDecodingYaml:      register(http.StatusInternalServerError, "YAML数据无法解码"),
	ErrLoginFailed:       register(http.StatusUnauthorized, "用户名或密码错误"),
}

type BusinessError struct {
	Code    int
	Message string
	Err     error `json:"-"` //原始错误
}

// 实现了Error()string，就是了error类型
func (e *BusinessError) Error() string {
	return e.Message
}

func (e *BusinessError) Unwrap() error {
	return e.Err
}

func NewBusinessError(code int, customMessage string) *BusinessError {
	if meta, exists := codes[code]; exists {
		message := customMessage
		if message == "" {
			message = meta.Message
		}
		return &BusinessError{
			Code:    code,
			Message: message,
		}
	}
	return &BusinessError{
		Code:    ErrUnknown,
		Message: "未知错误",
	}
}

// WrapBusinessError 包装原始错误
func WrapBusinessError(code int, originalErr error, customMessage string) *BusinessError {
	bizErr := NewBusinessError(code, customMessage)
	bizErr.Err = originalErr
	return bizErr
}
