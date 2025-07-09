package shopservice

import (
	"review/dal/model"
	"strings"
)

type shopRequest struct {
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
func (r *shopRequest) ToModel() *model.TbShop {
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

type shopLoationDTO struct {
	ID        uint64   `json:"id"`         // 主键
	Name      string   `json:"name"`       // 商铺名称
	TypeID    uint64   `json:"type_id"`    // 商铺类型的id
	Images    []string `json:"images"`     // 商铺图片
	Area      string   `json:"area"`       // 商圈，例如陆家嘴
	Address   string   `json:"address"`    // 地址
	X         float64  `json:"x"`          // 经度
	Y         float64  `json:"y"`          // 维度
	AvgPrice  uint64   `json:"avg_price"`  // 均价，取整数
	Sold      uint32   `json:"sold"`       // 销量
	Comments  uint32   `json:"comments"`   // 评论数量
	Score     uint32   `json:"score"`      // 评分，1~5分，乘10保存，避免小数
	OpenHours string   `json:"open_hours"` // 营业时间，例如 10:00-22:00
	Distance  float64  `json:"distance"`   // 距离，单位米
}

func convertTbShopToResponse(shops []model.TbShop) []shopLoationDTO {
	apiShops := make([]shopLoationDTO, len(shops))
	for i, val := range shops {
		var images []string
		if val.Images != "" {
			images = strings.Split(val.Images, ",")
		}

		apiShops[i] = shopLoationDTO{
			ID:        val.ID,
			Name:      val.Name,
			TypeID:    val.TypeID,
			Images:    images,
			Area:      val.Area,
			Address:   val.Address,
			X:         val.X,
			Y:         val.Y,
			AvgPrice:  val.AvgPrice,
			Sold:      val.Sold,
			Comments:  val.Comments,
			Score:     val.Score,
			OpenHours: val.OpenHours,
		}
	}
	return apiShops
}

func convertTbShopPtrToResponse(shops []*model.TbShop) []shopLoationDTO {
	apiShops := make([]shopLoationDTO, len(shops))
	for i, val := range shops {
		var images []string
		if val.Images != "" {
			images = strings.Split(val.Images, ",")
		}

		apiShops[i] = shopLoationDTO{
			ID:        val.ID,
			Name:      val.Name,
			TypeID:    val.TypeID,
			Images:    images,
			Area:      val.Area,
			Address:   val.Address,
			X:         val.X,
			Y:         val.Y,
			AvgPrice:  val.AvgPrice,
			Sold:      val.Sold,
			Comments:  val.Comments,
			Score:     val.Score,
			OpenHours: val.OpenHours,
		}
	}
	return apiShops
}

type shopQueryParams struct {
	TypeID      int     `json:"type_id"`
	CurrentPage int     `json:"current_page"`
	Longitude   float64 `json:"longitude,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Distance    float64 `json:"distance,omitempty"`
	UseGeoQuery bool    `json:"use_geo_query"`
}
