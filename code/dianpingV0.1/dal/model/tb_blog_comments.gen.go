// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"
)

const TableNameTbBlogComment = "tb_blog_comments"

// TbBlogComment mapped from table <tb_blog_comments>
type TbBlogComment struct {
	ID         uint64    `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement:true" json:"id"`                  // 主键
	UserID     uint64    `gorm:"column:user_id;type:bigint unsigned;not null" json:"user_id"`                             // 用户id
	BlogID     uint64    `gorm:"column:blog_id;type:bigint unsigned;not null" json:"blog_id"`                             // 探店id
	ParentID   uint64    `gorm:"column:parent_id;type:bigint unsigned;not null" json:"parent_id"`                         // 关联的1级评论id，如果是一级评论，则值为0
	AnswerID   uint64    `gorm:"column:answer_id;type:bigint unsigned;not null" json:"answer_id"`                         // 回复的评论id
	Content    string    `gorm:"column:content;type:varchar(255);not null" json:"content"`                                // 回复的内容
	Liked      uint64    `gorm:"column:liked;type:int unsigned" json:"liked"`                                             // 点赞数
	Status     uint8     `gorm:"column:status;type:tinyint unsigned" json:"status"`                                       // 状态，0：正常，1：被举报，2：禁止查看
	CreateTime time.Time `gorm:"column:create_time;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"create_time"` // 创建时间
	UpdateTime time.Time `gorm:"column:update_time;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"update_time"` // 更新时间
}

// TableName TbBlogComment's table name
func (*TbBlogComment) TableName() string {
	return TableNameTbBlogComment
}
