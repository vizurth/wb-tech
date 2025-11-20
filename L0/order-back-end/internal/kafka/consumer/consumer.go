package kfk

import (
	"context"
	"fmt"
	"order-back-end/internal/cache"
	"order-back-end/internal/logger"
	"order-back-end/internal/model"
	"order-back-end/internal/validator"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Consumer дополненая структура с db и cache
type Consumer struct {
	consumer       *kafka.Consumer
	db             *pgxpool.Pool
	cache          cache.Cache
	stop           bool
	consumerNumber int
}

// NewConsumer создаем экземпляр Consumer куда прокидывыем db и cache
func NewConsumer(brokers []string, topic, consumerGroup string, db *pgxpool.Pool, cache cache.Cache, consInt int) (*Consumer, error) {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers":        strings.Join(brokers, ","),
		"group.id":                 consumerGroup,
		"enable.auto.offset.store": false,
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  5000,
		"auto.offset.reset":        "earliest",
	}

	c, err := kafka.NewConsumer(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating kafka consumer: %w", err)
	}

	err = c.Subscribe(topic, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}

	return &Consumer{
		consumer:       c,
		db:             db,
		cache:          cache,
		stop:           false,
		consumerNumber: consInt,
	}, nil
}

func (c *Consumer) Start() {
	ctx := context.Background()
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	for {
		if c.stop {
			break
		}
		kafkaMsg, err := c.consumer.ReadMessage(-1)
		if err != nil {
			log.Error(ctx, fmt.Sprintf("Error reading message from consumer: %v", err))
		}
		if kafkaMsg == nil {
			continue
		}
		if err = c.prepareMessage(kafkaMsg); err != nil {
			log.Error(ctx, fmt.Sprintf("Error to transwer message to db from consumer: %v", err))
			continue
		}
		// сохраняем offset сообщения
		if _, err := c.consumer.StoreMessage(kafkaMsg); err != nil {
			log.Error(ctx, fmt.Sprintf("Error storing message in consumer: %v", err))
			continue
		}
	}
}

func (c *Consumer) Stop() error {
	c.stop = true
	// вручную коммитим то что сами не успели закоммитить это делать сама kafka
	if _, err := c.consumer.Commit(); err != nil {
		return err
	}
	return c.consumer.Close()
}

func (c *Consumer) prepareMessage(kafkaMsg *kafka.Message) (err error) {
	var msg model.OrderInfo
	err = validator.ValidateOrderInfo(kafkaMsg.Value, &msg)
	if err != nil {
		fmt.Printf("Error validating message: %s\n", err)
		return err
	}

	// Сохраняем в кэш
	c.cache.Set(msg.OrderUID, msg)
	fmt.Println(msg.OrderUID)

	ctx := context.Background()

	// начинаем транзакцию
	tx, err := c.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		fmt.Printf("Failed to start transaction: %s\n", err)
		return nil
	}

	// подготавливаем транзакцию
	if err := c.processMessage(ctx, tx, msg); err != nil {
		fmt.Printf("Failed to process message: %s\n", err)
		tx.Rollback(ctx)
		return nil
	}

	// коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		fmt.Printf("Failed to commit transaction: %s\n", err)
		return nil
	}
	return nil
}

// processMessage подготавливаем msg для отправки в бд
func (consumer *Consumer) processMessage(ctx context.Context, tx pgx.Tx, msg model.OrderInfo) error {
	if err := insertOrder(ctx, tx, msg); err != nil {
		return fmt.Errorf("insert order failed: %w", err)
	}
	if err := insertDelivery(ctx, tx, msg); err != nil {
		return fmt.Errorf("insert delivery failed: %w", err)
	}
	if err := insertPayment(ctx, tx, msg); err != nil {
		return fmt.Errorf("insert payment failed: %w", err)
	}
	if err := insertItems(ctx, tx, msg); err != nil {
		return fmt.Errorf("insert items failed: %w", err)
	}
	return nil
}

// insertOrder вставляем order
func insertOrder(ctx context.Context, tx pgx.Tx, msg model.OrderInfo) error {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	orderBuilder := psql.Insert("orders").
		Columns("order_uid", "track_number", "entry", "locale", "internal_signature",
			"customer_id", "delivery_service", "shardkey", "sm_id", "date_created", "oof_shard").
		Values(msg.OrderUID, msg.TrackNumber, msg.Entry, msg.Locale, msg.InternalSignature,
			msg.CustomerID, msg.DeliveryService, msg.ShardKey, msg.SmID, msg.DateCreated, msg.OofShard)

	sqlStr, args, _ := orderBuilder.ToSql()
	_, err := tx.Exec(ctx, sqlStr, args...)
	return err
}

// insertDelivery вставляем delivery
func insertDelivery(ctx context.Context, tx pgx.Tx, msg model.OrderInfo) error {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	deliveryBuilder := psql.Insert("deliveries").
		Columns("order_uid", "name", "phone", "zip", "city", "address", "region", "email").
		Values(msg.OrderUID, msg.Delivery.Name, msg.Delivery.Phone, msg.Delivery.Zip,
			msg.Delivery.City, msg.Delivery.Address, msg.Delivery.Region, msg.Delivery.Email)

	sqlStr, args, _ := deliveryBuilder.ToSql()
	_, err := tx.Exec(ctx, sqlStr, args...)
	return err
}

// insertPayment вставляем payment
func insertPayment(ctx context.Context, tx pgx.Tx, msg model.OrderInfo) error {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	paymentBuilder := psql.Insert("payments").
		Columns("transaction", "order_uid", "request_id", "currency", "provider",
			"amount", "payment_dt", "bank", "delivery_cost", "goods_total", "custom_fee").
		Values(msg.Payment.Transaction, msg.OrderUID, msg.Payment.RequestID, msg.Payment.Currency,
			msg.Payment.Provider, msg.Payment.Amount, msg.Payment.PaymentDT, msg.Payment.Bank,
			msg.Payment.DeliveryCost, msg.Payment.GoodsTotal, msg.Payment.CustomFee)

	sqlStr, args, _ := paymentBuilder.ToSql()
	_, err := tx.Exec(ctx, sqlStr, args...)
	return err
}

// insertItems вставляем items
func insertItems(ctx context.Context, tx pgx.Tx, msg model.OrderInfo) error {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	for _, item := range msg.Items {
		itemBuilder := psql.Insert("items").
			Columns("order_uid", "chrt_id", "track_number", "price", "rid", "name",
				"sale", "size", "total_price", "nm_id", "brand", "status").
			Values(msg.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name,
				item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)

		sqlStr, args, _ := itemBuilder.ToSql()
		if _, err := tx.Exec(ctx, sqlStr, args...); err != nil {
			return err
		}
	}
	return nil
}

// StartConsuming начинаем прослушку
func StartConsuming(ctx context.Context, brokers []string, groupID, topic string, db *pgxpool.Pool, cache cache.Cache) {
	log := logger.GetOrCreateLoggerFromCtx(ctx)
	c1, err := NewConsumer(brokers, topic, "my-consumer-group", db, cache, 1)
	if err != nil {
		log.Error(ctx, "error creating consumer", zap.Error(err))
	}
	c2, err := NewConsumer(brokers, topic, "my-consumer-group", db, cache, 2)
	if err != nil {
		log.Error(ctx, "error creating consumer", zap.Error(err))
	}
	c3, err := NewConsumer(brokers, topic, "my-consumer-group", db, cache, 3)
	if err != nil {
		log.Error(ctx, "error creating consumer", zap.Error(err))
	}
	go c1.Start()
	go c2.Start()
	go c3.Start()
}
