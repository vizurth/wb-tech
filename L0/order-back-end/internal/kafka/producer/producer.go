package kfk

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"order-back-end/internal/logger"
	"order-back-end/internal/model"
	"time"

	"github.com/IBM/sarama"
	"github.com/brianvoe/gofakeit"
	"github.com/hashicorp/go-uuid"
	"go.uber.org/zap"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// newSyncProducer создаем синхронного продюсера
func newSyncProducer(brokers []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Return.Successes = true
	config.Producer.MaxMessageBytes = 10 * 1024 * 1024
	producer, err := sarama.NewSyncProducer(brokers, config)

	return producer, err
}

// newAsyncProducer создаем асинхронного продюсера
func newAsyncProducer(brokers []string) (sarama.AsyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Return.Successes = true
	config.Producer.MaxMessageBytes = 10 * 1024 * 1024
	producer, err := sarama.NewAsyncProducer(brokers, config)

	return producer, err
}

// prepareMessage подготавливаем сообщение к отправке
func prepareMessage(topic string, message []byte) *sarama.ProducerMessage {
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Partition: -1,
		Value:     sarama.ByteEncoder(message),
	}
	return msg
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
	syncProducer, err := newSyncProducer(brokers)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "Error creating sync producer", zap.Error(err))
	}

	asyncProducer, err := newAsyncProducer(brokers)

	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "Failed to create async producer", zap.Error(err))
	}

	go func() {
		for err = range asyncProducer.Errors() {
			logger.GetLoggerFromCtx(ctx).Error(ctx, "Failed to produce message", zap.Error(err))
		}
	}()

	go func() {
		for succ := range asyncProducer.Successes() {
			logger.GetLoggerFromCtx(ctx).Info(ctx, "Successfully produced message", zap.Any("message", succ.Topic))
		}
	}()

	for {
		oi := generateOrder()

		oiJson, err := json.Marshal(oi)
		if err != nil {
			logger.GetLoggerFromCtx(ctx).Fatal(ctx, "Failed to marshal order info", zap.Error(err))
		}

		msg := prepareMessage(topic, oiJson)

		if rand.Int()%2 == 0 {
			partition, offset, err := syncProducer.SendMessage(msg)
			if err != nil {
				fmt.Printf("Msg sync err: %v\n", err)
			} else {
				fmt.Printf("Msg written sync. Patrition: %d. Ossfet: %d\n", partition, offset)
			}
		} else {
			asyncProducer.Input() <- msg
		}

		time.Sleep(10 * time.Second)
	}
}
