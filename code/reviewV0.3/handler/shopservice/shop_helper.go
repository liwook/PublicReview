package shopservice

import (
	"context"
	"log/slog"
	"review/dal/model"
	"review/dal/query"
	"review/db"
	"time"

	"github.com/bytedance/sonic"
)

// 商铺的数据
func getShopFromCache(id string) (*model.TbShop, error) {
	val, err := db.RedisDb.Get(context.Background(), shopKeyPrefix+id).Result()
	if err != nil {
		return nil, err
	}

	var shop model.TbShop
	err = sonic.Unmarshal([]byte(val), &shop)
	if err != nil {
		return nil, err
	}
	return &shop, nil
}

func getShopFromDB(id int) (*model.TbShop, error) {
	shopQuery := query.TbShop
	return shopQuery.Where(shopQuery.ID.Eq(uint64(id))).First()
}

func setShopToCache(id string, shop *model.TbShop) error {
	data, err := sonic.Marshal(shop)
	if err != nil {
		slog.Error("sonic marshal error", "error", err)
		return err
	}

	err = db.RedisDb.Set(context.Background(), shopKeyPrefix+id, data, shopCacheTTL*time.Second).Err()
	if err != nil {
		slog.Error("redis set error", "error", err)
		return err
	}
	return nil
}

// 商铺类型的数据
func getShopTypeFromCache() ([]*model.TbShopType, error) {
	val, err := db.RedisDb.LRange(context.Background(), shopTypeKey, 0, -1).Result()
	if err != nil || len(val) == 0 {
		return nil, err
	}

	shopTypes := make([]*model.TbShopType, 0, len(val))
	for _, v := range val {
		var shopType model.TbShopType
		err = sonic.UnmarshalString(v, &shopType)
		if err != nil {
			slog.Error("sonic unmarshal error", "error", err)
			continue
		}
		shopTypes = append(shopTypes, &shopType)
	}

	return shopTypes, nil
}

func getShopTypesFromDB() ([]*model.TbShopType, error) {
	shopType := query.TbShopType
	return shopType.Order(shopType.Sort).Find()
}

func setShopTypeToCache(shopTypes []*model.TbShopType) {
	pipeline := db.RedisDb.Pipeline()
	//删除原来的数据
	pipeline.Del(context.Background(), shopTypeKey)

	for _, shopType := range shopTypes {
		val, err := sonic.Marshal(shopType)
		if err != nil {
			slog.Error("sonic marshal error", "error", err)
			continue
		}
		pipeline.RPush(context.Background(), shopTypeKey, string(val))
	}

	pipeline.Expire(context.Background(), shopTypeKey, shopTypeCacheTTL*time.Second)
	_, err := pipeline.Exec(context.Background())
	if err != nil {
		slog.Error("redis pipeline exec failed", "error", err)
	}
}
