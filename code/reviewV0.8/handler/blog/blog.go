package blog

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"review/dal/model"
	"review/dal/query"
	"review/db"

	"review/middleware"
	"review/pkg/consts"
	"review/pkg/response"
	"review/pkg/util"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gen"
	"gorm.io/gorm"
)

const (
	savePath            = "/images"
	maxFileSize         = 5 << 20 // 5MB
	defaultBlogsPerPage = 10
	feedKey             = "feed:"
	rawBlogSql          = "SELECT id, title, content, liked, comments, images, update_time, create_time FROM tb_blog WHERE id IN (%s) ORDER BY FIELD(id,%s)"
)

// 上传博客图片
// post /api/v1/blog/images
func UploadImages(c *gin.Context) {
	file, err := c.FormFile(consts.FormFileKey)
	if err != nil {
		slog.Error("上传图片失败", "err", err)
		response.Error(c, response.ErrDecodingFailed)
		return
	}

	// 验证文件大小 (例如限制5MB)
	if file.Size > maxFileSize {
		slog.Error("文件大小超出限制", "size", file.Size)
		response.Error(c, response.ErrValidation, "文件大小不能超过5MB")
		return
	}

	// 验证文件类型
	if !util.IsValidImageType(file.Filename) {
		slog.Error("不支持的文件类型", "filename", file.Filename)
		response.Error(c, response.ErrValidation, "这个是不支持的文件类型")
		return
	}

	//按照 日期 来保存图片
	// images/202401/01/xxxx.jpg	就是按照 年月/日/xxx.jpg 来保存图片
	//找到当前年月
	currentTime := time.Now()
	yearMonth := currentTime.Format("2006-01")
	day := currentTime.Format("02")
	name := util.CreateNewFileName(file.Filename)
	dst := filepath.Join(savePath, yearMonth, day, name)
	// 检查,要是没有该目录，就创建目标目录
	dstDir := filepath.Dir(dst)
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			slog.Error("创建目录失败", "path", dstDir, "err", err)
			response.Error(c, response.ErrCreateFailed)
			return
		}
	}

	//保存文件
	if err := c.SaveUploadedFile(file, dst); err != nil {
		slog.Error("保存文件失败", "path", dst, "err", err)
		response.Error(c, response.ErrSaveFailed)
		return
	}

	// 返回图片url
	publicURL := filepath.Join("/images", yearMonth, day, name)
	publicURL = filepath.ToSlash(publicURL) //将路径中的分隔符统一转换为正斜杠（/）
	response.Success(c, gin.H{"url": publicURL, "filename": name})
}

// 发布博客内容，保存到数据库
// post /api/v1/blogs
func SaveBlog(c *gin.Context) {
	var blog model.TbBlog
	if err := c.ShouldBindJSON(&blog); err != nil {
		slog.Error("绑定JSON失败", "err", err)
		response.Error(c, response.ErrBind)
		return
	}

	// 将博客模型保存到数据库中
	if err := query.TbBlog.Create(&blog); err != nil {
		slog.Error("保存博客失败", "err", err)
		response.Error(c, response.ErrDatabase)
		return
	}

	// 将博客内容推送给粉丝,异步操作
	//要是担忧会开启很多新协程的话，那可以使用协程池的的，指定最多能开启的协程数量，防止开启过多的协程
	go pushBlogToFans(blog.ID, blog.UserID)

	// 返回保存成功的博客ID
	response.Success(c, gin.H{"blogId": blog.ID})
}

// get /api/v1/blogs/:blogId
func GetBlogById(c *gin.Context) {
	// 获取url中的id参数
	id := c.Param(consts.BlogIdKey)
	// 将id转换为int类型
	idInt, err := strconv.Atoi(id)
	if err != nil || idInt <= 0 {
		response.Error(c, response.ErrValidation, "id必须是正整数")
		return
	}

	//通过blog找到该blog的用户id
	queryBlog := query.TbBlog
	// 查询该id的blog
	blog, err := queryBlog.Where(queryBlog.ID.Eq(uint64(idInt))).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, response.ErrNotFound, "博客不存在")
			return
		}
		slog.Error("查询博客失败", "err", err)
		response.Error(c, response.ErrDatabase)
		return
	}

	// 查询该blog的用户
	user, err := query.TbUser.Where(query.TbUser.ID.Eq(blog.UserID)).First()
	// 如果查询不到该用户，则返回错误
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, response.ErrNotFound)
			return
		}
		// 记录错误日志
		slog.Error("查询用户失败", "err", err)
		response.Error(c, response.ErrDatabase)
		return
	}

	// 返回查询结果
	resp := singleBlogResponse{
		Blog:     convertOneTbBlogPtrToResponse(blog),
		Username: user.NickName,
		Icon:     user.Icon, // 头像
	}

	// 判断是否登录
	IsLogin := c.GetBool(middleware.CtxKeyIsAuthenticated)
	if IsLogin {
		resp.IsLiked = isBlogLiked(idInt, c.GetInt64(middleware.CtxKeyUserId))
	}
	response.Success(c, resp)
}

