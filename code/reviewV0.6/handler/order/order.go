package order

import (
	"context"
	"errors"
	"log/slog"
	"review/dal/model"
	"review/dal/query"
	"review/db"
	"review/pkg/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redsync/redsync/v4"
	"gorm.io/gorm"
)

const (
	coutBits          = 32
	incr              = "incr:"
	seckillUserKeyPre = "seckill:user:"
)

// 生成分布式id, 这个是订单id。不是优惠卷id
func NextId(keyPrefix string) int64 {
	now := time.Now().Unix()
	//生成序列号
	//Go语言的时间格式是通过一个特定的参考时间来定义的，这个参考时间是Mon Jan 2 15:04:05 MST 2006
	date := time.Now().Format("2006:01:01") //要用2006才能确保时间格式化正确

	count, err := db.RedisDb.Incr(context.Background(), incr+keyPrefix+":"+date).Result()
	if err != nil {
		slog.Error("Incr bad", "err", err)
		return -1
	}
	//拼接并返回
	return now<<coutBits | count
}

// post /api/v1/seckill/vouchers
func SeckillVoucher(c *gin.Context) {
	var req seckillResquest
	err := c.BindJSON(&req)
	if err != nil {
		slog.Error("bind json bad", "err", err)
		response.Error(c, response.ErrBind)
		return
	}

	//1.查询该优惠卷
	seckill := query.TbSeckillVoucher
	voucher, err := seckill.Where(seckill.VoucherID.Eq(uint64(req.VoucherId))).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, response.ErrNotFound, "秒杀卷不存在")
			return
		}
		response.Error(c, response.ErrDatabase)
		return
	}

	//2.判断秒杀卷是否合法，开始结束时间,库存
	now := time.Now()
	if voucher.BeginTime.After(now) || voucher.EndTime.Before(now) {
		response.Error(c, response.ErrValidation, "不在秒杀时间范围内")
		return
	}
	if voucher.Stock < 1 {
		response.Error(c, response.ErrValidation, "秒杀卷已被抢空")
		return
	}

	//使用锁
	// UserLockMap.Lock(req.UserId)
	// orderId, err := createVoucherOrder(req)
	// UserLockMap.Unlock(req.UserId)

	//使用redids分布式锁
	// lock := redislock.NewRedisLock(db.RedisDb, seckillUserKeyPre+strconv.Itoa(req.UserId), 0)
	// err = lock.Lock()
	// if err != nil { //获取锁失败
	// 	if errors.Is(err, redislock.ErrLockAlreadyHeld) {
	// 		// 锁被占用，用户操作过于频繁
	// 		response.Error(c, response.ErrValidation, "操作过于频繁，请稍后重试")
	// 		return
	// 	}
	// 	slog.Error("redis lock failed", "err", err, "userId", req.UserId)
	// 	response.Error(c, response.ErrDatabase, "系统繁忙，请稍后重试")
	// 	return
	// }
	// orderId, err := createVoucherOrder(req)
	// // 立即释放锁，不影响后续的HTTP响应处理
	// if unlockErr := lock.Unlock(); unlockErr != nil {
	// 	slog.Error("unlock failed", "err", unlockErr, "userId", req.UserId)
	// }

	//使用redsync进行加锁
	mutex := db.Rs.NewMutex(seckillUserKeyPre+strconv.Itoa(req.UserId), redsync.WithTries(1), redsync.WithExpiry(10*time.Second))
	if err = mutex.Lock(); err != nil {
		slog.Warn("failed to acquire lock", "err", err, "userId", req.UserId, "voucherId", req.VoucherId)
		// 根据错误类型给用户不同的响应
		if errors.Is(err, redsync.ErrFailed) {
			response.Error(c, response.ErrValidation, "操作过于频繁，请稍后重试")
		} else {
			response.Error(c, response.ErrDatabase, "系统繁忙，请稍后重试")
		}
		return
	}
	orderId, err := createVoucherOrder(req)
	if unlockOk, unlockErr := mutex.Unlock(); !unlockOk || unlockErr != nil {
		slog.Error("unlock failed", "unlockOk", unlockOk, "err", unlockErr, "userId", req.UserId)
	}

	if err != nil {
		response.HandleBusinessError(c, err)
		return
	}
	response.Success(c, gin.H{"orderId": orderId})
}

func createVoucherOrder(req seckillResquest) (int64, error) {
	//添加 判断是否是第一单
	VoucherOrder := query.TbVoucherOrder
	_, err := VoucherOrder.Where(VoucherOrder.VoucherID.Eq(uint64(req.VoucherId)), VoucherOrder.UserID.Eq(uint64(req.UserId))).First()
	if err == nil {
		return 0, response.WrapBusinessError(response.ErrValidation, nil, "当前用户已购买过该优惠卷")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 数据库查询出错
		return 0, response.WrapBusinessError(response.ErrDatabase, err, "查询用户订单失败")
	}

	order := model.TbVoucherOrder{
		ID:        NextId("order"),
		VoucherID: uint64(req.VoucherId),
		UserID:    uint64(req.UserId),
	}

	//处理两张表(订单表，秒杀卷表)，使用事务
	q := query.Use(db.DBEngine)
	err = q.Transaction(func(tx *query.Query) error {
		//3.合法，库存数量减1

		//3.合法，进行。库存数量减1
		// info, err := tx.TbSeckillVoucher.Where(tx.TbSeckillVoucher.VoucherID.Eq(uint64(req.VoucherId))).UpdateSimple(tx.TbSeckillVoucher.Stock.Add(-1))
		//每次都需要判断之前查询到的库存是否和现在的一致
		// info, err := tx.TbSeckillVoucher.Where(tx.TbSeckillVoucher.VoucherID.Eq(uint64(req.VoucherId)),
		// 	tx.TbSeckillVoucher.Stock.Eq(voucher.Stock)).UpdateSimple(tx.TbSeckillVoucher.Stock.Add(-1))
		//只需要判断是否>0即可
		info, err := tx.TbSeckillVoucher.Where(tx.TbSeckillVoucher.VoucherID.Eq(uint64(req.VoucherId)), tx.TbSeckillVoucher.Stock.Gt(0)).UpdateSimple(tx.TbSeckillVoucher.Stock.Add(-1))
		if err != nil {
			return response.WrapBusinessError(response.ErrDatabase, err, "")
		}
		if info.RowsAffected == 0 {
			slog.Warn("库存扣减失败", "voucherID", req.VoucherId, "reason", "库存不足或券不存在")
			return response.WrapBusinessError(response.ErrValidation, nil, "秒杀卷已被抢空")
		}

		//4.成功，创建对应的订单，并保存到数据中
		// err = tx.TbVoucherOrder.Create(&order)
		//出现问题Error 1292 (22007): Incorrect datetime value: '0000-00-00' for column 'pay_time' at row 1
		//表 `tb_voucher_order` 的字段`pay_time`,`use_time`,`refund_time`类型是timestamp，不允许插入'00000-00-00 00:00:00',数据库不接受这种无效的日期时间值。
		//可以指定更新需要的字段，不更新其他字段
		// err = tx.TbVoucherOrder.Select(tx.TbVoucherOrder.ID, tx.TbVoucherOrder.UserID, tx.TbVoucherOrder.VoucherID).Create(&order)
		//也可以这样写
		err = tx.TbVoucherOrder.Omit(tx.TbVoucherOrder.PayTime, tx.TbVoucherOrder.UseTime, tx.TbVoucherOrder.RefundTime).Create(&order)
		if err != nil {
			return response.WrapBusinessError(response.ErrDatabase, err, "")
		}
		return nil
	})
	return order.ID, err
}
