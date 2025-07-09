package shopservice

import (
	"review/dal/model"
)

type ShopRequest struct {
	ID        uint64  `json:"id" binding:"omitempty"`                 // 商铺ID，必须提供
	Name      string  `json:"name" binding:"omitempty,min=2,max=128"` // 商铺名称
	TypeID    uint64  `json:"type_id" binding:"omitempty"`            // 商铺类型ID
	Images    string  `json:"images" binding:"omitempty"`             // 商铺图片，多张以逗号分隔
	Area      string  `json:"area" binding:"omitempty,max=128"`       // 商圈
	Address   string  `json:"address" binding:"omitempty,max=255"`    // 地址
	X         float64 `json:"x" binding:"omitempty"`                  // 经度
	Y         float64 `json:"y" binding:"omitempty"`                  // 纬度
	AvgPrice  uint64  `json:"avg_price" binding:"omitempty"`          // 均价
	OpenHours string  `json:"open_hours" binding:"omitempty,max=32"`  // 营业时间
}

// ToModel 将DTO转换为数据库模型对象
func (r *ShopRequest) ToModel() *model.TbShop {
	return &model.TbShop{
		ID:        r.ID,
		Name:      r.Name,
		TypeID:    r.TypeID,
		Images:    r.Images,
		Area:      r.Area,
		Address:   r.Address,
		X:         r.X,
		Y:         r.Y,
		AvgPrice:  r.AvgPrice,
		OpenHours: r.OpenHours,
	}
}