// 点赞博客
// POST /api/v1/blogs/:blogId/like
func LikeBlog(c *gin.Context) {
	//获取blogId,被点赞的博客id
	id := c.Param(consts.BlogIdKey)
	idInt, err := strconv.Atoi(id)
	if err != nil || idInt <= 0 {
		response.Error(c, response.ErrValidation, "id必须是正整数")
		return
	}

	userId := c.GetInt64(middleware.CtxKeyUserId)
	isLiked := isBlogLiked(idInt, userId)
	redisKey := blogLikedKey + id
	blogQuery := query.TbBlog.Where(query.TbBlog.ID.Eq(uint64(idInt)))
	var info gen.ResultInfo

	// tbBlog := query.TbBlog
	if isLiked { //已经点赞过了，则取消点赞
		info, err = blogQuery.UpdateSimple(query.TbBlog.Liked.Sub(1))
	} else {
		info, err = blogQuery.UpdateSimple(query.TbBlog.Liked.Add(1))
	}

	if err != nil {
		action := map[bool]string{true: "取消点赞", false: "点赞"}[isLiked]
		slog.Error(action+"失败", "err", err)
		response.Error(c, response.ErrDatabase, action+"失败")
		return
	}
	if info.RowsAffected == 0 { //没有更新到数据，则说明没有点赞过
		response.Error(c, response.ErrNotFound, "博客不存在")
		return
	}

	if isLiked {
		// db.RedisDb.SRem(context.Background(), redisKey, userId)
		err := db.RedisDb.ZRem(context.Background(), redisKey, userId).Err()
		if err != nil {
			slog.Error("Redis操作失败", "err", err)
		}
	} else {
		// if err := db.RedisDb.SAdd(context.Background(), redisKey, userId).Err(); err != nil {
		// 	slog.Error("Redis操作失败", "err", err)
		// }
		err := db.RedisDb.ZAdd(context.Background(), redisKey, redis.Z{Score: float64(time.Now().Unix()), Member: userId}).Err()
		if err != nil {
			slog.Error("Redis操作失败", "err", err)
		}
	}

	msg := map[bool]string{true: "取消点赞成功", false: "点赞成功"}[isLiked]
	response.Success(c, msg)
}

// GET /api/v1/blogs/:blogId/likes
func GetBlogLikes(c *gin.Context) {
	//获取blogId，验证ID
	id := c.Param(consts.BlogIdKey)
	idInt, err := strconv.Atoi(id)
	if err != nil || idInt <= 0 {
		response.Error(c, response.ErrValidation, "id必须是正整数")
		return
	}

	//获取点赞的用户id列表
	userIds, err := db.RedisDb.ZRevRangeWithScores(context.Background(), blogLikedKey+id, 0, 4).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// 没有点赞记录，返回空列表
			response.Success(c, BloglikedUsersResponse{
				LikedUsers: convertTbUsersTolikedUsers([]model.TbUser{}),
			})
			return
		}
		slog.Error("获取点赞用户列表失败", "err", err)
		response.Error(c, response.ErrDatabase)
		return
	}
	if len(userIds) == 0 {
		response.Success(c, BloglikedUsersResponse{
			LikedUsers: convertTbUsersTolikedUsers([]model.TbUser{}),
		})
		return
	}

	Ids := make([]string, 0, len(userIds))
	placeholders := make([]string, 0, len(userIds))
	for _, userId := range userIds {
		if id, ok := userId.Member.(string); ok {
			Ids = append(Ids, id)
			placeholders = append(placeholders, "?")
		}
	}

	// 构建安全的SQL查询,防止sql注入
	inClause := strings.Join(placeholders, ",")

	// 合并参数：IN子句的参数 + FIELD函数的参数
	allArgs := make([]any, len(Ids)*2)
	for i, id := range Ids {
		allArgs[i] = id          // 用于IN子句
		allArgs[len(Ids)+i] = id // 用于FIELD函数
	}

	// 在数据库中查找到该些用户的信息和图标
	var dbUsers []model.TbUser
	sql := fmt.Sprintf("SELECT id, nick_name, icon FROM tb_user WHERE id IN (%s) ORDER BY FIELD(id,%s)",
		inClause, inClause)
	err = db.DBEngine.Raw(sql, allArgs...).Scan(&dbUsers).Error
	if err != nil {
		slog.Error("获取点赞用户信息失败", "err", err)
		response.Error(c, response.ErrDatabase)
		return
	}

	response.Success(c, BloglikedUsersResponse{
		LikedUsers: convertTbUsersTolikedUsers(dbUsers),
	})
}

