package user

import (
	"fmt"
	"time"
)

const (
	userNickNamePrefix = "user"
	phoneKeyPrefix     = "phone:"
	codeExpiration     = 4 * time.Minute
)

type sendCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type userResponse struct {
	ID       uint64 `json:"id"`
	Phone    string `json:"phone"`
	NickName string `json:"nick_name"`
	Icon     string `json:"icon"`
}

// loginReq 登录请求结构体（统一使用 phone 作为用户标识）
type loginReq struct {
	Phone    string `json:"phone" binding:"required"`                  // 手机号（必填）
	Password string `json:"password" binding:"omitempty,min=6,max=20"` // 密码（仅密码登录时用）
	Code     string `json:"code" binding:"omitempty,min=4,max=6"`      // 验证码（仅验证码登录时用）
}

// Validate 校验登录方式是否合法（自定义校验）
func (l *loginReq) Validate() error {
	hasPassword := l.Password != ""
	hasCode := l.Code != ""

	if !hasPassword && !hasCode {
		return fmt.Errorf("请选择密码或验证码登录")
	}
	if hasPassword && hasCode {
		return fmt.Errorf("不能同时使用密码和验证码登录")
	}
	return nil
}
