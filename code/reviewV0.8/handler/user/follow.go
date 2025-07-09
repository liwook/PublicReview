package user

import (
	"context"
	"log/slog"
	"review/dal/model"
	"review/dal/query"
	"review/db"
	"review/pkg/consts"
	"review/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

const followUserId = "follow:userId:"

// 判断是否关注
// GET /api/v1/users/:userId/follow/:followId
func IsFollow(c *gin.Context) {
	userId, followId, err := validateFollowParams(c.Param(consts.UserIdKey), c.Param(consts.FollowIdKey))
	if err != nil {
		response.HandleBusinessError(c, err)
		return
	}

	f := query.TbFollow
	count, err := f.Where(f.UserID.Eq(uint64(userId)), f.FollowUserID.Eq(uint64(followId))).Count()
	if err != nil {
		slog.Error("查询关注关系失败", "userId", userId, "followId", followId, "err", err)
		response.Error(c, response.ErrDatabase)
		return
	}
	response.Success(c, gin.H{"isAlreadyFollow": count > 0})
}

// 关注
// POST /api/v1/users/:userId/follow/:followId
func FollowUser(c *gin.Context) {
	userId, followId, err := validateFollowParams(c.Param(consts.UserIdKey), c.Param(consts.FollowIdKey))
	if err != nil {
		response.HandleBusinessError(c, err)
		return
	}

	f := query.TbFollow
	// 检查是否已经关注
	count, err := f.Where(f.UserID.Eq(uint64(userId)), f.FollowUserID.Eq(uint64(followId))).Count()
	if err != nil {
		slog.Error("查询关注状态失败", "userId", userId, "followId", followId, "err", err)
		response.Error(c, response.ErrDatabase)
		return
	}
	if count > 0 {
		response.Error(c, response.ErrValidation, "already followed")
		return
	}

	err = f.Create(&model.TbFollow{
		UserID:       uint64(userId),
		FollowUserID: uint64(followId),
	})
	if err != nil {
		slog.Error("创建关注关系失败", "userId", userId, "followId", followId, "err", err)
		response.Error(c, response.ErrDatabase, "follow failed")
		return
	}

	//在redis中set添加关注的对象
	err = db.RedisDb.SAdd(context.Background(), followUserId+strconv.Itoa(userId), followId).Err()
	if err != nil {
		slog.Error("Redis添加关注失败", "userId", userId, "followId", followId, "err", err)
		// 这里可以选择是否要回滚数据库操作，或者仅记录日志
	}
	response.Success(c, "follow success")
}

// 取消关注
// DELETE /api/v1/users/:userId/follow/:followId
func UnfollowUser(c *gin.Context) {
	userId, followId, err := validateFollowParams(c.Param(consts.UserIdKey), c.Param(consts.FollowIdKey))
	if err != nil {
		response.HandleBusinessError(c, err)
		return
	}

	f := query.TbFollow
	info, err := f.Where(f.UserID.Eq(uint64(userId)), f.FollowUserID.Eq(uint64(followId))).Delete()
	if err != nil {
		slog.Error("取消关注失败", "userId", userId, "followId", followId, "err", err)
		response.Error(c, response.ErrDatabase, "unfollow failed")
		return
	}
	if info.RowsAffected == 0 {
		response.Error(c, response.ErrNotFound, "关注关系不存在")
		return
	}

	//在redis中移除关注的对象
	err = db.RedisDb.SRem(context.Background(), followUserId+strconv.Itoa(int(userId)), followId).Err()
	if err != nil {
		slog.Error("Redis移除关注失败", "userId", userId, "followId", followId, "err", err)
	}

	response.Success(c, "unfollow success")
}

// GET /api/v1/users/follow/commons
func FollowCommons(c *gin.Context) {
	// GET /api/v1/users/follow/commons?user1=123&user2=456
	user1 := c.Query(consts.User1Key)
	user2 := c.Query(consts.User2Key)
	// 参数验证
	if user1 == "" || user2 == "" {
		response.Error(c, response.ErrValidation, "user1和user2参数不能为空")
		return
	}

	// 验证user1
	user1Int, err := strconv.Atoi(user1)
	if err != nil || user1Int <= 0 {
		slog.Error("参数验证失败", "user1", user1, "err", err)
		response.Error(c, response.ErrValidation, "user1必须是正整数")
		return
	}

	// 验证user2
	user2Int, err := strconv.Atoi(user2)
	if err != nil || user2Int <= 0 {
		slog.Error("参数验证失败", "user2", user2, "err", err)
		response.Error(c, response.ErrValidation, "user2必须是正整数")
		return
	}

	// 检查是否查询同一个用户的共同关注（无意义）
	if user1 == user2 {
		response.Error(c, response.ErrValidation, "不能查询同一个用户的共同关注")
		return
	}

	res, err := db.RedisDb.SInter(context.Background(), followUserId+user1, followUserId+user2).Result()
	if err != nil {
		slog.Error("查询共同关注失败", "user1", user1, "user2", user2, "err", err)
		response.Error(c, response.ErrDatabase, "查询共同关注失败")
		return
	}

	// 添加成功日志
	slog.Info("查询共同关注成功", "user1", user1, "user2", user2, "count", len(res))
	response.Success(c, gin.H{"followCommons": res})
}

func validateFollowParams(userId, followId string) (userIdInt, followIdInt int, err error) {
	// 检查参数是否为空
	if userId == "" || followId == "" {
		return 0, 0, response.NewBusinessError(response.ErrValidation, "userId or followId is empty")
	}

	// 验证userId
	userIdInt, err = strconv.Atoi(userId)
	if err != nil || userIdInt <= 0 {
		return 0, 0, response.WrapBusinessError(response.ErrValidation, err, "userId must be a positive integer")
	}

	// 验证followId
	followIdInt, err = strconv.Atoi(followId)
	if err != nil || followIdInt <= 0 {
		return 0, 0, response.WrapBusinessError(response.ErrValidation, err, "followId must be a positive integer")
	}

	// 检查用户不能关注自己
	if userIdInt == followIdInt {
		return 0, 0, response.NewBusinessError(response.ErrValidation, "用户不能关注自己")
	}

	return userIdInt, followIdInt, nil
}
