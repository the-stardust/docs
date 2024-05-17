package helper

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"interview/common/global"
	"strings"
	"time"
)

// 消费者
func KafkaNewReader(topic string, groupId string) *kafka.Reader {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  strings.Split(global.CONFIG.Kafka.Address, ","),
		GroupID:  groupId,
		Topic:    topic,
		MaxBytes: 10e6, // 10MB
	})

	return r
}

// 生产者 记得 调用close方法
func KafkaNewWriter() *kafka.Writer {
	w := &kafka.Writer{
		Addr:     kafka.TCP(strings.Split(global.CONFIG.Kafka.Address, ",")...),
		Balancer: &kafka.RoundRobin{},
	}
	return w
}

func KafkaSendMassage(w *kafka.Writer, topic string, content []byte) error {
	messages := kafka.Message{
		Topic: topic,
		Key:   []byte(fmt.Sprintf("%s-%d", topic, time.Now().UnixNano())),
		Value: content,
	}
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = w.WriteMessages(ctx, messages)
	if err != nil {
		return err
	}
	return nil
}
