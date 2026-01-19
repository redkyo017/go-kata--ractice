package orders

import (
	"context"
	repo "ecom-api/internal/adapters/postgresql/sqlc"
)

type orderItem struct {
	ProductID int64 `json:"productID"`
	Quantity  int32 `json:"quantity"`
}

type createOrderParams struct {
	CustomerID int64       `json:"customerId"`
	Items      []orderItem `json:"items"`
}

type Service interface {
	PlaceOrder(ctx context.Context, tempOrder createOrderParams) (repo.Order, error)
	GetOrderByID(ctx context.Context, orderId int) (repo.FindOrderByIDRow, error)
}
