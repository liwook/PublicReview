package user

import (
	"context"
	"dianping/dal/model"
	"dianping/dal/query"
	"dianping/internal/db"
	"dianping/internal/middleware"
	"dianping/pkg/code"
	"log/slog"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/exp/rand"
)

const (
	UserNickNamePrefix = "user"
	phoneKey           = "phone:"
)

type LoginRequest struct {
	Phone       string `json:"name" binding:"required"`
	CodeOrPwd   string `json:"codeOrPwd" binding:"required"`
	LoginMethod string `json:"loginMethod" binding:"required"`
}

// 得到验证码
// get /user/verificationcode/:phone
func GetVerificationCode(c *gin.Context) {
	phone := c.Param("phone")
	if phone == "" || !isPhoneInvalid(phone) {
		code.WriteResponse(c, code.ErrValidation, "phone is empty or invalid")
		return
	}

	//生成验证码,6位数
	num := rand.Intn(1000000) + 100000
	//用redis的string类型保存
	key := phoneKey + phone
	success, err := db.RedisClient.SetNX(context.Background(), key, num, 4*time.Minute).Result()
	if !success || err != nil {
		code.WriteResponse(c, code.ErrDatabase, nil)
		return
	}

	code.WriteResponse(c, code.ErrSuccess, gin.H{"VerificationCode": num})
}

// post /user/login
func Login(c *gin.Context) {
	var login LoginRequest
	err := c.BindJSON(&login)
	if err != nil {
		slog.Error("codelogin bind bad", "err", err)
		code.WriteResponse(c, code.ErrBind, nil)
		return
	}
	if !isPhoneInvalid(login.Phone) {
		code.WriteResponse(c, code.ErrValidation, "phone is invalid")
		return
	}

	switch login.LoginMethod {
	case "code":
		loginCode(c, login)
	case "password":
		loginPassword(c, login)
	default:
		code.WriteResponse(c, code.ErrValidation, "loginMethod bad")
	}
}

func loginCode(c *gin.Context, login LoginRequest) {
	//为空是返回error中的，值为redis.Nil
	//对比号码是否有验证码
	val, err := db.RedisClient.Get(context.Background(), phoneKey+login.Phone).Result()
	if err == redis.Nil {
		code.WriteResponse(c, code.ErrExpired, "验证码过期或没有该验证码")
		return
	}
	if err != nil {
		slog.Error("redis get bad", "err", err)
		code.WriteResponse(c, code.ErrDatabase, nil)
		return
	}
	if val != login.CodeOrPwd {
		code.WriteResponse(c, code.ErrExpired, "验证码错误")
		return
	}

	//之后判断是否是新用户，若是新用户，就创建
	u := query.TbUser
	count, err := u.Where(u.Phone.Eq(login.Phone)).Count()
	if err != nil {
		slog.Error("find by phone bad", "err", err)
		code.WriteResponse(c, code.ErrDatabase, nil)
		return
	}
	if count == 0 {
		err := u.Create(&model.TbUser{Phone: login.Phone, NickName: UserNickNamePrefix + strconv.Itoa(rand.Intn(100000))})
		if err != nil {
			slog.Error("create user failed", "err", err)
			code.WriteResponse(c, code.ErrDatabase, "create user failed")
			return
		}
	}

	generateTokenResponse(c, login.Phone)
}

func loginPassword(c *gin.Context, login LoginRequest) {
	//从mysql中判断账号和密码是否正确
	u := query.TbUser
	count, err := u.Where(u.Phone.Eq(login.Phone), u.Password.Eq(login.CodeOrPwd)).Count()
	if err != nil {
		slog.Error("find by phone and password bad", "err", err)
		code.WriteResponse(c, code.ErrDatabase, nil)
		return
	}
	if count == 0 {
		code.WriteResponse(c, code.ErrPasswordIncorrect, "phone or password is Incorrect")
		return
	}
	generateTokenResponse(c, login.Phone)
}

func generateTokenResponse(c *gin.Context, phone string) {
	token, err := middleware.GenerateToken(phone)
	if err != nil {
		slog.Error("generate token bad", "err", err)
		code.WriteResponse(c, code.ErrTokenGenerationFailed, nil)
		return
	}
	code.WriteResponse(c, code.ErrSuccess, gin.H{"token": token})
}

func isPhoneInvalid(phone string) bool {
	// 匹配规则: ^1第一位为一, [345789]{1} 后接一位345789 的数字
	// \\d \d的转义 表示数字 {9} 接9位 ,   $ 结束符
	regRuler := "^1[123456789]{1}\\d{9}$"
	reg := regexp.MustCompile(regRuler) // 正则调用规则
	// 返回 MatchString 是否匹配
	return reg.MatchString(phone)
}
