package db

import (
	"context"
	"review/config"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

var RedisDb *redis.Client

var Rs *redsync.Redsync

func NewRedisClient(configRedis *config.RedisSetting) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     configRedis.Host,     //自己的redis实例的ip和port
		Password: configRedis.Password, //密码，有设置的话，就需要填写
		PoolSize: configRedis.PoolSize, //最大的可连接数量
	})
	_, err := client.Ping(context.Background()).Result() //测试ping
	if err != nil {
		return nil, err
	}

	// 创建redsync的客户端连接池
	pool := goredis.NewPool(client)
	// 创建redsync实例
	Rs = redsync.New(pool)

	// fmt.Println("redis测试: ", val)
	return client, err
}
