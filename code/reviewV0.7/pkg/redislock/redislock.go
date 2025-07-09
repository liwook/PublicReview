package redislock

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	defaultExpireTime = 5 * time.Second
)

var (
	ErrLockAlreadyHeld = errors.New("lock is already held by another client")
	ErrLockTimeout     = errors.New("failed to acquire lock within timeout")
	ErrLockNotFound    = errors.New("lock not found or already released")
)

type RedisLock struct {
	key      string
	expire   time.Duration
	redisCli *redis.Client
	Id       string //锁的标识，新添加的，也即是键的value
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
		redisCli: cli,
		Id:       id,
	}
}

func (lock *RedisLock) Lock() error {
	// success, err := lock.redisCli.SetNX(context.Background(), lock.key, "111111", lock.expire).Result()
	success, err := lock.redisCli.SetNX(context.Background(), lock.key, lock.Id, lock.expire).Result()
	if err != nil {
		return fmt.Errorf("redis operation failed: %w", err)
	}
	if !success {
		return ErrLockAlreadyHeld
	}
	return nil
}

// // 解锁，锁的误删除实现
// func (lock *RedisLock) Unlock() error {
// 	//获取锁并进行判断该锁是否是自己的
// 	val, err := lock.redisCli.Get(context.Background(), lock.key).Result()
// 	if err != nil {
// 		if errors.Is(err, redis.Nil) {
// 			// 锁不存在，可能已经过期或被释放
// 			return ErrLockNotFound
// 		}
// 		return fmt.Errorf("failed to get lock value: %w", err)
// 	}
// 	if val != lock.Id {
// 		return fmt.Errorf("lock not held by this client")
// 	}

// 	//进行删除锁
// 	res, err := lock.redisCli.Del(context.Background(), lock.key).Result()
// 	if err != nil {
// 		return fmt.Errorf("failed to delete lock: %w", err)
// 	}
// 	if res != 1 {
// 		// 锁在检查后可能被其他客户端删除或过期
// 		return ErrLockNotFound
// 	}
// 	return nil
// }

func (lock *RedisLock) Unlock() error {
	res, err := unlockScript.Run(context.Background(), lock.redisCli, []string{lock.key}, lock.Id).Int64()
	if err != nil {
		return fmt.Errorf("failed to execute unlock script: %w", err)
	}
	if res != 1 {
		return ErrLockNotFound
	}
	return nil
}
