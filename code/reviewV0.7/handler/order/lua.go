package order

import "github.com/redis/go-redis/v9"

const adjustSeckill = `
-- 秒杀优化需求二:基于Lua脚本,判断秒杀库存、一人一单,决定用户是否有购买资格
-- 1.参数列表
-- 1.1.优惠券id
local voucherId = KEYS[1]
-- 1.2.用户id
local userId = KEYS[2]
-- 1.3.订单id
local orderId = KEYS[3]

-- 2.数据key
-- 2.1.库存key  ..lua的字符串拼接
local stockKey = 'seckill:stock:' .. voucherId
-- 2.2 订单key
local orderKey = 'seckill:order:' .. voucherId;

-- 3.脚本业务
-- 3.1.判断库存是否充足 get stockKey  tonumber()将字符串转换为数字
local stock = redis.call('get', stockKey)
if(stock == false or tonumber(stock) <= 0) then
    -- 3.2.库存不足,返回1
    return 1
end

-- 3.2.判断用户是否下单 SISMEMBER:判断set集合中是否存在某个元素
if(redis.call('sismember', orderKey, userId) == 1) then
    -- 3.3.存在，说明是重复下单,返回2
    return 2
end
-- 3.4.扣库存 incrby stockKey -1
  redis.call('incrby', stockKey, -1)
-- 3.5.下单(保存用户) sadd orderKey userId
  redis.call('sadd', orderKey, userId)

-- 3.6.发送消息到队列中， XADD stream.orders * k1 v1 k2 v2 ...
redis.call('xadd', 'stream.orders', '*', 'userId', userId, 'voucherId', voucherId, 'id', orderId)

return 0
`

var adjustSeckillScript = redis.NewScript(adjustSeckill)
