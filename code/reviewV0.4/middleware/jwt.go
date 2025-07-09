package middleware

import (
	"fmt"
	"review/config"
	"review/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	CtxKeyUserPhone       = "userPhone"
	CtxKeyUserId          = "userId"
	CtxKeyIsAuthenticated = "isAuthenticated"
)

type userClaims struct {
	UserId               int64
	Phone                string
	jwt.RegisteredClaims // v5版本新加的方法
}

func GetJWTSecret() []byte {
	return []byte(config.JwtOption.Secret)
}

// OptionalJWT 可选的JWT认证中间件，总是尝试解析token
func OptionalJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			// 未提供token，设置未认证状态
			c.Set(CtxKeyIsAuthenticated, false)
			c.Next()
			return
		}

		//处理Bearer token格式
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		} else {
			// token格式错误，设置未认证状态
			c.Set(CtxKeyIsAuthenticated, false)
			c.Next()
			return
		}

		claims, err := ParseToken(token)
		if err != nil {
			// token解析失败，设置未认证状态
			c.Set(CtxKeyIsAuthenticated, false)
			c.Next()
			return
		}

		// 验证关键字段是否为空
		if claims.Phone == "" || claims.UserId == 0 {
			// 字段为空，设置未认证状态
			c.Set(CtxKeyIsAuthenticated, false)
			c.Next()
			return
		}

		// token有效且字段完整，设置用户信息
		c.Set(CtxKeyUserPhone, claims.Phone)
		c.Set(CtxKeyUserId, claims.UserId)
		c.Set(CtxKeyIsAuthenticated, true)
		c.Next()
	}
}

// RequireAuth 强制认证中间件，检查用户是否已登录
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAuthenticated := c.GetBool(CtxKeyIsAuthenticated)
		if !isAuthenticated {
			response.Error(c, response.ErrInvalidAuthHeader, "请先登录")
			c.Abort()
			return
		}
		c.Next()
	}
}

func GenerateToken(phone string, userId int64) (string, error) {
	claims := userClaims{
		Phone:  phone,
		UserId: userId,
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

func ParseToken(token string) (*userClaims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &userClaims{}, func(token *jwt.Token) (any, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return GetJWTSecret(), nil
	})
	if err != nil {
		return nil, err
	}

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*userClaims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, err
}
