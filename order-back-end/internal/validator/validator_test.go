package validator_test

import (
	"encoding/json"
	"testing"
	"time"

	"order-back-end/internal/model"
	"order-back-end/internal/validator"

	"github.com/stretchr/testify/require"
)

func makeValidOrder() *model.OrderInfo {
	return &model.OrderInfo{
		OrderUID:        "123",
		TrackNumber:     "TRACK123",
		Entry:           "WEB",
		CustomerID:      "cust01",
		DeliveryService: "DHL",
		SmID:            1,
		DateCreated:     time.Now(),
		Delivery: model.Delivery{
			Name:    "John Doe",
			Phone:   "123456",
			Address: "Street 1",
			City:    "Moscow",
			Email:   "test@test.com",
		},
		Payment: model.Payment{
			Transaction: "tr-123",
			Currency:    "RUB",
			Provider:    "card",
			Amount:      1000,
			PaymentDT:   time.Now().Unix(),
		},
		Items: []model.Item{
			{
				ChrtID:      1,
				TrackNumber: "TRACK123",
				Name:        "Item1",
				Price:       500,
				TotalPrice:  500,
			},
		},
	}
}

func TestValidateOrderInfo_Valid(t *testing.T) {
	order := makeValidOrder()
	data, _ := json.Marshal(order)

	var parsed model.OrderInfo
	err := validator.ValidateOrderInfo(data, &parsed)
	require.NoError(t, err, "valid order must pass validation")
}

func TestValidateOrderInfo_InvalidJSON(t *testing.T) {
	var parsed model.OrderInfo
	err := validator.ValidateOrderInfo([]byte("{invalid json}"), &parsed)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JSON")
}

func TestValidateOrderInfo_MissingFields(t *testing.T) {
	// пример: убираем OrderUID
	order := makeValidOrder()
	order.OrderUID = ""
	data, _ := json.Marshal(order)

	var parsed model.OrderInfo
	err := validator.ValidateOrderInfo(data, &parsed)
	require.Error(t, err)
	require.Equal(t, "order_uid is required", err.Error())
}

func TestValidateOrderInfo_InvalidPayment(t *testing.T) {
	order := makeValidOrder()
	order.Payment.Amount = 0
	data, _ := json.Marshal(order)

	var parsed model.OrderInfo
	err := validator.ValidateOrderInfo(data, &parsed)
	require.Error(t, err)
	require.Equal(t, "payment.amount must be > 0", err.Error())
}

func TestValidateOrderInfo_NoItems(t *testing.T) {
	order := makeValidOrder()
	order.Items = []model.Item{}
	data, _ := json.Marshal(order)

	var parsed model.OrderInfo
	err := validator.ValidateOrderInfo(data, &parsed)
	require.Error(t, err)
	require.Equal(t, "at least one item is required", err.Error())
}

func TestValidateOrderInfo_InvalidTrackNumber(t *testing.T) {
	order := model.OrderInfo{
		OrderUID:    "123",
		TrackNumber: "",
	}

	data, _ := json.Marshal(order)

	var parsed model.OrderInfo
	err := validator.ValidateOrderInfo(data, &parsed)
	require.Error(t, err)
	require.Equal(t, "track_number is required", err.Error())
}

func TestValidateOrderInfo_InvalidEntry(t *testing.T) {
	order := model.OrderInfo{
		OrderUID:    "123",
		TrackNumber: "TRACK123",
		Entry:       "",
	}

	data, _ := json.Marshal(order)

	var parsed model.OrderInfo
	err := validator.ValidateOrderInfo(data, &parsed)
	require.Error(t, err)
	require.Equal(t, "entry is required", err.Error())
}

func TestValidateOrderInfo_InvalidCustomerID(t *testing.T) {
	order := model.OrderInfo{
		OrderUID:    "123",
		TrackNumber: "TRACK123",
		Entry:       "Entry",
		CustomerID:  "",
	}

	data, _ := json.Marshal(order)

	var parsed model.OrderInfo
	err := validator.ValidateOrderInfo(data, &parsed)
	require.Error(t, err)
	require.Equal(t, "customer_id is required", err.Error())
}

func TestValidateOrderInfo_InvalidDeliveryService(t *testing.T) {
	order := model.OrderInfo{
		OrderUID:        "123",
		TrackNumber:     "TRACK123",
		Entry:           "Entry",
		CustomerID:      "29384",
		DeliveryService: "",
	}

	data, _ := json.Marshal(order)

	var parsed model.OrderInfo
	err := validator.ValidateOrderInfo(data, &parsed)
	require.Error(t, err)
	require.Equal(t, "delivery_service is required", err.Error())
}

func TestValidateOrderInfo_InvalidSmID(t *testing.T) {
	order := model.OrderInfo{
		OrderUID:        "123",
		TrackNumber:     "TRACK123",
		Entry:           "Entry",
		CustomerID:      "29384",
		DeliveryService: "delivery-service",
		SmID:            0,
	}

	data, _ := json.Marshal(order)

	var parsed model.OrderInfo
	err := validator.ValidateOrderInfo(data, &parsed)
	require.Error(t, err)
	require.Equal(t, "sm_id must be > 0", err.Error())
}
