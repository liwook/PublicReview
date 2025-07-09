package order

import (
	"context"
	"log/slog"
	"review/dal/model"
	"review/dal/query"
	"review/db"
	"review/pkg/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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

	orderId := NextId("order")
	// res, err := adjustSeckillScript.Run(context.Background(), db.RedisDb, []string{strconv.Itoa(req.VoucherId), strconv.Itoa(req.UserId)}).Result()
	res, err := adjustSeckillScript.Run(context.Background(), db.RedisDb, []string{strconv.Itoa(req.VoucherId), strconv.Itoa(req.UserId), strconv.Itoa(int(orderId))}).Result()
	if err != nil {
		slog.Error("run script bad", "err", err)
		response.Error(c, response.ErrDatabase)
		return
	}

	//需要注意，res的类型是interface{}，需要转换。
	if res.(int64) != 0 {
		// if res != 0 {
		var e string
		if res == 1 {
			e = "stock not enough"
		} else {
			e = "order already exist"
		}
		response.Error(c, response.ErrDatabase, e)
	}

	//判断一人一单已经在redis中完成，接下来是生成订单，所以不需要加锁了
	// orderId, err := createVoucherOrder(orderId, req)
	// go createVoucherOrder(orderId, req)

	response.Success(c, gin.H{"orderId": orderId})
}

// func createVoucherOrder(orderId int64, req seckillResquest) (int64, error) {
// 	order := model.TbVoucherOrder{
// 		// ID:        NextId("order"),
// 		ID:        orderId,
// 		VoucherID: uint64(req.VoucherId),
// 		UserID:    uint64(req.UserId),
// 	}

// 	//处理两张表(订单表，秒杀卷表)，使用事务
// 	q := query.Use(db.DBEngine)
// 	err := q.Transaction(func(tx *query.Query) error {
// 		//3.合法，库存数量减1
// 		info, err := tx.TbSeckillVoucher.Where(tx.TbSeckillVoucher.VoucherID.Eq(uint64(req.VoucherId)), tx.TbSeckillVoucher.Stock.Gt(0)).UpdateSimple(tx.TbSeckillVoucher.Stock.Add(-1))
// 		if err != nil {
// 			return response.WrapBusinessError(response.ErrDatabase, err, "")
// 		}
// 		if info.RowsAffected == 0 {
// 			slog.Warn("库存扣减失败", "voucherID", req.VoucherId, "reason", "库存不足或券不存在")
// 			return response.WrapBusinessError(response.ErrValidation, nil, "秒杀卷已被抢空")
// 		}

// 		err = tx.TbVoucherOrder.Omit(tx.TbVoucherOrder.PayTime, tx.TbVoucherOrder.UseTime, tx.TbVoucherOrder.RefundTime).Create(&order)
// 		if err != nil {
// 			return response.WrapBusinessError(response.ErrDatabase, err, "")
// 		}
// 		return nil
// 	})
// 	return order.ID, err
// }

func CreateOrder(voucherId int, userId int, orderId int) error {
	order := model.TbVoucherOrder{
		ID:        int64(orderId),
		VoucherID: uint64(voucherId),
		UserID:    uint64(userId),
	}

	//处理两张表(订单表，秒杀卷表)，使用事务
	q := query.Use(db.DBEngine)
	return q.Transaction(func(tx *query.Query) error {
		//3.合法，库存数量减1
		//使用update，要是没有该条数据，不会返回gorm.ErrRecordNotFound或者有错误的。
		info, err := tx.TbSeckillVoucher.Where(tx.TbSeckillVoucher.VoucherID.Eq(uint64(voucherId)), tx.TbSeckillVoucher.Stock.Gt(0)).UpdateSimple(tx.TbSeckillVoucher.Stock.Add(-1))
		if err != nil {
			return response.WrapBusinessError(response.ErrDatabase, err, "")
		}
		if info.RowsAffected == 0 {
			slog.Warn("库存扣减失败", "voucherID", voucherId, "reason", "库存不足或券不存在")
			return response.WrapBusinessError(response.ErrValidation, nil, "秒杀卷已被抢空")
		}

		//4.成功，创建对应的订单，并保存到数据中
		err = tx.TbVoucherOrder.Select(tx.TbVoucherOrder.ID, tx.TbVoucherOrder.UserID, tx.TbVoucherOrder.VoucherID).Create(&order)
		if err != nil {
			return response.WrapBusinessError(response.ErrDatabase, err, "")
		}
		return nil
	})
}
