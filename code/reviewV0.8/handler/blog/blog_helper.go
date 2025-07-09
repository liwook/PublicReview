package blog

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"review/dal/query"
	"review/db"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const blogLikedKey = "blog:liked:"

// 判断是否有点赞过，通过redis的set
func isBlogLiked(blogId int, userId int64) bool {
	// key := blogLikedKey + strconv.Itoa(blogId)
	// success, err := db.RedisDb.SIsMember(context.Background(), key, strconv.FormatInt(userId, 10)).Result()
	// if err != nil {
	// 	slog.Error("检查点赞状态失败", "blogId", blogId, "userId", userId, "err", err)
	// 	return false
	// }
	// return success

	// 通过redis的sortedSet
	//当用户没有点赞时，ZScore 会返回错误（通常是 redis.Nil），函数返回 true
	key := blogLikedKey + strconv.Itoa(blogId)
	_, err := db.RedisDb.ZScore(context.Background(), key, strconv.FormatInt(userId, 10)).Result()
	if errors.Is(err, redis.Nil) { // 用户没有点赞过
		return false
	} else if err != nil {
		slog.Error("检查点赞状态失败", "blogId", blogId, "userId", userId, "err", err)
		return false // 发生错误时默认认为没有点赞
	}
	return true
}

func pushBlogToFans(blogId, userId uint64) {
	// 查询粉丝列表
	follow := query.TbFollow
	res, err := follow.Where(follow.FollowUserID.Eq(userId)).Select(follow.UserID).Find()
	if err != nil {
		slog.Error("异步推送：查询粉丝列表失败", "userId", userId, "err", err)
		return
	}

	// 推送给粉丝
	for _, v := range res {
		err := db.RedisDb.ZAdd(context.Background(), feedKey+fmt.Sprint(v.UserID), redis.Z{
			Score:  float64(time.Now().UnixMicro()),
			Member: blogId,
		}).Err()
		if err != nil {
			slog.Error("异步推送失败", "fanUserId", v.UserID, "blogId", blogId, "err", err)
		}
	}

	slog.Info("博客推送完成", "blogId", blogId, "fanCount", len(res))
}

func getTbBlogSelect() query.ITbBlogDo {
	return query.TbBlog.Select(query.TbBlog.CreateTime, query.TbBlog.ID, query.TbBlog.Title, query.TbBlog.Content, query.TbBlog.Liked, query.TbBlog.Comments, query.TbBlog.Images, query.TbBlog.UpdateTime)
}
