package order

import (
	"log/slog"
	"review/dal/model"
	"review/dal/query"
	"review/db"
	"review/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	timeLayout      = "2006-01-02 15:04:05"
	timeFormatError = "time format error, must be like 2006-01-02 15:04:05"
)

// post /api/v1/vouchers
func AddVoucher(c *gin.Context) {
	var value voucher
	err := c.BindJSON(&value)
	if err != nil {
		slog.Error("AddVoucher, bind json bad", "err", err)
		response.Error(c, response.ErrBind)
		return
	}

	var id uint64
	switch value.Type {
	case 0: //普通优惠卷
		id, err = addOrdinaryVoucher(value)
	case 1:
		id, err = addSeckillVoucher(value)
	default:
		response.Error(c, response.ErrValidation, "type must be 0 or 1")
		return
	}
	response.HandleBusinessResult(c, err, gin.H{"voucherId": id})
}

func addOrdinaryVoucher(voucherReq voucher) (uint64, error) {
	v := model.TbVoucher{
		ShopID:      uint64(voucherReq.ShopId),
		Title:       voucherReq.Title,
		SubTitle:    voucherReq.SubTitle,
		Rules:       voucherReq.Rules,
		PayValue:    uint64(voucherReq.PayValue),
		ActualValue: int64(voucherReq.ActualValue),
		Type:        uint32(voucherReq.Type),
	}

	//往数据库添加
	err := query.TbVoucher.Create(&v)
	if err != nil {
		return 0, response.WrapBusinessError(response.ErrDatabase, err, "")
	}

	return v.ID, nil
}

// 添加秒杀券
func addSeckillVoucher(voucherReq voucher) (uint64, error) {
	start, err := time.Parse(timeLayout, voucherReq.BeginTime)
	if err != nil {
		return 0, response.WrapBusinessError(response.ErrValidation, err, "BeginTime "+timeFormatError)
	}
	end, err := time.Parse(timeLayout, voucherReq.EndTime)
	if err != nil {
		return 0, response.WrapBusinessError(response.ErrValidation, err, "EndTime "+timeFormatError)
	}
	// 验证时间逻辑
	if !end.After(start) {
		return 0, response.WrapBusinessError(response.ErrValidation, nil, "EndTime must be after BeginTime")
	}

	v := model.TbVoucher{
		ShopID:      uint64(voucherReq.ShopId),
		Title:       voucherReq.Title,
		SubTitle:    voucherReq.SubTitle,
		Rules:       voucherReq.Rules,
		PayValue:    uint64(voucherReq.PayValue),
		ActualValue: int64(voucherReq.ActualValue),
		Type:        uint32(voucherReq.Type),
	}

	q := query.Use(db.DBEngine)
	//使用事务
	err = q.Transaction(func(tx *query.Query) error {
		//1.先添加到优惠卷表 tb_voucher
		err := tx.TbVoucher.Create(&v)
		if err != nil {
			return response.WrapBusinessError(response.ErrDatabase, err, "")
		}

		//2.再添加信息到秒杀卷表 tb_seckill_voucher
		seckill := model.TbSeckillVoucher{
			VoucherID: v.ID,
			Stock:     int32(voucherReq.Stock),
			BeginTime: start,
			EndTime:   end,
		}
		err = tx.TbSeckillVoucher.Create(&seckill)
		if err != nil {
			return response.WrapBusinessError(response.ErrDatabase, err, "")
		}
		return nil
	})
	return v.ID, err
}
