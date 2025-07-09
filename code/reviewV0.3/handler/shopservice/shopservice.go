package shopservice

import (
	"context"
	"errors"
	"log/slog"
	"review/dal/model"
	"review/dal/query"
	"review/db"
	"review/pkg/consts"
	"review/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	shopKeyPrefix    = "cache:shop:"
	shopCacheTTL     = 30 * 60 // 30分钟，单位：秒
	shopTypeKey      = "cache:shopType"
	shopTypeCacheTTL = 60 * 60 // 1小时，商店类型变化较少
)

// 根据商店id查找商店缓存数据
// get /api/v1/shop/:shopId
func QueryShopById(c *gin.Context) {
	// 1. 参数验证
	id := c.Param(consts.ShopIdKey)
	if id == "" {
		response.Error(c, response.ErrValidation, "id can not be empty")
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil || idInt <= 0 {
		response.Error(c, response.ErrValidation, "invalid shop id")
		return
	}

	//2. 从缓存中查找
	shop, err := getShopFromCache(id)
	if err == nil {
		response.Success(c, shop)
		return
	}

	//3.缓存未命中或者出错，从数据库中查找
	if !errors.Is(err, redis.Nil) {
		slog.Error("redis get error", "error", err) //记录错误
		//继续从数据库中查找，不直接返回错误
	}

	shop, err = getShopFromDB(idInt)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, response.ErrNotFound, "shop not found")
		return
	}
	if err != nil {
		slog.Error("mysql find shop by id failed", "error", err)
		response.Error(c, response.ErrDatabase)
		return
	}

	//4. 将数据写入缓存,可异步写入
	go setShopToCache(id, shop)

	//5. 返回结果
	response.Success(c, shop)
}

// 返回商铺类型的数据，给首页
// get /api/v1/shop/type-list
func QueryShopTypeList(c *gin.Context) {
	//1.尝试从缓存获取
	shopTypes, err := getShopTypeFromCache()
	if err == nil && len(shopTypes) > 0 {
		response.Success(c, shopTypes)
		return
	}
	if err != nil {
		slog.Error("redis get shop tyes error", "error", err) //只记录错误
	}

	//2.缓存未命中，从数据库中查找
	shopTypes, err = getShopTypesFromDB()
	if err != nil {
		slog.Error("mysql find shop types failed", "error", err)
		response.Error(c, response.ErrDatabase)
		return
	}
	if len(shopTypes) == 0 {
		response.Success(c, []model.TbShopType{})
		return
	}

	//3.将数据写入缓存，可异步写入
	go setShopTypeToCache(shopTypes)

	//4.返回结果
	response.Success(c, shopTypes)
}

// PUT /api/v1/shop
func UpdateShop(c *gin.Context) {
	var shop ShopRequest
	err := c.ShouldBindJSON(&shop)
	if err != nil {
		slog.Error("bindjson bad", "err", err)
		response.Error(c, response.ErrBind)
		return
	}

	// 更新时必须提供有效的ID
	if shop.ID == 0 {
		response.Error(c, response.ErrValidation, "shop id is required for update")
		return
	}
	update(c, &shop)
}

func update(c *gin.Context, shop *ShopRequest) {
	//1.更新数据库
	//当通过 struct 更新时，GORM 只会更新非零字段。
	//若想确保指定字段被更新,应使用Select更新选定字段，或使用map来完成更新
	data := shop.ToModel()
	tbshop := query.TbShop
	_, err := tbshop.Where(tbshop.ID.Eq(shop.ID)).Updates(data)
	if err != nil {
		slog.Error("update mysql bad", "err", err)
		response.Error(c, response.ErrDatabase)
		return
	}

	//2.删除缓存
	key := shopKeyPrefix + strconv.Itoa(int(shop.ID))
	db.RedisDb.Del(context.Background(), key)

	response.Success(c, nil)
}

// 添加商铺
// post /api/v1/shop
func AddShop(c *gin.Context) {
	var shop ShopRequest
	err := c.ShouldBindJSON(&shop)
	if err != nil {
		slog.Error("bindjson bad", "err", err)
		response.Error(c, response.ErrBind)
		return
	}

	// 添加时忽略ID字段
	shop.ID = 0

	data := shop.ToModel()
	err = query.TbShop.Create(data)
	if err != nil {
		slog.Error("mysql create shop err", "err", err)
		response.Error(c, response.ErrDatabase)
		return
	}

	response.Success(c, gin.H{"id": data.ID})
}

// 删除商铺
// delete /api/v1/shop/:shopId
func DelShop(c *gin.Context) {
	id := c.Param(consts.ShopIdKey)
	if id == "" {
		response.Error(c, response.ErrValidation, "id is null")
		return
	}

	val, err := strconv.ParseInt(id, 10, 64)
	if err != nil || val <= 0 {
		response.Error(c, response.ErrValidation, "invalid shop id")
		return
	}
	shop := query.TbShop
	result, err := shop.Where(shop.ID.Eq(uint64(val))).Delete()
	if err != nil {
		response.Error(c, response.ErrDatabase)
	}
	if result.RowsAffected == 0 {
		response.Error(c, response.ErrNotFound, "shop not found")
		return
	}
	//删除缓存
	key := shopKeyPrefix + id
	_, err = db.RedisDb.Del(context.Background(), key).Result()
	if err != nil {
		slog.Error("redis delete shop error", "error", err)
	}

	response.Success(c, nil)
}
