package kfk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"order-back-end/internal/logger"
	"order-back-end/internal/model"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/hashicorp/go-uuid"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

var errUnknownType = errors.New("unknown type")

const (
	flushTimeout = 5000
)

type Producer struct {
	producer *kafka.Producer
}

func NewProducer(brokers []string) (*Producer, error) {
	conf := &kafka.ConfigMap{
		"bootstrap.servers": strings.Join(brokers, ","),
	}
	p, err := kafka.NewProducer(conf)
	if err != nil {
		return nil, fmt.Errorf("error creating kafka producer: %w", err)
	}
	return &Producer{producer: p}, nil
}

func (p *Producer) Produce(topic string) error {
	ctx := context.Background()
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	order := generateOrder()
	orderJson, err := json.Marshal(order)
	if err != nil {
		return err
	}
	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: orderJson,
		Key:   nil,
	}
	kafkaChan := make(chan kafka.Event)
	if err = p.producer.Produce(kafkaMsg, kafkaChan); err != nil {
		return fmt.Errorf("kafka producer error: %w", err)
	}
	e := <-kafkaChan
	switch ev := e.(type) {
	case *kafka.Message:
		log.Info(ctx, fmt.Sprintf("message to topic: %v", kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny}))
		return nil
	case kafka.Error:
		return ev
	default:
		return errUnknownType
	}
}

func (p *Producer) Close() {
	p.producer.Flush(flushTimeout)
	p.producer.Close()
}

// randomString генерируем строку
func randomString(n int, charset string) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[gofakeit.Number(0, len(charset)-1)]
	}
	return string(b)
}

// randomLocale генерируем локацию
func randomLocale() string {
	return gofakeit.RandString([]string{"en", "ru", "de", "fr", "es", "it", "pl"})
}

// generateOrder генерируем model.OrderInfo
func generateOrder() model.OrderInfo {
	gofakeit.Seed(time.Now().UnixNano())

	orderID, _ := uuid.GenerateUUID()
	RID, _ := uuid.GenerateUUID()
	trackNum := randomString(10, charset)

	return model.OrderInfo{
		OrderUID:    orderID,
		TrackNumber: trackNum,
		Entry:       randomString(4, charset),
		Delivery: model.Delivery{
			Name:    gofakeit.Name(),
			Phone:   gofakeit.Phone(),
			Zip:     gofakeit.Zip(),
			City:    gofakeit.City(),
			Address: gofakeit.Street(),
			Region:  gofakeit.State(),
			Email:   gofakeit.Email(),
		},
		Payment: model.Payment{
			Transaction:  orderID,
			RequestID:    "",
			Currency:     gofakeit.CurrencyShort(),
			Provider:     "wbpay",
			Amount:       gofakeit.Number(100, 5000),
			PaymentDT:    time.Now().Unix(),
			Bank:         gofakeit.Company(),
			DeliveryCost: gofakeit.Number(100, 1000),
			GoodsTotal:   gofakeit.Number(50, 2000),
			CustomFee:    0,
		},
		Items: []model.Item{
			{
				ChrtID:      gofakeit.Number(1000000, 9999999),
				TrackNumber: trackNum,
				Price:       gofakeit.Number(100, 1000),
				RID:         RID,
				Name:        gofakeit.Name(),
				Sale:        gofakeit.Number(0, 50),
				Size:        randomString(1, charset),
				TotalPrice:  gofakeit.Number(50, 1000),
				NmID:        gofakeit.Number(1000000, 9999999),
				Brand:       gofakeit.Company(),
				Status:      gofakeit.Number(100, 300),
			},
		},
		Locale:            randomLocale(),
		InternalSignature: "",
		CustomerID:        gofakeit.Username(),
		DeliveryService:   gofakeit.Company(),
		ShardKey:          randomString(1, "0123456789"),
		SmID:              gofakeit.Number(1, 999),
		DateCreated:       time.Now(),
		OofShard:          randomString(1, "0123456789"),
	}
}

// StartProducer начинаем отправку сообщений
func StartProducer(ctx context.Context, brokers []string, topic string) {
	p, err := NewProducer(brokers)
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	if err != nil {
		log.Info(ctx, "error creating kafka producer")
	}
	for {
		if err := p.Produce(topic); err != nil {
			log.Info(ctx, "error producing message")
		}
		log.Info(ctx, "message produced")
		time.Sleep(10 * time.Second)
	}
}
