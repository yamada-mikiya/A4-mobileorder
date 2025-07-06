package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/google/uuid"
)

type OrderServicer interface {
	CreateOrder(ctx context.Context, shopID int, reqProd []models.OrderItemRequest) (*models.Order, error)
	CreateAuthenticatedOrder(ctx context.Context, userID int, shopID int, items []models.OrderItemRequest) (*models.Order, error)
	GetUserOrders(ctx context.Context, userID int) ([]models.OrderListResponse, error)
	GetOrderStatus(ctx context.Context, userID int, orderID int) (*models.OrderStatusResponse, error)
}

type orderService struct {
	orr repositories.OrderRepository
	prr repositories.ItemRepository
}

func NewOrderService(orr repositories.OrderRepository, prr repositories.ItemRepository) OrderServicer {
	return &orderService{orr, prr}
}

func generateguestToken() (string, error) {
	token, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID for guest token: %v", err)
	}

	return token.String(), nil
}

func (s *orderService) CreateOrder(ctx context.Context, shopID int, items []models.OrderItemRequest) (*models.Order, error) {

	totalAmount, orderItemsToCreate, err := s.validateAndPrepareOrderItems(ctx, shopID, items)
	if err != nil {
		return nil, fmt.Errorf("fail to calculate total amount: %v", err)
	}

	guestToken, err := generateguestToken()
	if err != nil {
		return nil, err
	}

	order := &models.Order{
		ShopID:          shopID,
		TotalAmount:     totalAmount,
		Status:          models.Cooking,
		GuestOrderToken: sql.NullString{String: guestToken, Valid: true},
	}

	if err := s.orr.CreateOrder(ctx, order, orderItemsToCreate); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *orderService) CreateAuthenticatedOrder(ctx context.Context, userID int, shopID int, items []models.OrderItemRequest) (*models.Order, error) {
	totalAmount, orderItemsToCreate, err := s.validateAndPrepareOrderItems(ctx, shopID, items)
	if err != nil {
		return nil, err
	}

	order := &models.Order{
		UserID:      sql.NullInt64{Int64: int64(userID), Valid: true},
		ShopID:      shopID,
		TotalAmount: totalAmount,
		Status:      models.Cooking,
	}

	if err := s.orr.CreateOrder(ctx, order, orderItemsToCreate); err != nil {
		return nil, err
	}

	return order, nil
}

// 商品が店のものとあっているかの検証と合計金額とorder_itemテーブルに入れるためのデータを作るヘルパーメソッド
func (s *orderService) validateAndPrepareOrderItems(ctx context.Context, shopID int, items []models.OrderItemRequest) (float64, []models.OrderItem, error) {

	if len(items) == 0 {
		return 0, nil, errors.New("cannot create order with no items")
	}

	itemIDs := make([]int, len(items))
	for i, item := range items {
		itemIDs[i] = item.ItemID
	}

	validItemMap, err := s.prr.ValidateAndGetItemsForShop(ctx, shopID, itemIDs)
	if err != nil {
		return 0, nil, err
	}

	var totalAmount float64 = 0
	orderItemsToCreate := make([]models.OrderItem, len(items))
	for i, item := range items {
		itemModel := validItemMap[item.ItemID]
		priceAtOrder := itemModel.Price
		totalAmount += priceAtOrder * float64(item.Quantity)

		orderItemsToCreate[i] = models.OrderItem{
			ItemID:       item.ItemID,
			Quantity:     item.Quantity,
			PriceAtOrder: priceAtOrder,
		}
	}
	return totalAmount, orderItemsToCreate, nil
}

// GetUserOrders は、注文一覧ページのためのやつ
func (s *orderService) GetUserOrders(ctx context.Context, userID int) ([]models.OrderListResponse, error) {

	orders, err := s.orr.FindActiveUserOrders(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return []models.OrderListResponse{}, nil
	}

	orderIDs := make([]int, len(orders))
	for i, o := range orders {
		orderIDs[i] = o.OrderID
	}

	orderItemsMap, err := s.orr.FindItemsByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, err
	}

	resDTOs := make([]models.OrderListResponse, len(orders))
	for i, repoOrder := range orders {
		resDTOs[i] = models.OrderListResponse{
			OrderID:      repoOrder.OrderID,
			ShopName:     repoOrder.ShopName,
			OrderDate:    repoOrder.OrderDate,
			TotalAmount:  repoOrder.TotalAmount,
			Status:       repoOrder.Status.String(),
			WaitingCount: repoOrder.WaitingCount,
			Items:        orderItemsMap[repoOrder.OrderID],
		}
	}

	return resDTOs, nil
}

// GetOrderStatus は、単一注文のステータスと待ち人数を取得
func (s *orderService) GetOrderStatus(ctx context.Context, userID int, orderID int) (*models.OrderStatusResponse, error) {

	order, err := s.orr.FindOrderByIDAndUser(ctx, orderID, userID)
	if err != nil {
		return nil, err
	}

	var waitingCount int
	if order.Status == models.Cooking {
		waitingCount, err = s.orr.CountWaitingOrders(ctx, order.ShopID, order.OrderDate)
		if err != nil {
			return nil, err
		}
	}

	return &models.OrderStatusResponse{
		OrderID:      order.OrderID,
		Status:       order.Status.String(),
		WaitingCount: waitingCount,
	}, nil
}
