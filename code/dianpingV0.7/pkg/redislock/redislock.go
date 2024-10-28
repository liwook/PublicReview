package redislock

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	defaultExpireTime = 5 * time.Second
)

type RedisLock struct {
	key      string
	expire   time.Duration
	Id       string //锁的标识，新添加的，也即是键的value
	redisCli *redis.Client
}

// expire: 锁的过期时间,为0则使用默认过期时间
func NewRedisLock(cli *redis.Client, key string, expire time.Duration) *RedisLock {
	if expire == 0 {
		expire = defaultExpireTime
	}
	id := strings.Join(strings.Split(uuid.New().String(), "-"), "")
	return &RedisLock{
		key:      key,
		expire:   expire,
		Id:       id,
		redisCli: cli,
	}
}

// 加锁, 设置了键的value
func (lock *RedisLock) Lock() (bool, error) {
	return lock.redisCli.SetNX(context.Background(), lock.key, lock.Id, lock.expire).Result()
}

// // lock.redisCli.Del(lock.key)对Redis中的lock.key进行删除.当删除后，竞争者才有机会对该键进行 SETNX。
// func (lock *RedisLock) Unlock() error {
// 	res, err := lock.redisCli.Del(context.Background(), lock.key).Result()
// 	if err != nil {
// 		return err
// 	}
// 	if res != 1 {
// 		return fmt.Errorf("can not unlock because del result not is one")
// 	}
// 	return nil
// }

// // 解锁，锁的误删除实现
// func (lock *RedisLock) Unlock() error {
// 	//获取锁并进行判断该锁是否是自己的
// 	val, err := lock.redisCli.Get(context.Background(), lock.key).Result()
// 	if err != nil {
// 		fmt.Println("lock not exit")
// 		return err
// 	}
// 	if val == "" || val != lock.Id {
// 		return fmt.Errorf("lock not belong to myself")
// 	}

// 	//进行删除锁
// 	res, err := lock.redisCli.Del(context.Background(), lock.key).Result()
// 	if err != nil {
// 		return err
// 	}
// 	if res != 1 {
// 		return fmt.Errorf("can not unlock because del result not is one")
// 	}
// 	return nil
// }

func (lock *RedisLock) Unlock() error {
	script := redis.NewScript(LauCheckAndDelete)
	res, err := script.Run(context.Background(), lock.redisCli, []string{lock.key}, lock.Id).Int64()
	if err != nil {
		return err
	}
	if res != 1 {
		return fmt.Errorf("can not unlock because del result not is one")
	}
	return nil
}
