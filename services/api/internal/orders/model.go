package orders

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusPaid      Status = "paid"
	StatusShipped   Status = "shipped"
	StatusDelivered Status = "delivered"
	StatusCancelled Status = "cancelled"
)

type Order struct {
	ID              uuid.UUID   `json:"id"`
	UserID          *uuid.UUID  `json:"user_id,omitempty"`
	Status          Status      `json:"status"`
	TotalCents      int64       `json:"total_cents"`
	CheckoutSessionID *string     `json:"checkout_session_id,omitempty"`
	Items           []OrderItem `json:"items"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID         uuid.UUID `json:"id"`
	OrderID    uuid.UUID `json:"order_id"`
	VariantID  uuid.UUID `json:"variant_id"`
	Quantity   int       `json:"quantity"`
	PriceCents int64     `json:"price_cents"`
}

type CreateOrderRequest struct {
	Items []OrderItemInput `json:"items"`
}

type OrderItemInput struct {
	VariantID  uuid.UUID `json:"variant_id"`
	Quantity   int       `json:"quantity"`
	PriceCents int64     `json:"price_cents"`
}
