// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"
)

const TableNameTbVoucher = "tb_voucher"

// TbVoucher mapped from table <tb_voucher>
type TbVoucher struct {
	ID          uint64    `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement:true" json:"id"`                  // 主键
	ShopID      uint64    `gorm:"column:shop_id;type:bigint unsigned" json:"shop_id"`                                      // 商铺id
	Title       string    `gorm:"column:title;type:varchar(255);not null" json:"title"`                                    // 代金券标题
	SubTitle    string    `gorm:"column:sub_title;type:varchar(255)" json:"sub_title"`                                     // 副标题
	Rules       string    `gorm:"column:rules;type:varchar(1024)" json:"rules"`                                            // 使用规则
	PayValue    uint64    `gorm:"column:pay_value;type:bigint unsigned;not null" json:"pay_value"`                         // 支付金额，单位是分。例如200代表2元
	ActualValue int64     `gorm:"column:actual_value;type:bigint;not null" json:"actual_value"`                            // 抵扣金额，单位是分。例如200代表2元
	Type        uint8     `gorm:"column:type;type:tinyint unsigned;not null" json:"type"`                                  // 0,普通券；1,秒杀券
	Status      uint8     `gorm:"column:status;type:tinyint unsigned;not null;default:1" json:"status"`                    // 1,上架; 2,下架; 3,过期
	CreateTime  time.Time `gorm:"column:create_time;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"create_time"` // 创建时间
	UpdateTime  time.Time `gorm:"column:update_time;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"update_time"` // 更新时间
}

// TableName TbVoucher's table name
func (*TbVoucher) TableName() string {
	return TableNameTbVoucher
}
