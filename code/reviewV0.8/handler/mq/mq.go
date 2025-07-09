package mq

import (
	"context"
	"fmt"
	"log/slog"
	"review/config"
	"review/db"
	"review/handler/order"
	"review/pkg/mail"
	"strconv"

	"github.com/redis/go-redis/v9"
)

const (
	streamName      = "stream.orders"
	streamGroupName = "group1"
)

func StartStream() {
	// 创建消费组（如果不存在）
	err := db.RedisDb.XGroupCreateMkStream(context.Background(), streamName, streamGroupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		slog.Error("Failed to create consumer group", "err", err)
		panic(err)
	}

	// 从消费组中读取消息
	for i := 1; i <= 4; i++ {
		name := "consumer" + fmt.Sprint(i)
		go startStream(name)
	}
}

func startStream(name string) {
	// 从消费组中读取消息
	for {
		msgs, err := db.RedisDb.XReadGroup(context.Background(), &redis.XReadGroupArgs{
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

		err = order.CreateOrder(voucherIdInt, userIdInt, orderIdInt)
		if err != nil {
			slog.Error("Failed to create voucher order", "err", err)
			//再次尝试
			err = order.CreateOrder(voucherIdInt, userIdInt, orderIdInt)
			if err != nil {
				// 发送邮件让人工处理。或者发送到死信队列
				body := "voucherId:" + voucherId + ", userId:" + userId + ", orderId:" + orderId
				mail.SendMail(*config.MailOption, body)
			}
		}

		//确认消息,发送ack
		db.RedisDb.XAck(context.Background(), streamName, streamGroupName, msgs[0].Messages[0].ID)
	}
}
