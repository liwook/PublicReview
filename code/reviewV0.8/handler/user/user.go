package user

import (
	"context"
	"errors"
	"log/slog"
	"review/dal/query"
	"review/db"
	"review/pkg/consts"
	"review/pkg/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	signKeyPre      = "sign:"
	yearMonthFormat = "200601" // YYYYMM格式，用于Redis键名
	expireDays      = 90       // 签到数据过期天数
	expireDuration  = expireDays * 24 * time.Hour
)

// 查看用户主页
// GET /api/v1/user/:userId
func QueryUserById(c *gin.Context) {
	id := c.Param(consts.UserIdKey)
	idInt, err := strconv.Atoi(id)
	if err != nil || idInt <= 0 {
		slog.Error("参数验证失败", "id", id, "err", err)
		response.Error(c, response.ErrValidation, "id must be a positive integer")
		return
	}
	u := query.TbUserInfo
	info, err := u.Where(u.UserID.Eq(uint64(idInt))).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, response.ErrNotFound, "用户不存在")
		return
	}
	if err != nil {
		slog.Error("Database error", "err", err)
		response.Error(c, response.ErrDatabase)
		return
	}
	response.Success(c, userInfo{
		UserID:    info.UserID,
		City:      info.City,
		Introduce: info.Introduce,
		Fans:      info.Fans,
		Followee:  info.Followee,
		Gender:    info.Gender,
		Credits:   info.Credits,
		Level:     info.Level,
		Birthday:  info.Birthday,
	})
}

// post /api/v1/user/:userId/signIn
func SignIn(c *gin.Context) {
	userId := c.Param(consts.UserIdKey)
	userIdInt, err := strconv.Atoi(userId)
	if err != nil || userIdInt <= 0 {
		response.Error(c, response.ErrValidation, "invalid userId")
		return
	}

	key := signKeyPre + userId + ":" + time.Now().Format(yearMonthFormat)
	dayOfMonth := time.Now().Day()
	//（从右到左读取，右边是最低位）,使用int64(dayOfMonth)-1，所以最低位是第一天; 1号：dayOfMonth = 1，索引 = 1-1 = 0（最低位）
	oldBit, err := db.RedisDb.SetBit(context.Background(), key, int64(dayOfMonth)-1, 1).Result() // 1为签到，0为未签到
	if err != nil {
		slog.Error("Database error", "err", err)
		response.Error(c, response.ErrDatabase)
		return
	}
	if oldBit == 1 { // 重复签到
		response.Success(c, "already signed in today")
		return
	}
	// 首次签到
	// 设置签到数据3个月后过期，防止Redis内存占用过多
	db.RedisDb.Expire(context.Background(), key, expireDuration)

	response.Success(c, "sign in successful")
}

// 统计到当前时间的连续签到次数
// get  /api/v1/user/:userId/signin-statistics
func ContinuousSigninStatistics(c *gin.Context) {
	userId := c.Param(consts.UserIdKey)
	userIdInt, err := strconv.Atoi(userId)
	if err != nil || userIdInt <= 0 {
		response.Error(c, response.ErrValidation, "invalid userId")
		return
	}

	key := signKeyPre + userId + ":" + time.Now().Format(yearMonthFormat)
	dayOfMonth := time.Now().Day()

	// 类型u代表无符号十进制，i代表带符号十进制
	//0表示偏移量。 从偏移量offset=0开始取dayOfMonth位，获取无符号整数的值（将前dayOfMonth-1位二进制转为无符号10进制返回）
	//res 是一个 []int64 数组，其长度等于 BITFIELD 命令中子命令的数量（此处只有 1 个子命令 GET，因此 len(res) == 1）
	res, err := db.RedisDb.BitField(context.Background(), key, "GET", "u"+strconv.Itoa(dayOfMonth), 0).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			response.Error(c, response.ErrNotFound, "用户没有签到记录")
			return
		}
		slog.Error("BitField error", "err", err)
		response.Error(c, response.ErrDatabase)
		return
	}

	if len(res) == 0 {
		response.Success(c, signInStats{Count: 0})
		return
	}

	count := getSignInStats(res[0], dayOfMonth)
	response.Success(c, signInStats{Count: count})
}

func getSignInStats(num int64, dayOfMonth int) int {
	// 从当天开始找到最近的签到日期
	lastSignDay := -1
	for day := dayOfMonth; day >= 1; day-- {
		bitPos := day - 1
		if (num>>bitPos)&1 == 1 {
			lastSignDay = day
			break
		}
	}

	if lastSignDay == -1 {
		return 0
	}

	// 从最近签到日期开始向前统计连续天数
	count := 0
	for day := lastSignDay; day >= 1; day-- {
		bitPos := day - 1
		if (num>>bitPos)&1 == 0 {
			break
		}
		count++
	}
	return count
}
