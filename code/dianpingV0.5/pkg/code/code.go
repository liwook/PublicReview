package code

import (
	"net/http"
	"sync"
)

// error是个接口，只要实现了Error()string方法，就是error类型了
//
//	type error interface {
//		Error() string
//	}
type Errcode struct {
	code    int
	HTTP    int
	message string
}

func (coder Errcode) Error() string {
	return coder.message
}

func (coder Errcode) Code() int {
	return coder.code
}

func (coder Errcode) String() string {
	return coder.message
}

func (coder Errcode) HTTPStatus() int {
	if coder.HTTP == 0 {
		return 500
	}
	return coder.HTTP
}

var codes = map[int]*Errcode{}
var codeMux = &sync.Mutex{}

func register(code int, httpStaus int, message string) {
	if code == 0 {
		panic("code 0 is reserved")
	}
	if _, ok := OnlyUseHTTPStatus[httpStaus]; !ok {
		panic("httstatuscode and code  are not good")
	}

	codeMux.Lock()
	defer codeMux.Unlock()
	errcode := &Errcode{
		code:    code,
		HTTP:    httpStaus,
		message: message,
	}
	codes[code] = errcode
}

func ParseCoder(code int) *Errcode {
	if coder, ok := codes[code]; ok {
		return coder
	}

	return &Errcode{code: 1, HTTP: http.StatusInternalServerError, message: "unknown error"}
}
