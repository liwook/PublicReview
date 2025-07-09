package blog

import (
	"review/dal/model"
	"strings"
	"time"
)

// API响应专用结构体
type blogDTO struct {
	ID           uint64    `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Images       []string  `json:"images"`     // 将逗号分隔字符串转换为数组
	LikeCount    uint32    `json:"like_count"` // 重命名字段，更符合API规范
	CommentCount uint32    `json:"comment_count"`
	CreateTime   time.Time `json:"create_time"`
	UpdateTime   time.Time `json:"update_time"`
	AuthorID     uint64    `json:"author_id"` // 新增字段，博客作者的主键ID
	// 注意：不包含 ShopID 等内部字段
}

type blogListResponse struct {
	IsEnd  bool      `json:"is_end"`
	Blogs  []blogDTO `json:"blogs"`
	NextId *uint64   `json:"nextId,omitempty"` // 下一页的游标,使用指针可以明确区分"没有值"和"值为0"的情况
}

// 关注博客查询响应结构体
type followBlogsResponse struct {
	Blogs   []blogDTO `json:"blogs"`   // 博客列表
	Offset  int       `json:"offset"`  // 下次查询的偏移量
	MinTime int64     `json:"minTime"` // 当前批次的最小时间戳，用作下次查询的maxTime
}

func convertTbBlogToResponse(dbBlogs []model.TbBlog) []blogDTO {
	// 预分配容量，避免切片扩容
	apiBlogs := make([]blogDTO, 0, len(dbBlogs))

	for _, blog := range dbBlogs {
		var images []string
		if blog.Images != "" {
			images = strings.Split(blog.Images, ",")
		}

		apiBlogs = append(apiBlogs, blogDTO{
			ID:           blog.ID,
			Title:        blog.Title,
			Content:      blog.Content,
			Images:       images,
			LikeCount:    blog.Liked,
			CommentCount: blog.Comments,
			CreateTime:   blog.CreateTime,
			UpdateTime:   blog.UpdateTime,
			AuthorID:     blog.UserID, // 赋值作者ID
		})
	}
	return apiBlogs
}

// 处理指针切片的转换函数（用于GORM Gen）
func convertTbBlogPtrToResponse(dbBlogs []*model.TbBlog) []blogDTO {
	apiBlogs := make([]blogDTO, len(dbBlogs))
	for i, blog := range dbBlogs {
		var images []string
		if blog.Images != "" {
			images = strings.Split(blog.Images, ",")
		}

		apiBlogs[i] = blogDTO{
			ID:           blog.ID,
			Title:        blog.Title,
			Content:      blog.Content,
			Images:       images,
			LikeCount:    blog.Liked,
			CommentCount: blog.Comments,
			CreateTime:   blog.CreateTime,
			UpdateTime:   blog.UpdateTime,
			AuthorID:     blog.UserID, // 赋值作者ID
		}
	}
	return apiBlogs
}

func convertOneTbBlogPtrToResponse(blog *model.TbBlog) blogDTO {
	var images []string
	if blog.Images != "" {
		images = strings.Split(blog.Images, ",")
	}

	return blogDTO{
		ID:           blog.ID,
		Title:        blog.Title,
		Content:      blog.Content,
		Images:       images,
		LikeCount:    blog.Liked,
		CommentCount: blog.Comments,
		CreateTime:   blog.CreateTime,
		UpdateTime:   blog.UpdateTime,
		AuthorID:     blog.UserID, // 赋值作者ID
	}
}

// singleBlogResponse 定义获取单个博客信息接口的响应结构体
type singleBlogResponse struct {
	Blog     blogDTO `json:"blog"`
	Username string  `json:"username"`
	Icon     string  `json:"icon"`
	IsLiked  bool    `json:"is_liked,omitempty"`
}

// likedUser 定义点赞用户信息的结构体
type likedUser struct {
	UserID   uint64 `json:"user_id"`   // 点赞用户的 ID
	NickName string `json:"nick_name"` // 点赞用户的昵称
	Icon     string `json:"icon"`      // 点赞用户的头像
}

// BloglikedUsersResponse 定义获取博客点赞用户列表接口的响应结构体
type BloglikedUsersResponse struct {
	LikedUsers []likedUser `json:"liked_users"` // 点赞用户列表
}

// convertTbUserTolikedUser 将单个 model.TbUser 转换为 likedUser
func convertTbUserTolikedUser(user model.TbUser) likedUser {
	return likedUser{
		UserID:   user.ID,
		NickName: user.NickName,
		Icon:     user.Icon,
	}
}

// convertTbUsersTolikedUsers 将 model.TbUser 切片转换为 likedUser 切片
func convertTbUsersTolikedUsers(users []model.TbUser) []likedUser {
	likedUsers := make([]likedUser, len(users))
	for i, user := range users {
		likedUsers[i] = convertTbUserTolikedUser(user)
	}
	return likedUsers
}
