package shopservice

import (
	"context"
	"dianping/dal/model"
	"dianping/dal/query"
	"dianping/internal/db"
	"dianping/pkg/code"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

const COUNT_BITS = 32

type Voucher struct {
	ShopId      int    `json:"shopId"` //关联的商店id
	Title       string `json:"title"`
	SubTitle    string `json:"subTitle"`
	Rules       string `json:"rules"`
	PayValue    int    `json:"payValue"` //优惠的价格
	ActualValue int    `json:"actualValue"`
	Type        int    `json:"type"`  //优惠卷类型
	Stock       int    `json:"stock"` //库存
	BeginTime   string `json:"beginTime"`
	EndTime     string `json:"endTime"`
}

type seckillResquest struct {
	VoucherId int `json:"voucherId"`
	UserId    int `json:"userId"`
}

// 生成分布式id, 这个是订单id。不是优惠卷id
func NextId(keyPrefix string) int64 {
	now := time.Now().Unix()

	//生成序列号
	//Go语言的时间格式是通过一个特定的参考时间来定义的，这个参考时间是Mon Jan 2 15:04:05 MST 2006
	date := time.Now().Format("2006:01:01") //要用2006才能确保时间格式化正确

	count, err := db.RedisClient.Incr(context.Background(), "incr:"+keyPrefix+":"+date).Result()
	if err != nil {
		log.Println("Incr bad:", err)
		return -1
	}
	//拼接并返回
	return now<<COUNT_BITS | count
}

// 添加优惠卷
// post /voucher/add
func AddVoucher(c *gin.Context) {
	var value Voucher
	err := c.BindJSON(&value)
	if err != nil {
		slog.Error("AddVoucher, bind json bad", "err", err)
		code.WriteResponse(c, code.ErrBind, nil)
		return
	}

	switch value.Type {
	case 0: //普通优惠卷
		err = addOrdinaryVoucher(value)
	case 1:
		err = addSeckillVoucher(value)
	default:
		code.WriteResponse(c, code.ErrValidation, "type must be 0 or 1")
		return
	}

	if err != nil {
		if err.Error() == "time format error,must like 2006-01-02 15:04:05" {
			code.WriteResponse(c, code.ErrValidation, err.Error())
			return
		}
		code.WriteResponse(c, code.ErrDatabase, nil)
	}
	code.WriteResponse(c, code.ErrSuccess, nil)
}

// 添加秒杀券
func addSeckillVoucher(voucher Voucher) error {
	// 定义时间字符串的格式
	layout := "2006-01-02 15:04:05"
	start, err := time.Parse(layout, voucher.BeginTime)
	if err != nil {
		slog.Error("parse startTime bad:", "err", err)
		return fmt.Errorf("time format error,must like 2006-01-02 15:04:05")
	}
	end, err := time.Parse(layout, voucher.EndTime)
	if err != nil {
		slog.Error("parse endTime bad:", "err", err)
		return fmt.Errorf("time format error,must like 2006-01-02 15:04:05")
	}

	v := model.TbVoucher{
		ShopID:      uint64(voucher.ShopId),
		Title:       voucher.Title,
		SubTitle:    voucher.SubTitle,
		Rules:       voucher.Rules,
		PayValue:    uint64(voucher.PayValue),
		ActualValue: int64(voucher.ActualValue),
		Type:        uint8(voucher.Type),
	}

	q := query.Use(db.DBEngine)
	//使用事务
	return q.Transaction(func(tx *query.Query) error {
		//1.先添加到优惠卷表 tb_voucher
		err := tx.TbVoucher.Create(&v)
		if err != nil {
			slog.Error("create voucher bad", "err", err)
			return err
		}

		//2.再添加信息到秒杀卷表 tb_seckill_voucher
		seckill := model.TbSeckillVoucher{
			VoucherID: v.ID,
			Stock:     int64(voucher.Stock),
			BeginTime: start,
			EndTime:   end,
		}
		err = tx.TbSeckillVoucher.Create(&seckill)
		if err != nil {
			slog.Error("create seckill voucher bad", "err", err)
			return err
		}
		return nil
	})
}

// 添加普通优惠卷
func addOrdinaryVoucher(voucher Voucher) error {
	v := model.TbVoucher{
		ShopID:      uint64(voucher.ShopId),
		Title:       voucher.Title,
		SubTitle:    voucher.SubTitle,
		Rules:       voucher.Rules,
		PayValue:    uint64(voucher.PayValue),
		ActualValue: int64(voucher.ActualValue),
		Type:        uint8(voucher.Type),
	}

	//往数据库添加
	err := query.TbVoucher.Create(&v)
	if err != nil {
		slog.Error("create voucher bad", "err", err)
		return err
	}

	return nil
}

// 秒杀
// post /voucher/seckill
func SeckillVoucher(c *gin.Context) {
	var req seckillResquest
	err := c.BindJSON(&req)
	if err != nil {
		slog.Error("SeckillVoucher, bind json bad", "err", err)
		code.WriteResponse(c, code.ErrBind, nil)
		return
	}

	//1.查询该优惠卷
	seckill := query.TbSeckillVoucher
	voucher, err := seckill.Where(seckill.VoucherID.Eq(uint64(req.VoucherId))).Find()
	if err != nil {
		code.WriteResponse(c, code.ErrDatabase, nil)
		return
	}
	if len(voucher) == 0 {
		code.WriteResponse(c, code.ErrDatabase, "秒杀卷不存在")
		return
	}
	//2.判断秒杀卷是否合法，开始结束时间,库存
	now := time.Now()
	if voucher[0].BeginTime.After(now) || voucher[0].EndTime.Before(now) {
		code.WriteResponse(c, code.ErrValidation, "不在秒杀时间范围内")
		return
	}
	if voucher[0].Stock < 1 {
		code.WriteResponse(c, code.ErrValidation, "秒杀卷已被抢空")
		return
	}

	createVoucherOrder(c, req)
}
func createVoucherOrder(c *gin.Context, req seckillResquest) {
	order := model.TbVoucherOrder{
		ID:        NextId("order"),
		VoucherID: uint64(req.VoucherId),
		UserID:    uint64(req.UserId),
	}

	//处理两张表(订单表，秒杀卷表)，使用事务
	q := query.Use(db.DBEngine)
	err := q.Transaction(func(tx *query.Query) error {
		//3.合法，库存数量减1
		//使用update，要是没有该条数据，不会返回gorm.ErrRecordNotFound或者有错误的。
		// info, err := tx.TbSeckillVoucher.Where(tx.TbSeckillVoucher.VoucherID.Eq(uint64(req.VoucherId))).UpdateSimple(tx.TbSeckillVoucher.Stock.Add(-1))
		info, err := tx.TbSeckillVoucher.Where(tx.TbSeckillVoucher.VoucherID.Eq(uint64(req.VoucherId)), tx.TbSeckillVoucher.Stock.Gt(0)).UpdateSimple(tx.TbSeckillVoucher.Stock.Add(-1))
		if err != nil {
			return err
		}
		if info.RowsAffected == 0 {
			return fmt.Errorf("affected rows is 0")
		}

		//4.成功，创建对应的订单，并保存到数据中
		// err = tx.TbVoucherOrder.Create(&order)
		//出现问题Error 1292 (22007): Incorrect datetime value: '0000-00-00' for column 'pay_time' at row 1
		//表 `tb_voucher_order` 的字段`pay_time`,`use_time`,`refund_time`类型是timestamp，不允许插入'00000-00-00 00:00:00',数据库不接受这种无效的日期时间值。
		//可以指定更新需要的字段，不更新其他字段
		err = tx.TbVoucherOrder.Select(tx.TbVoucherOrder.ID, tx.TbVoucherOrder.UserID, tx.TbVoucherOrder.VoucherID).Create(&order)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		slog.Error("seckill voucher bad", "err", err)
		code.WriteResponse(c, code.ErrDatabase, nil)
		return
	}
	code.WriteResponse(c, code.ErrSuccess, order.ID)
}
