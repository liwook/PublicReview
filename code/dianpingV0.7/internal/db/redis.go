package db

import (
	"context"
	"dianping/internal/config"
	"fmt"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var Rs *redsync.Redsync

func NewRedisClient(config *config.RedisSetting) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Host,     //自己的redis实例的ip和port
		Password: config.Password, //密码，有设置的话，就需要填写
		PoolSize: config.PoolSize, //最大的可连接数量
	})
	val, err := client.Ping(context.Background()).Result() //测试ping
	if err != nil {
		return nil, err
	}
	fmt.Println("redis测试: ", val)

	// 创建redsync的客户端连接池
	pool := goredis.NewPool(client)
	// 创建redsync实例
	Rs = redsync.New(pool)

	return client, err
}
