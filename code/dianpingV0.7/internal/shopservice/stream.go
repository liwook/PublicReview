package shopservice

import (
	"context"
	"dianping/dal/model"
	"dianping/dal/query"
	"dianping/internal/config"
	"dianping/internal/db"
	"dianping/pkg/mail"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/redis/go-redis/v9"
)

const (
	streamName      = "stream.orders"
	streamGroupName = "group1"
)

func StartStream() {
	// 创建消费组（如果不存在）
	err := db.RedisClient.XGroupCreateMkStream(context.Background(), streamName, streamGroupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		slog.Error("Failed to create consumer group", "err", err)
		panic(err)
	}

	// 从消费组中读取消息
	for i := 1; i <= 5; i++ {
		name := "consumer" + fmt.Sprint(i)
		go startStream(name)
	}
}

func startStream(name string) {
	// 从消费组中读取消息
	for {
		msgs, err := db.RedisClient.XReadGroup(context.Background(), &redis.XReadGroupArgs{
			Group:    streamGroupName,
			Consumer: name,
			Streams:  []string{streamName, ">"}, //streamName：这是要读取的 Redis Stream 的名称。表示从这个特定的 Stream 中读取消息。
			//">"：这个特殊的标识符在 Redis Stream 中用于表示从 Stream 的末尾开始读取，即只读取尚未被任何消费者处理的新消息。
			Count: 1,
			Block: 0,
		}).Result()
		if err != nil {
			slog.Error("Failed to read messages from stream", "err", err)
			continue
		}

		//处理信息
		if len(msgs) == 0 {
			continue
		}
		msg := msgs[0].Messages[0]
		fmt.Printf("Received message: %v", msg.Values)

		voucherId := msg.Values["voucherId"].(string)
		userId := msg.Values["userId"].(string)
		orderId := msg.Values["id"].(string)

		voucherIdInt, _ := strconv.Atoi(voucherId)
		userIdInt, _ := strconv.Atoi(userId)

		orderIdInt, _ := strconv.Atoi(orderId)

		err = createOrder(voucherIdInt, userIdInt, orderIdInt)
		if err != nil {
			slog.Error("Failed to create voucher order", "err", err)
			//再次尝试
			err = createOrder(voucherIdInt, userIdInt, orderIdInt)
			if err != nil {
				// 发送邮件让人工处理。或者发送到死信队列
				body := "voucherId:" + voucherId + ", userId:" + userId + ", orderId:" + orderId
				mail.SendMail(*config.MailOption, body)
			}
		}

		//确认消息,发送ack
		db.RedisClient.XAck(context.Background(), streamName, streamGroupName, msgs[0].Messages[0].ID)
	}
}

func createOrder(voucherId int, userId int, orderId int) error {
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
		return tx.TbVoucherOrder.Select(tx.TbVoucherOrder.ID, tx.TbVoucherOrder.UserID, tx.TbVoucherOrder.VoucherID).Create(&order)
	})
}
