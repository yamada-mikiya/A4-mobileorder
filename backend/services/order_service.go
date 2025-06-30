package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/google/uuid"
)

type OrderServicer interface {
	GetProductListService(shopID int) error
	CreateOrder(ctx context.Context, shopID int, reqProd []models.OrderProductRequest) (*models.Order, error)
	CreateAuthenticatedOrder(ctx context.Context, userID int, shopID int, products []models.OrderProductRequest) (*models.Order, error)
}

type orderService struct {
	repo repositories.OrderRepository
}

func NewOrderService(r repositories.OrderRepository) OrderServicer {
	return &orderService{repo: r}
}

func (s *orderService) GetProductListService(shopID int) error {
	//TODO
	return nil
}

func generateUserToken() string {
	token, err := uuid.NewRandom()
	if err != nil {
		log.Fatalf("FATAL: Failed to generate UUID for guest token: %v", err)
	}

	return token.String()
}

func (s *orderService) CreateOrder(ctx context.Context, shopID int, products []models.OrderProductRequest) (*models.Order, error) {

	totalAmount, orderProductsToCreate, err := s.validateAndPrepareOrderProducts(ctx, shopID, products)
	if err != nil {
		return nil, fmt.Errorf("fail to calculate total amount: %v", err)
	}

	userToken := generateUserToken()

	order := &models.Order{
		ShopID:         shopID,
		TotalAmount:    totalAmount,
		Status:         models.Cooking,
		UserOrderToken: sql.NullString{String: userToken, Valid: true},
	}

	if err := s.repo.CreateOrder(ctx, order, orderProductsToCreate); err != nil {
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

	if err := s.repo.CreateOrder(ctx, order, orderProductsToCreate); err != nil {
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

	validProductMap, err := s.repo.ValidateAndGetProductsForShop(ctx, shopID, productIDs)
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
