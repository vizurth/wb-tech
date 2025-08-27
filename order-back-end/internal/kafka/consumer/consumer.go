package kfk

import (
	"context"
	"fmt"
	"order-back-end/internal/cache"
	"order-back-end/internal/logger"
	"order-back-end/internal/model"
	"order-back-end/internal/validator"

	"github.com/IBM/sarama"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Consumer дополненая структура с db и cache
type Consumer struct {
	db    *pgxpool.Pool
	cache cache.Cache
}

// NewConsumer создаем экземпляр Consumer куда прокидывыем db и cache
func NewConsumer(db *pgxpool.Pool, cache cache.Cache) *Consumer {
	return &Consumer{
		db:    db,
		cache: cache,
	}
}

// Setup функция для имплементации интерфейса
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup функция для имплементации интерфейса
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim принимаем сообщение из продюсера
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var msg model.OrderInfo
		err := validator.ValidateOrderInfo(message.Value, &msg)
		if err != nil {
			fmt.Printf("Error validating message: %s\n", err)
			return err
		}

		// Сохраняем в кэш
		consumer.cache.Set(msg.OrderUID, msg)
		fmt.Println(msg.OrderUID)

		ctx := context.Background()

		// начинаем транзакцию
		tx, err := consumer.db.BeginTx(ctx, pgx.TxOptions{
			IsoLevel:   pgx.ReadCommitted,
			AccessMode: pgx.ReadWrite,
		})
		if err != nil {
			fmt.Printf("Failed to start transaction: %s\n", err)
			continue
		}

		// подготавливаем транзакцию
		if err := consumer.processMessage(ctx, tx, msg); err != nil {
			fmt.Printf("Failed to process message: %s\n", err)
			tx.Rollback(ctx)
			continue
		}

		// коммитим транзакцию
		if err := tx.Commit(ctx); err != nil {
			fmt.Printf("Failed to commit transaction: %s\n", err)
			continue
		}

		// подтверждаем брокер
		session.MarkMessage(message, "")
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

// subscribe подписываемся на consumer
func (consumer *Consumer) subscribe(ctx context.Context, topic string, consumerGroup sarama.ConsumerGroup) error {

	go func() {
		for {
			if err := consumerGroup.Consume(ctx, []string{topic}, consumer); err != nil {
				fmt.Printf("Error from consumerGroup: %s\n\n", err)
			}
			if ctx.Err() != nil {
				logger.GetLoggerFromCtx(ctx).Error(ctx, "Error from consumerGroup.Consume()")
				return
			}
		}
	}()

	return nil
}

// StartConsuming начинаем прослушку
func (consumer *Consumer) StartConsuming(ctx context.Context, brokers []string, groupID, topic string) error {
	config := sarama.NewConfig()

	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGrop, err := sarama.NewConsumerGroup(brokers, groupID, config)

	if err != nil {
		return err
	}

	return consumer.subscribe(ctx, topic, consumerGrop)
}
