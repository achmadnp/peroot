package model

import "time"

type OrderRequest struct {
	OrderID      string      `json:"order_id" binding:"required"`
	CustomerName string      `json:"customer_name" binding:"required"`
	TotalAmount  float64     `json:"total_amount" binding:"required,gt=0"`
	Currency     string      `json:"currency" binding:"required"`
	Items        []OrderItem `json:"items" binding:"required,min=1,dive"`
}

type OrderItem struct {
	Name     string  `json:"name" binding:"required"`
	Quantity int     `json:"quantity" binding:"required,gt=0"`
	Price    float64 `json:"price" binding:"required,gt=0"`
}

type OrderResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	OrderID   string    `json:"order_id"`
	MessageID string    `json:"message_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}
