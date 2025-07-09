package shopservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"review/dal/model"
	"review/dal/query"
	"review/db"
	"review/pkg/response"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var sg singleflight.Group

// 商铺的数据
func getShopFromCache(id string) (*model.TbShop, error) {
	val, err := db.RedisDb.Get(context.Background(), shopKeyPrefix+id).Result()
	if err != nil {
		return nil, err
	}
	//可以存空值
	if val == "" { //因为传入的是nil的结构体变量的，所以
		return nil, nil
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
	// data, err := sonic.Marshal(shop)
	// if err != nil {
	// 	slog.Error("sonic marshal error", "error", err)
	// 	return err
	// }

	var data []byte //无论是nil切片还是[]byte("")，在Redis中都会被存储为空字符串
	if shop != nil {
		var err error
		data, err = sonic.Marshal(shop)
		if err != nil {
			slog.Error("sonic marshal error", "error", err)
			return err
		}
	}
	//添加随机的ttl，解决缓存雪崩
	err := db.RedisDb.Set(context.Background(), shopKeyPrefix+id, data, shopCacheTTL*time.Second+time.Duration(rand.Int31n(10000))).Err()
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

func singleflightFunc(id string) (any, error) {
	//2. 从缓存中查找
	shop, err := getShopFromCache(id)
	if err == nil {
		//现在使用可以存储空值的了
		if shop == nil {
			fmt.Println("shop is nil")
			return nil, response.NewBusinessError(response.ErrNotFound, "shop not found")
		}

		return shop, nil
	}

	//3.缓存未命中或者出错，从数据库中查找
	if !errors.Is(err, redis.Nil) {
		slog.Error("redis get error", "error", err) //记录错误
		//继续从数据库中查找，不直接返回错误
	}

	idInt, _ := strconv.Atoi(id)
	shop, err = getShopFromDB(idInt)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		setShopToCache(id, nil)
		return nil, response.NewBusinessError(response.ErrNotFound, "shop not found")
	}
	if err != nil {
		return nil, response.WrapBusinessError(response.ErrDatabase, err, "")
	}

	//4. 将数据写入缓存,可异步写入
	go setShopToCache(id, shop)
	return shop, nil
}
