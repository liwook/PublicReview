package user

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"
	"regexp"
	"review/dal/model"
	"review/dal/query"
	"review/db"
	"review/middleware"
	"review/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// post /api/v1/send-code
func SendCode(c *gin.Context) {
	var codeRequest sendCodeRequest
	err := c.ShouldBindJSON(&codeRequest)
	if err != nil {
		response.Error(c, response.ErrBind)
		return
	}
	if !isPhoneInvalid(codeRequest.Phone) {
		response.Error(c, response.ErrValidation, "phone is invalid")
		return
	}

	//生成验证码,6位数的字符串
	code := strconv.Itoa(rand.Intn(100000) + 100000)
	//用redis的string类型保存
	key := phoneKeyPrefix + codeRequest.Phone
	success, err := db.RedisDb.SetNX(context.Background(), key, code, codeExpiration).Result()
	if err != nil {
		slog.Error("redis setnx failed", "err", err, "phone", codeRequest.Phone)
		response.Error(c, response.ErrDatabase)
		return
	}
	if !success {
		response.Error(c, response.ErrValidation, "验证码已发送，请稍后再试")
		return
	}

	response.Success(c, gin.H{"code": code})
	//真实环境是调用短信服务商API（如阿里云、腾讯云短信服务）发送验证码到用户手机,不是直接在http请求中返回数据的。
	// 生产环境完整流程：
	// 1. 验证手机号格式和频率限制
	// 2. 生成6位随机验证码
	// 3. 将验证码存储到Redis（设置过期时间）
	// 4. 调用短信服务发送验证码：sms.Send(phone, code)
	// 5. 如果短信发送成功，返回：{"message": "验证码已发送"}
	// 6. 如果短信发送失败，清理Redis并返回错误
	// 7. 记录发送日志：时间、手机号、发送状态等
}

func isPhoneInvalid(phone string) bool {
	// 匹配规则: ^1第一位为一, [345789]{1} 后接一位345789 的数字
	// \\d \d的转义 表示数字 {9} 接9位 ,   $ 结束符
	regRuler := "^1[123456789]{1}\\d{9}$"
	reg := regexp.MustCompile(regRuler) // 正则调用规则
	// 返回 MatchString 是否匹配
	return reg.MatchString(phone)
}

// post /api/v1/login
func Login(c *gin.Context) {
	var loginRequest loginReq
	err := c.ShouldBindJSON(&loginRequest)
	if err != nil {
		slog.Error("failed to bind login request", "err", err)
		response.Error(c, response.ErrBind)
		return
	}
	if !isPhoneInvalid(loginRequest.Phone) {
		response.Error(c, response.ErrValidation, "phone is invalid")
		return
	}

	if err := loginRequest.Validate(); err != nil {
		response.Error(c, response.ErrValidation, err.Error())
		return
	}

	//根据参数判断登录方式
	var user *model.TbUser
	if loginRequest.Password != "" {
		user, err = loginPassword(loginRequest)
	} else if loginRequest.Code != "" {
		user, err = loginCode(loginRequest)
	} else {
		response.Error(c, response.ErrValidation, "login method is invalid")
		return
	}

	if err != nil {
		response.HandleBusinessError(c, err)
		return
	}

	token, err := middleware.GenerateToken(loginRequest.Phone, int64(user.ID))
	if err != nil {
		slog.Error("generate token bad", "err", err)
		response.Error(c, response.ErrLoginFailed)
		return
	}
	response.Success(c, gin.H{
		"token": token,
		"user": userResponse{
			ID:       user.ID,
			Phone:    user.Phone,
			NickName: user.NickName,
			Icon:     user.Icon,
		},
	})
	// response.Success(c, gin.H{
	// 	"user": userResponse{ //返回给前端展示或者缓存的用户信息
	// 		ID:       user.ID,
	// 		Phone:    user.Phone,
	// 		NickName: user.NickName,
	// 		Icon:     user.Icon,
	// 	},
	// })
}

// 验证码登录
func loginCode(login loginReq) (*model.TbUser, error) {
	//为空是返回error中的，值为redis.Nil
	// 获取Redis中存储的验证码
	val, err := db.RedisDb.Get(context.Background(), phoneKeyPrefix+login.Phone).Result()
	if err != nil {
		// if err != redis.Nil
		if errors.Is(err, redis.Nil) { //这种写法更好，智能比较：会递归检查错误链；处理包装错误：即使错误被fmt.Errorf等函数包装，也能正确识别
			return nil, response.NewBusinessError(response.ErrExpired, "验证码过期或没有该验证码")
		}
		// Redis连接错误等系统错误
		return nil, response.WrapBusinessError(response.ErrDatabase, err, "")
	}
	if val != login.Code {
		return nil, response.NewBusinessError(response.ErrLoginFailed, "验证码错误")
	}
	// 验证码验证成功后，删除Redis中的验证码
	db.RedisDb.Del(context.Background(), phoneKeyPrefix+login.Phone)

	//之后判断是否是新用户，若是新用户，就创建
	u := query.TbUser
	user, err := u.Where(u.Phone.Eq(login.Phone)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = &model.TbUser{Phone: login.Phone, NickName: userNickNamePrefix + strconv.Itoa(rand.Intn(100000000))}
			err := u.Create(user)
			if err != nil {
				return nil, response.WrapBusinessError(response.ErrDatabase, err, "")
			}
		}
		return nil, response.WrapBusinessError(response.ErrDatabase, err, "")
	}

	return user, nil
}

func loginPassword(login loginReq) (*model.TbUser, error) {
	//从mysql中判断账号和密码是否正确
	u := query.TbUser
	user, err := u.Where(u.Phone.Eq(login.Phone)).Select(u.Password).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// return response.NewBusinessError(response.ErrLoginFailed, "该用户不存在")
			return nil, response.NewBusinessError(response.ErrLoginFailed, "用户名或密码错误")
		}
		return nil, response.WrapBusinessError(response.ErrDatabase, err, "")
	}
	// 2. 校验密码（数据库存储 BCrypt 哈希值）
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, response.NewBusinessError(response.ErrLoginFailed, "用户名或密码错误")
		}
		// response.Error(c, response.ErrPasswordIncorrect, "密码错误")
		return nil, response.WrapBusinessError(response.ErrUnknown, err, "")
	}

	//用户不存在和密码错误若是返回了不同的错误信息：1.用户不存在："该用户不存在";2密码错误："密码错误"
	//从安全角度考虑，这种区别可能被恶意用户利用来枚举系统中的有效账户。建议将两种情况都返回相同的通用错误信息。
	return user, nil
}
