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
	CreateOrder(ctx context.Context, shopID int, reqProd []models.OrderProductRequest) (*models.Order, error)
	CreateAuthenticatedOrder(ctx context.Context, userID int, shopID int, products []models.OrderProductRequest) (*models.Order, error)
	GetUserOrders(ctx context.Context, userID int, statusParams []string) ([]models.OrderListResponse, error)
	GetOrderStatus(ctx context.Context, userID int, orderID int) (*models.OrderStatusResponse, error)
}

type orderService struct {
	orr repositories.OrderRepository
	prr repositories.ProductRepository
}

func NewOrderService(orr repositories.OrderRepository, prr repositories.ProductRepository) OrderServicer {
	return &orderService{orr, prr}
}

func generateguestToken() (string, error) {
	token, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID for guest token: %v", err)
	}

	return token.String(), nil
}

func (s *orderService) CreateOrder(ctx context.Context, shopID int, products []models.OrderProductRequest) (*models.Order, error) {

	totalAmount, orderProductsToCreate, err := s.validateAndPrepareOrderProducts(ctx, shopID, products)
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

	if err := s.orr.CreateOrder(ctx, order, orderProductsToCreate); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *orderService) CreateAuthenticatedOrder(ctx context.Context, userID int, shopID int, products []models.OrderProductRequest) (*models.Order, error) {
	totalAmount, orderProductsToCreate, err := s.validateAndPrepareOrderProducts(ctx, shopID, products)
	if err != nil {
		return nil, err
	}

	order := &models.Order{
		UserID:      sql.NullInt64{Int64: int64(userID), Valid: true},
		ShopID:      shopID,
		TotalAmount: totalAmount,
		Status:      models.Cooking,
	}

	if err := s.orr.CreateOrder(ctx, order, orderProductsToCreate); err != nil {
		return nil, err
	}

	return order, nil
}

// 商品が店のものとあっているかの検証と合計金額とorder_productテーブルに入れるためのデータを作るヘルパーメソッド
func (s *orderService) validateAndPrepareOrderProducts(ctx context.Context, shopID int, products []models.OrderProductRequest) (float64, []models.OrderProduct, error) {

	if len(products) == 0 {
		return 0, nil, errors.New("cannot create order with no products")
	}

	productIDs := make([]int, len(products))
	for i, product := range products {
		productIDs[i] = product.ProductID
	}

	validProductMap, err := s.prr.ValidateAndGetProductsForShop(ctx, shopID, productIDs)
	if err != nil {
		return 0, nil, err
	}

	var totalAmount float64 = 0
	orderProductsToCreate := make([]models.OrderProduct, len(products))
	for i, product := range products {
		productModel := validProductMap[product.ProductID]
		priceAtOrder := productModel.Price
		totalAmount += priceAtOrder * float64(product.Quantity)

		orderProductsToCreate[i] = models.OrderProduct{
			ProductID:    product.ProductID,
			Quantity:     product.Quantity,
			PriceAtOrder: priceAtOrder,
		}
	}
	return totalAmount, orderProductsToCreate, nil
}

// GetUserOrders は、注文一覧ページのためのやつ
func (s *orderService) GetUserOrders(ctx context.Context, userID int, statusParams []string) ([]models.OrderListResponse, error) {
	var statuses []models.OrderStatus
	for _, p := range statusParams {
		switch p {
		case "cooking":
			statuses = append(statuses, models.Cooking)
		case "completed":
			statuses = append(statuses, models.Completed)
		}
	}
	if len(statuses) == 0 {
		return []models.OrderListResponse{}, nil
	}

	orders, err := s.orr.FindUserOrdersWithDetails(ctx, userID, statuses)
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

	orderItemsMap, err := s.orr.FindProductsByOrderIDs(ctx, orderIDs)
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
