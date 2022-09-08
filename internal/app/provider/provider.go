package provider

import (
	"encoding/json"

	"github.com/go-resty/resty/v2"

	"github.com/hikjik/gophermart/internal/app/models"
)

type Provider interface {
	GetOrderAccrual(orderID string) (*models.Order, error)
}

type provider struct {
	address string
}

type OrderAccrual struct {
	Num     string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

func New(address string) Provider {
	return &provider{address: address}
}

func (p *provider) GetOrderAccrual(orderNum string) (*models.Order, error) {
	client := resty.New().SetRetryCount(5)

	url := p.address + "/api/orders/" + orderNum
	resp, err := client.R().Get(url)
	if err != nil {
		return nil, err
	}

	var order OrderAccrual
	if err = json.Unmarshal(resp.Body(), &order); err != nil {
		return nil, err
	}
	return &models.Order{
		Number:  order.Num,
		Status:  models.OrderStatus(order.Status),
		Accrual: order.Accrual,
	}, nil
}
