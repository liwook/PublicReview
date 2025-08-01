// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"
)

const TableNameTbUser = "tb_user"

// TbUser mapped from table <tb_user>
type TbUser struct {
	ID         uint64    `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement:true;comment:主键" json:"id"`                    // 主键
	Phone      string    `gorm:"column:phone;type:varchar(11);not null;comment:手机号码" json:"phone"`                                     // 手机号码
	Password   string    `gorm:"column:password;type:varchar(128);comment:密码，加密存储" json:"password"`                                    // 密码，加密存储
	NickName   string    `gorm:"column:nick_name;type:varchar(32);comment:昵称，默认是用户id" json:"nick_name"`                                // 昵称，默认是用户id
	Icon       string    `gorm:"column:icon;type:varchar(255);comment:人物头像" json:"icon"`                                               // 人物头像
	CreateTime time.Time `gorm:"column:create_time;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"create_time"` // 创建时间
	UpdateTime time.Time `gorm:"column:update_time;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"update_time"` // 更新时间
}

// TableName TbUser's table name
func (*TbUser) TableName() string {
	return TableNameTbUser
}
