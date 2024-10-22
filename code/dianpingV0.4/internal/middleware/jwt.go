package middleware

import (
	"crypto/sha1"
	"dianping/internal/config"
	"dianping/pkg/code"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	Phone                string
	jwt.RegisteredClaims // v5版本新加的方法
}

func GetJWTSecret() []byte {
	return []byte(config.JwtOption.Secret)
}

func GenerateToken(phone string) (string, error) {
	//sha1加密phone
	hash := sha1.New()
	hash.Write([]byte(phone))
	claims := UserClaims{
		Phone: string(hash.Sum(nil)),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.JwtOption.Expire)),
			Issuer:    config.JwtOption.Issuer,
			NotBefore: jwt.NewNumericDate(time.Now()), //生效时间
		},
	}

	//使用指定的加密方式(hs256)和声明类型创建新令牌
	tokenStruct := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//获得完整的签名的令牌
	return tokenStruct.SignedString(GetJWTSecret())
}

func ParseToken(token string) (*UserClaims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &UserClaims{}, func(token *jwt.Token) (any, error) {
		return GetJWTSecret(), nil
	})
	if err != nil {
		return nil, err
	}

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*UserClaims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, err
}

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		//登录和获取验证码是不用JWT验证的
		if c.Request.RequestURI == "/user/login" || c.Request.RequestURI == "/user/getcode" {
			return
		}

		ecode := code.ErrSuccess
		token := c.GetHeader("token")

		if token == "" {
			ecode = code.ErrInvalidAuthHeader
		} else {
			_, err := ParseToken(token)
			if err != nil {
				ecode = code.ErrTokenInvalid
			}
		}
		if ecode != code.ErrSuccess {
			code.WriteResponse(c, ecode, nil)
			c.Abort()
			return
		}
		c.Next()
	}
}
