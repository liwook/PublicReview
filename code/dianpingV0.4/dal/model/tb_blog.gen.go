// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"
)

const TableNameTbBlog = "tb_blog"

// TbBlog mapped from table <tb_blog>
type TbBlog struct {
	ID         uint64    `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement:true" json:"id"` // 主键
	ShopID     int64     `gorm:"column:shop_id;type:bigint;not null" json:"shop_id"`                     // 商户id
	UserID     uint64    `gorm:"column:user_id;type:bigint unsigned;not null" json:"user_id"`            // 用户id
	Title      string    `gorm:"column:title;type:varchar(255);not null" json:"title"`                   // 标题
	Images     string    `gorm:"column:images;type:varchar(2048)" json:"images"`
	Content    string    `gorm:"column:content;type:varchar(2048)" json:"content"`
	Liked      uint64    `gorm:"column:liked;type:int unsigned" json:"liked"`                                             // 点赞数量
	Comments   uint64    `gorm:"column:comments;type:int unsigned" json:"comments"`                                       // 评论数量
	CreateTime time.Time `gorm:"column:create_time;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"create_time"` // 创建时间
	UpdateTime time.Time `gorm:"column:update_time;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"update_time"` // 更新时间
}

// TableName TbBlog's table name
func (*TbBlog) TableName() string {
	return TableNameTbBlog
}
