package code

var OnlyUseHTTPStatus = map[int]bool{200: true, 400: true, 401: true, 403: true, 404: true, 500: true}

// http状态码 5开头表示服务器端错误。4开头表示客户端错误
// 400 Bad Request（错误请求）,401 Unauthorized（未授权）,403 Forbidden（禁止访问）

// 基础错误
// code must start with 1xxxxx
const (
	ErrSuccess int = iota + 100001
	ErrUnknown
	ErrBind              //将 HTTP 请求中的数据绑定到特定的数据结构（如结构体）的过程
	ErrValidation        //validation failed,	对已经绑定或接收到的数据进行进一步验证的阶段，
	ErrTokenInvalid      //token invalid
	ErrHttpHeaderInvalid //http header invalid
)

// 数据库类错误
const (
	ErrDatabase int = iota + 100101
)

// 认证授权类错误
const (
	ErrEncrypt int = iota + 100201
	ErrSignatureInvalid
	ErrExpired
	ErrInvalidAuthHeader
	ErrMissingHeader //The `Authorization` header was empty.
	ErrPasswordIncorrect
	ErrPermissionDenied //Permission denied.
	ErrTokenGenerationFailed
)

// 编解码类错误
const (
	// ErrEncodingFailed - 500: Encoding failed due to an error with the data.
	ErrEncodingFailed int = iota + 100301
	ErrDecodingFailed
	ErrInvalidJSON
	ErrEncodingJSON
	ErrDecodingJSON
	// ErrInvalidYaml - 500: Data is not valid Yaml.
	ErrInvalidYaml
	ErrEncodingYaml
	ErrDecodingYaml
)
