package db

import (
	"context"
	"review/config"

	"github.com/redis/go-redis/v9"
)

var RedisDb *redis.Client

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
	// fmt.Println("redis测试: ", val)
	return client, err
}
