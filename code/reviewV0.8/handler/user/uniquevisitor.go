package user

import (
	"context"
	"errors"
	"log/slog"
	"review/db"
	"review/pkg/consts"
	"review/pkg/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	uvKeyPrefix = "UV:"
	dateFormat  = "20060102"
)

type uvStatistics struct {
	BlogId int `json:"blogId"`
	UserId int `json:"userId"`
}

// post /api/v1/unique-visitor
func AddUniqueVisitor(c *gin.Context) {
	var uv uvStatistics
	err := c.ShouldBindJSON(&uv)
	if err != nil {
		slog.Error("bind json error", "error", err)
		response.Error(c, response.ErrBind)
		return
	}
	if uv.BlogId <= 0 || uv.UserId <= 0 {
		response.Error(c, response.ErrValidation, "invalid blogId or userId")
		return
	}

	now := time.Now().Format(dateFormat)
	err = db.RedisDb.PFAdd(context.Background(), buildUVKey(now, strconv.Itoa(uv.BlogId)), uv.UserId).Err()
	if err != nil {
		slog.Error("add unique visitor error", "error", err)
		response.Error(c, response.ErrDatabase)
		return
	}
	response.Success(c, nil)
}

// GET /api/v1/blogs/:blogId/unique-visitor?date=20240101
func GetUniqueVisitor(c *gin.Context) {
	blogId := c.Param(consts.BlogIdKey)
	date := c.Query("date")

	if date == "" {
		date = time.Now().Format(dateFormat)
	}
	if blogId == "" {
		response.Error(c, response.ErrValidation, "blogId is required")
		return
	}
	// 验证日期格式
	if _, err := time.Parse(dateFormat, date); err != nil {
		response.Error(c, response.ErrValidation, "invalid date format, expected YYYYMMDD")
		return
	}
	if _, err := strconv.Atoi(blogId); err != nil {
		response.Error(c, response.ErrValidation, "invalid blogId format")
		return
	}
	res, err := db.RedisDb.PFCount(context.Background(), buildUVKey(date, blogId)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			response.Success(c, 0)
			return
		}

		slog.Error("get unique visitor count error", "error", err, "blogId", blogId, "date", date)
		response.Error(c, response.ErrDatabase)
		return
	}
	response.Success(c, res)
}

func buildUVKey(date, blogId string) string {
	return uvKeyPrefix + date + ":" + blogId
}
