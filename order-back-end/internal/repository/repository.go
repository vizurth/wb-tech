package order

import (
	"context"
	"order-back-end/internal/model"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OrderRepo репозиторий, часть слоистой архитектуры
type OrderRepo struct {
	db   *pgxpool.Pool
	psql sq.StatementBuilderType
}

// NewRepository создаем экземпляр репозитория
func NewRepository(db *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{
		db:   db,
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// GetAllOrders загружает все заказы с подгрузкой связанных данных
func (r *OrderRepo) GetAllOrders(ctx context.Context) ([]model.OrderInfo, error) {
	query, args, err := r.psql.
		Select("order_uid", "track_number", "entry", "locale", "internal_signature",
			"customer_id", "delivery_service", "shardkey", "sm_id", "date_created", "oof_shard").
		From("orders").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.OrderInfo

	for rows.Next() {
		var o model.OrderInfo
		if err := rows.Scan(
			&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale, &o.InternalSignature,
			&o.CustomerID, &o.DeliveryService, &o.ShardKey, &o.SmID, &o.DateCreated, &o.OofShard,
		); err != nil {
			return nil, err
		}

		// Delivery
		dq, dargs, _ := r.psql.
			Select("name", "phone", "zip", "city", "address", "region", "email").
			From("deliveries").
			Where(sq.Eq{"order_uid": o.OrderUID}).
			ToSql()

		err = r.db.QueryRow(ctx, dq, dargs...).Scan(
			&o.Delivery.Name, &o.Delivery.Phone, &o.Delivery.Zip, &o.Delivery.City,
			&o.Delivery.Address, &o.Delivery.Region, &o.Delivery.Email,
		)
		if err != nil {
			return nil, err
		}

		// Payment
		pq, pargs, _ := r.psql.
			Select("transaction", "request_id", "currency", "provider", "amount",
				"payment_dt", "bank", "delivery_cost", "goods_total", "custom_fee").
			From("payments").
			Where(sq.Eq{"order_uid": o.OrderUID}).
			ToSql()

		err = r.db.QueryRow(ctx, pq, pargs...).Scan(
			&o.Payment.Transaction, &o.Payment.RequestID, &o.Payment.Currency,
			&o.Payment.Provider, &o.Payment.Amount, &o.Payment.PaymentDT,
			&o.Payment.Bank, &o.Payment.DeliveryCost, &o.Payment.GoodsTotal, &o.Payment.CustomFee,
		)
		if err != nil {
			return nil, err
		}

		// Items
		iq, iargs, _ := r.psql.
			Select("chrt_id", "track_number", "price", "rid", "name",
				"sale", "size", "total_price", "nm_id", "brand", "status").
			From("items").
			Where(sq.Eq{"order_uid": o.OrderUID}).
			ToSql()

		itemRows, err := r.db.Query(ctx, iq, iargs...)
		if err != nil {
			return nil, err
		}

		for itemRows.Next() {
			var it model.Item
			if err := itemRows.Scan(
				&it.ChrtID, &it.TrackNumber, &it.Price, &it.RID, &it.Name,
				&it.Sale, &it.Size, &it.TotalPrice, &it.NmID, &it.Brand, &it.Status,
			); err != nil {
				itemRows.Close()
				return nil, err
			}
			o.Items = append(o.Items, it)
		}
		itemRows.Close()

		orders = append(orders, o)
	}

	return orders, nil
}

// GetOrderFromDB один заказ по ID
func (r *OrderRepo) GetOrderFromDB(ctx context.Context, orderID string) (*model.OrderInfo, error) {
	var o model.OrderInfo

	// Order
	oq, oargs, _ := r.psql.
		Select("order_uid", "track_number", "entry", "locale", "internal_signature",
			"customer_id", "delivery_service", "shardkey", "sm_id", "date_created", "oof_shard").
		From("orders").
		Where(sq.Eq{"order_uid": orderID}).
		ToSql()

	err := r.db.QueryRow(ctx, oq, oargs...).Scan(
		&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale, &o.InternalSignature,
		&o.CustomerID, &o.DeliveryService, &o.ShardKey, &o.SmID, &o.DateCreated, &o.OofShard,
	)
	if err != nil {
		return nil, err
	}

	// Delivery
	dq, dargs, _ := r.psql.
		Select("name", "phone", "zip", "city", "address", "region", "email").
		From("deliveries").
		Where(sq.Eq{"order_uid": o.OrderUID}).
		ToSql()

	err = r.db.QueryRow(ctx, dq, dargs...).Scan(
		&o.Delivery.Name, &o.Delivery.Phone, &o.Delivery.Zip, &o.Delivery.City,
		&o.Delivery.Address, &o.Delivery.Region, &o.Delivery.Email,
	)
	if err != nil {
		return nil, err
	}

	// Payment
	pq, pargs, _ := r.psql.
		Select("transaction", "request_id", "currency", "provider", "amount",
			"payment_dt", "bank", "delivery_cost", "goods_total", "custom_fee").
		From("payments").
		Where(sq.Eq{"order_uid": o.OrderUID}).
		ToSql()

	err = r.db.QueryRow(ctx, pq, pargs...).Scan(
		&o.Payment.Transaction, &o.Payment.RequestID, &o.Payment.Currency,
		&o.Payment.Provider, &o.Payment.Amount, &o.Payment.PaymentDT,
		&o.Payment.Bank, &o.Payment.DeliveryCost, &o.Payment.GoodsTotal, &o.Payment.CustomFee,
	)
	if err != nil {
		return nil, err
	}

	// Items
	iq, iargs, _ := r.psql.
		Select("chrt_id", "track_number", "price", "rid", "name",
			"sale", "size", "total_price", "nm_id", "brand", "status").
		From("items").
		Where(sq.Eq{"order_uid": o.OrderUID}).
		ToSql()

	itemRows, err := r.db.Query(ctx, iq, iargs...)
	if err != nil {
		return nil, err
	}

	for itemRows.Next() {
		var it model.Item
		if err := itemRows.Scan(
			&it.ChrtID, &it.TrackNumber, &it.Price, &it.RID, &it.Name,
			&it.Sale, &it.Size, &it.TotalPrice, &it.NmID, &it.Brand, &it.Status,
		); err != nil {
			itemRows.Close()
			return nil, err
		}
		o.Items = append(o.Items, it)
	}
	itemRows.Close()

	return &o, nil
}
