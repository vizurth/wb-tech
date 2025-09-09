package validator

import (
	"encoding/json"
	"errors"
	"order-back-end/internal/model"
	"strings"
	"time"
)

// ValidateOrderInfo парсит JSON и валидирует данные заказа
func ValidateOrderInfo(value []byte, order *model.OrderInfo) error {

	// пробуем распарсить JSON
	if err := json.Unmarshal(value, &order); err != nil {
		return errors.New("invalid JSON: " + err.Error())
	}

	// обязательные строковые поля
	if strings.TrimSpace(order.OrderUID) == "" {
		return errors.New("order_uid is required")
	}
	if strings.TrimSpace(order.TrackNumber) == "" {
		return errors.New("track_number is required")
	}
	if strings.TrimSpace(order.Entry) == "" {
		return errors.New("entry is required")
	}
	if strings.TrimSpace(order.CustomerID) == "" {
		return errors.New("customer_id is required")
	}
	if strings.TrimSpace(order.DeliveryService) == "" {
		return errors.New("delivery_service is required")
	}
	if order.SmID == 0 {
		return errors.New("sm_id must be > 0")
	}
	if order.DateCreated.IsZero() || order.DateCreated.After(time.Now().Add(24*time.Hour)) {
		return errors.New("date_created is invalid")
	}

	// Delivery
	if strings.TrimSpace(order.Delivery.Name) == "" {
		return errors.New("delivery.name is required")
	}
	if strings.TrimSpace(order.Delivery.Phone) == "" {
		return errors.New("delivery.phone is required")
	}
	if strings.TrimSpace(order.Delivery.Address) == "" {
		return errors.New("delivery.address is required")
	}
	if strings.TrimSpace(order.Delivery.City) == "" {
		return errors.New("delivery.city is required")
	}
	if strings.TrimSpace(order.Delivery.Email) == "" {
		return errors.New("delivery.email is required")
	}

	// Payment
	if strings.TrimSpace(order.Payment.Transaction) == "" {
		return errors.New("payment.transaction is required")
	}
	if strings.TrimSpace(order.Payment.Currency) == "" {
		return errors.New("payment.currency is required")
	}
	if strings.TrimSpace(order.Payment.Provider) == "" {
		return errors.New("payment.provider is required")
	}
	if order.Payment.Amount <= 0 {
		return errors.New("payment.amount must be > 0")
	}
	if order.Payment.PaymentDT <= 0 {
		return errors.New("payment.payment_dt must be valid unix timestamp")
	}

	// Items
	if len(order.Items) == 0 {
		return errors.New("at least one item is required")
	}
	for i, item := range order.Items {
		if item.ChrtID == 0 {
			return errors.New("items[" + string(rune(i)) + "].chrt_id must be > 0")
		}
		if strings.TrimSpace(item.TrackNumber) == "" {
			return errors.New("items[" + string(rune(i)) + "].track_number is required")
		}
		if strings.TrimSpace(item.Name) == "" {
			return errors.New("items[" + string(rune(i)) + "].name is required")
		}
		if item.Price <= 0 {
			return errors.New("items[" + string(rune(i)) + "].price must be > 0")
		}
		if item.TotalPrice < 0 {
			return errors.New("items[" + string(rune(i)) + "].total_price must be >= 0")
		}
	}

	return nil
}
