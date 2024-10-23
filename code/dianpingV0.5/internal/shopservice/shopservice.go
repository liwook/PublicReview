package shopservice

import (
	"context"
	"dianping/dal/model"
	"dianping/dal/query"
	"dianping/internal/db"
	"dianping/pkg/code"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/exp/rand"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

const (
	ShopKeyPriex = "cache:shop:"
	ShopTypeKey  = "cache:shopType"
	CacheNullTTL = 10 * time.Minute
	CacheShopTTL = 2 * time.Hour
)

var sg singleflight.Group

// 根据商店id查找商店缓存数据
// get /shop/:id
// func QueryShopById(c *gin.Context) {
// 	id := c.Param("id") //获取定义的路由参数的值
// 	if id == "" {
// 		code.WriteResponse(c, code.ErrValidation, "id cannot be empty")
// 		return
// 	}

// 	//1.从redis查询商铺缓存，是string类型的
// 	val, err := db.RedisClient.Get(context.Background(), ShopKeyPriex+id).Result()
// 	if err == nil { //若redis存在该缓存，直接返回
// 		var shop model.TbShop
// 		sonic.UnmarshalString(val, &shop)
// 		code.WriteResponse(c, code.ErrSuccess, shop)
// 	} else if err == redis.Nil { //2.若是redis没有该缓存，从mysql中查询
// 		tbSop := query.TbShop
// 		idInt, _ := strconv.Atoi(id)
// 		shop, err := tbSop.Where(tbSop.ID.Eq(uint64(idInt))).First()
// 		if err == gorm.ErrRecordNotFound {
// 			//3.mysql若不存在该商铺，返回错误
// 			code.WriteResponse(c, code.ErrDatabase, "this shop not found")
// 			return
// 		}
// 		if err != nil {
// 			slog.Error("mysql find shop by id bad", "error", err)
// 			code.WriteResponse(c, code.ErrDatabase, nil)
// 			return
// 		}

// 		//4.找到商铺，写回redis,并发送给客户端
// 		//把shop进行序列化，不然写入redis会出错。序列化就是把该数据对象变成json，即是变成一个字符串
// 		v, _ := sonic.Marshal(shop) //这里使用github.com/bytedance/sonic
// 		_, err = db.RedisClient.Set(context.Background(), ShopKeyPriex+id, v, 0).Result()
// 		if err != nil {
// 			slog.Error("redis set val bad", "error", err)
// 			code.WriteResponse(c, code.ErrDatabase, nil)
// 			return
// 		}
// 		code.WriteResponse(c, code.ErrSuccess, shop)
// 	} else { //若redis查询出错，返回错误
// 		slog.Error("redis get val bad", "error", err)
// 		code.WriteResponse(c, code.ErrDatabase, nil)
// 	}
// }

// 使用缓存空值来解决缓存穿透
// func QueryShopById(c *gin.Context) {
// 	id := c.Param("id") //获取定义的路由参数的值
// 	if id == "" {
// 		code.WriteResponse(c, code.ErrValidation, "id cannot be empty")
// 		return
// 	}

// 	//1.从redis查询商铺缓存，是string类型的
// 	val, err := db.RedisClient.Get(context.Background(), ShopKeyPriex+id).Result()
// 	if err == nil { //若redis存在该缓存，直接返回
// 		//有缓存，但可能是控制，需要判断
// 		if val == "" {
// 			code.WriteResponse(c, code.ErrDatabase, "this shop not found")
// 			return
// 		}
// 		var shop model.TbShop
// 		sonic.Unmarshal([]byte(val), &shop)
// 		code.WriteResponse(c, code.ErrSuccess, shop)
// 	} else if err == redis.Nil { //2.若是redis没有该缓存，从mysql中查询
// 		tbSop := query.TbShop
// 		idInt, _ := strconv.Atoi(id)
// 		shop, err := tbSop.Where(tbSop.ID.Eq(uint64(idInt))).First()
// 		if err == gorm.ErrRecordNotFound {
// 			//3.mysql若不存在该商铺，返回错误
// 			//需要往redis中写入空值，防止缓存穿透,并设置较短的过期时间
// 			_, _ = db.RedisClient.Set(context.Background(), ShopKeyPriex+id, "", CacheNullTTL).Result()
// 			code.WriteResponse(c, code.ErrDatabase, "this shop not found")
// 			return
// 		}
// 		if err != nil {
// 			slog.Error("mysql find shop by id bad", "error", err)
// 			code.WriteResponse(c, code.ErrDatabase, nil)
// 			return
// 		}

// 		//4.找到商铺，写回redis,并发送给客户端
// 		//把shop进行序列化，不然写入redis会出错。序列化就是把该数据对象变成json，即是变成一个字符串
// 		v, _ := sonic.Marshal(shop) //这里使用github.com/bytedance/sonic
// 		_, err = db.RedisClient.Set(context.Background(), ShopKeyPriex+id, v, 0).Result()
// 		if err != nil {
// 			slog.Error("redis set val bad", "error", err)
// 			code.WriteResponse(c, code.ErrDatabase, nil)
// 			return
// 		}
// 		code.WriteResponse(c, code.ErrSuccess, shop)
// 	} else { //若redis查询出错，返回错误
// 		slog.Error("redis get val bad", "error", err)
// 		code.WriteResponse(c, code.ErrDatabase, nil)
// 	}
// }

// // 使用ttl是随机值来解决缓存雪崩
// func QueryShopById(c *gin.Context) {
// 	id := c.Param("id") //获取定义的路由参数的值
// 	if id == "" {
// 		code.WriteResponse(c, code.ErrValidation, "id cannot be empty")
// 		return
// 	}

// 	//1.从redis查询商铺缓存，是string类型的
// 	val, err := db.RedisClient.Get(context.Background(), ShopKeyPriex+id).Result()
// 	if err == nil { //若redis存在该缓存，直接返回
// 		//有缓存，但可能是控制，需要判断
// 		if val == "" {
// 			code.WriteResponse(c, code.ErrDatabase, "this shop not found")
// 			return
// 		}
// 		var shop model.TbShop
// 		sonic.Unmarshal([]byte(val), &shop)
// 		code.WriteResponse(c, code.ErrSuccess, shop)
// 	} else if err == redis.Nil { //2.若是redis没有该缓存，从mysql中查询
// 		tbSop := query.TbShop
// 		idInt, _ := strconv.Atoi(id)
// 		shop, err := tbSop.Where(tbSop.ID.Eq(uint64(idInt))).First()
// 		if err == gorm.ErrRecordNotFound {
// 			//3.mysql若不存在该商铺，返回错误
// 			//需要往redis中写入空值，防止缓存穿透,并设置较短的过期时间
// 			_, _ = db.RedisClient.Set(context.Background(), ShopKeyPriex+id, "", CacheNullTTL).Result()
// 			code.WriteResponse(c, code.ErrDatabase, "this shop not found")
// 			return
// 		}
// 		if err != nil {
// 			slog.Error("mysql find shop by id bad", "error", err)
// 			code.WriteResponse(c, code.ErrDatabase, nil)
// 			return
// 		}

// 		//4.找到商铺，写回redis,并发送给客户端
// 		//把shop进行序列化，不然写入redis会出错。序列化就是把该数据对象变成json，即是变成一个字符串
// 		//添加随机的ttl，解决缓存雪崩
// 		v, _ := sonic.Marshal(shop) //这里使用github.com/bytedance/sonic
// 		_, err = db.RedisClient.Set(context.Background(), ShopKeyPriex+id, v, CacheShopTTL+time.Duration(rand.Int31n(10000))).Result()
// 		if err != nil {
// 			slog.Error("redis set val bad", "error", err)
// 			code.WriteResponse(c, code.ErrDatabase, nil)
// 			return
// 		}
// 		code.WriteResponse(c, code.ErrSuccess, shop)
// 	} else { //若redis查询出错，返回错误
// 		slog.Error("redis get val bad", "error", err)
// 		code.WriteResponse(c, code.ErrDatabase, nil)
// 	}
// }

// 使用singleflight解决缓存击穿
func QueryShopById(c *gin.Context) {
	id := c.Param("id") //获取定义的路由参数的值
	if id == "" {
		code.WriteResponse(c, code.ErrValidation, "id cannot be empty")
		return
	}

	val, err, _ := sg.Do(ShopKeyPriex+id, func() (interface{}, error) {
		return singleflightFunc(id)
	})

	if err != nil {
		if err.Error() == "this shop not found" {
			code.WriteResponse(c, code.ErrDatabase, "this shop not found")
			return
		}
		code.WriteResponse(c, code.ErrDatabase, nil)
	}
	code.WriteResponse(c, code.ErrSuccess, val)
}

var num = 0 //用于测试有多少次请求会执行该函数，即有多少次请求会查询redis
func singleflightFunc(id string) (any, error) {
	//1.从redis查询商铺缓存，是string类型的
	val, err := db.RedisClient.Get(context.Background(), ShopKeyPriex+id).Result()
	num++
	fmt.Println(num)
	if err == nil { //若redis存在该缓存，直接返回
		//有缓存，但可能是控制，需要判断
		if val == "" {
			return "", nil
		}
		var shop model.TbShop
		sonic.UnmarshalString(val, &shop)
		return shop, nil
	} else if err == redis.Nil { //2.若是redis没有该缓存，从mysql中查询
		tbSop := query.TbShop
		idInt, _ := strconv.Atoi(id)
		shop, err := tbSop.Where(tbSop.ID.Eq(uint64(idInt))).First()
		if err == gorm.ErrRecordNotFound {
			//3.mysql若不存在该商铺，返回错误
			//需要往redis中写入空值，防止缓存穿透,并设置较短的过期时间
			_, _ = db.RedisClient.Set(context.Background(), ShopKeyPriex+id, "", CacheNullTTL).Result()
			return "", fmt.Errorf("this shop not found")
		}
		if err != nil {
			slog.Error("mysql find shop by id bad", "error", err)
			return "", err
		}

		//4.找到商铺，写回redis,并发送给客户端
		//把shop进行序列化，不然写入redis会出错。序列化就是把该数据对象变成json，即是变成一个字符串
		//添加随机的ttl，解决缓存雪崩
		v, _ := sonic.Marshal(shop) //这里使用github.com/bytedance/sonic
		_, err = db.RedisClient.Set(context.Background(), ShopKeyPriex+id, v, CacheShopTTL+time.Duration(rand.Int31n(10000))).Result()
		if err != nil {
			slog.Error("redis set val bad", "error", err)
			return "", err
		}
		return val, nil
	} else { //若redis查询出错，返回错误
		slog.Error("redis get val bad", "error", err)
		return "", err
	}
}

// 返回商铺类型的数据，给首页
// get /shop/type-list
func QueryShopTypeList(c *gin.Context) {
	//1.先从redis中查询
	// 获取List中的元素：起始索引~结束索引，当结束索引 > llen(list)或=-1时，取出全部数据
	val, err := db.RedisClient.LRange(context.Background(), ShopTypeKey, 0, -1).Result()
	if err == redis.Nil || len(val) == 0 {
		//2. 若是没有,从mysql中获取
		shopType := query.TbShopType
		typeList, err := shopType.Order(shopType.Sort).Find() //Find函数返回没有数据的话,err是nil
		if err != nil {
			slog.Error("shoptypelist mysql find bad", "err", err)
			code.WriteResponse(c, code.ErrDatabase, nil)
			return
		}
		if len(typeList) == 0 {
			code.WriteResponse(c, code.ErrSuccess, "no data in database")
			return
		}

		//3.序列化,并往redis中添加
		//注意：要是使用[]byte，会报错redis: can't marshal [][]uint8，所以要转换成string
		pipeline := db.RedisClient.Pipeline()
		for _, shop := range typeList {
			val, _ := sonic.Marshal(shop)
			fmt.Println(string(val))
			pipeline.RPush(context.Background(), ShopTypeKey, string(val))
		}
		_, err = pipeline.Exec(context.Background())
		if err != nil {
			slog.Error("redis list push bad", "err", err)
			code.WriteResponse(c, code.ErrDatabase, nil)
			return
		}
		code.WriteResponse(c, code.ErrSuccess, typeList)
	} else if err != nil {
		slog.Error("redis list find bad", "err", err)
		code.WriteResponse(c, code.ErrDatabase, nil)
	} else {
		var valList = make([]*model.TbShopType, len(val))
		for i, v := range val {
			_ = sonic.UnmarshalString(v, &valList[i])
		}
		code.WriteResponse(c, code.ErrSuccess, valList)
	}
}

// 更新商铺
// post /shop/update
func UpdateShop(c *gin.Context) {
	var shop model.TbShop
	err := c.BindJSON(&shop)
	if err != nil {
		slog.Error("bindjson bad", "err", err)
		code.WriteResponse(c, code.ErrBind, nil)
		return
	}
	update(c, &shop)
}

func update(c *gin.Context, shop *model.TbShop) {
	//1.更新数据库
	//当通过 struct 更新时，GORM 只会更新非零字段。
	//若想确保指定字段被更新,应使用Select更新选定字段，或使用map来完成更新
	tbshop := query.TbShop
	_, err := tbshop.Where(tbshop.ID.Eq(shop.ID)).Updates(shop)
	if err != nil {
		slog.Error("update mysql bad", "err", err)
		code.WriteResponse(c, code.ErrDatabase, nil)
		return
	}

	//2.删除缓存
	key := ShopKeyPriex + strconv.Itoa(int(shop.ID))
	db.RedisClient.Del(context.Background(), key)

	code.WriteResponse(c, code.ErrSuccess, nil)
}

// 添加商铺
// post /shop/add
func AddShop(c *gin.Context) {
	var shop model.TbShop
	err := c.BindJSON(&shop)
	if err != nil {
		slog.Error("bindjson bad", "err", err)
		code.WriteResponse(c, code.ErrBind, nil)
		return
	}

	err = query.TbShop.Create(&shop)
	if err != nil {
		slog.Error("mysql create shop err", "err", err)
		code.WriteResponse(c, code.ErrDatabase, nil)
	} else {
		code.WriteResponse(c, code.ErrSuccess, nil)
	}
}

// 删除商铺
// delet /shop/delete/:id
func DelShop(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		code.WriteResponse(c, code.ErrValidation, "id is null")
		return
	}
	val, _ := strconv.Atoi(id)
	shop := query.TbShop
	_, err := shop.Where(shop.ID.Eq(uint64(val))).Delete()
	if err != nil {
		code.WriteResponse(c, code.ErrDatabase, nil)
	}

	//删除缓存
	key := ShopKeyPriex + id
	db.RedisClient.Del(context.Background(), key)

	code.WriteResponse(c, code.ErrSuccess, nil)
}