// GET /api/v1/users/:userId/blogs
func GetBlogsByUserId(c *gin.Context) {
	userId := c.Param(consts.UserIdKey)
	lastId := c.Query(consts.LastIdKey) // 游标值，lastId作为查询参数
	//使用形式: /api/v1/users/123/blogs?lastId=456
	userIdInt, err := strconv.Atoi(userId)
	if err != nil || userIdInt <= 0 {
		slog.Error("参数验证失败", "userId", userId, "err", err)
		response.Error(c, response.ErrValidation, "userId must be a positive integer")
		return
	}

	var lastIdInt uint64 = 0
	if lastId != "" {
		parsed, err := strconv.ParseUint(lastId, 10, 64)
		if err != nil {
			slog.Error("参数验证失败", "lastId", lastId, "err", err)
			response.Error(c, response.ErrValidation, "lastId must be a valid integer")
			return
		}
		lastIdInt = parsed
	}

	tbBlog := query.TbBlog
	blogQuery := getTbBlogSelect().Where(tbBlog.UserID.Eq(uint64(userIdInt))) //这种是lastId为空或者为0的情况
	if lastIdInt > 0 {
		// 游标分页，查询ID小于lastId的记录（降序分页）
		blogQuery = getTbBlogSelect().Where(tbBlog.UserID.Eq(uint64(userIdInt)), tbBlog.ID.Lt(uint64(lastIdInt)))
	}

	// 多查一条来判断是否还有下一页
	blogs, err := blogQuery.Order(tbBlog.ID.Desc()).Limit(defaultBlogsPerPage + 1).Find()
	if err != nil {
		slog.Error("查询博客列表失败", "userId", userIdInt, "lastId", lastIdInt, "err", err)
		response.Error(c, response.ErrDatabase)
		return
	}
	// 判断是否还有下一页
	hasMore := len(blogs) > defaultBlogsPerPage
	if hasMore {
		blogs = blogs[:defaultBlogsPerPage]
	}

	// 构建响应
	resp := blogListResponse{
		IsEnd: !hasMore,
		Blogs: convertTbBlogPtrToResponse(blogs),
	}

	// 如果还有更多数据，设置下一页的游标
	if hasMore && len(blogs) > 0 {
		resp.NextId = &blogs[len(blogs)-1].ID
	}

	response.Success(c, resp)
}

// GET /api/v1/users/:userId/following-blogs?offset=0&maxTime=1234567890
func QueryBlogOfFollow(c *gin.Context) {
	userid := c.Param(consts.UserIdKey)
	offset := c.Query(consts.OffsetKey)
	maxTime := c.DefaultQuery(consts.MaxTimeKey, strconv.FormatInt(time.Now().UnixMicro(), 10))
	// 参数验证
	if userid == "" {
		response.Error(c, response.ErrValidation, "userId is required")
		return
	}

	offsetInt, err := strconv.Atoi(offset)
	if err != nil || offsetInt < 0 {
		response.Error(c, response.ErrValidation, "offset must be a non-negative integer")
		return
	}

	//按照score从大到小排序，即是按照时间戳从大到小排序
	res, err := db.RedisDb.ZRevRangeByScoreWithScores(context.Background(), feedKey+userid,
		&redis.ZRangeBy{Min: "0", Max: maxTime, Offset: int64(offsetInt), Count: defaultBlogsPerPage}).Result()
	if err != nil {
		slog.Error("Redis操作失败", "err", err)
		response.Error(c, response.ErrDatabase)
		return
	}

	//解析数据： blogId，minTime（时间戳）， offset
	minTime := int64(0) //这个minTime是上次查询的最小时间戳，作为当次查询的最大时间戳来开始查
	resOffset := 0
	ids := make([]string, 0, len(res))
	for _, v := range res {
		//安全的类型断言
		blogId, ok := v.Member.(string)
		if !ok {
			slog.Warn("Invalid member type in Redis result", "member", v.Member)
			continue
		}
		ids = append(ids, blogId)

		//获取分数      判读得到最后一次的时间戳，以及偏移量
		currentTime := int64(v.Score)
		if minTime == 0 {
			//第一条记录，初始化
			minTime = currentTime
			resOffset = 1 // 设置为1，因为当前记录就是第一个具有这个时间戳的记录
		} else if currentTime == minTime { //该时间戳有相同的，则偏移量加1。这样就可以去掉了因为socore相同(即时间戳相同），导致返回数据重复的问题
			resOffset++
		} else if currentTime < minTime { // 找到更小的时间戳，更新minTime并重置offset
			minTime = currentTime
			resOffset = 1 // 设置为1，因为当前记录就是第一个具有这个时间戳的记录
		}
	}

	if len(ids) == 0 {
		resp := followBlogsResponse{
			Blogs:   convertTbBlogToResponse([]model.TbBlog{}),
			Offset:  0,
			MinTime: 0,
		}
		response.Success(c, resp)
		return
	}

	// 构建安全的参数化查询,防止sql注入
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids)*2)

	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id          // 用于IN子句
		args[len(ids)+i] = id // 用于FIELD函数
	}

	inClause := strings.Join(placeholders, ",")

	var blogs []model.TbBlog
	err = db.DBEngine.Raw(fmt.Sprintf(rawBlogSql, inClause, inClause), args...).Scan(&blogs).Error
	if err != nil {
		slog.Error("查询博客详情失败", "userId", userid, "err", err)
		response.Error(c, response.ErrDatabase)
		return
	}

	resp := followBlogsResponse{
		Blogs:   convertTbBlogToResponse(blogs),
		Offset:  resOffset,
		MinTime: minTime,
	}
	response.Success(c, resp)
}
