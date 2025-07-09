package redislock

import "github.com/redis/go-redis/v9"

const (
	luaCheckAndDelete = `
		if(redis.call('get',KEYS[1])==ARGV[1]) then
		return redis.call('del',KEYS[1])
		else
		return 0
		end
	`
)

// 预编译的解锁脚本
var unlockScript = redis.NewScript(luaCheckAndDelete)
