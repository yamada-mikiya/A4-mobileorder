package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/internal/testhelpers"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/A4-dev-team/mobileorder.git/services"
	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
)

// ItemRepositoryMockForOrder - OrderService用のItemRepositoryモック（DBTX対応）
type ItemRepositoryMockForOrder struct {
	ValidateAndGetItemsForShopFunc func(ctx context.Context, dbtx repositories.DBTX, shopID int, itemIDs []int) (map[int]models.Item, error)
	GetItemListFunc                func(dbtx repositories.DBTX, shopID int) ([]models.ItemListResponse, error)
}

func NewItemRepositoryMockForOrder() *ItemRepositoryMockForOrder {
	return &ItemRepositoryMockForOrder{}
}

func (m *ItemRepositoryMockForOrder) ValidateAndGetItemsForShop(ctx context.Context, dbtx repositories.DBTX, shopID int, itemIDs []int) (map[int]models.Item, error) {
	if m.ValidateAndGetItemsForShopFunc != nil {
		return m.ValidateAndGetItemsForShopFunc(ctx, dbtx, shopID, itemIDs)
	}
	panic("not implemented")
}

func (m *ItemRepositoryMockForOrder) GetItemList(dbtx repositories.DBTX, shopID int) ([]models.ItemListResponse, error) {
	if m.GetItemListFunc != nil {
		return m.GetItemListFunc(dbtx, shopID)
	}
	panic("not implemented")
}

func (m *ItemRepositoryMockForOrder) UpdateItemAvailability(ctx context.Context, dbtx repositories.DBTX, itemID int, isAvailable bool) error {
	panic("not implemented")
}

func (m *ItemRepositoryMockForOrder) CreateItem(ctx context.Context, dbtx repositories.DBTX, item *models.Item) error {
	panic("not implemented")
}

func (m *ItemRepositoryMockForOrder) FindItemByID(ctx context.Context, dbtx repositories.DBTX, itemID int) (*models.Item, error) {
	panic("not implemented")
}

func (m *ItemRepositoryMockForOrder) UpdateItem(ctx context.Context, dbtx repositories.DBTX, item *models.Item) error {
	panic("not implemented")
}

func (m *ItemRepositoryMockForOrder) DeleteItem(ctx context.Context, dbtx repositories.DBTX, itemID int) error {
	panic("not implemented")
}

// OrderRepositoryMockForOrder - OrderService用のOrderRepositoryモック（DBTX対応）
type OrderRepositoryMockForOrder struct {
	CreateOrderFunc          func(ctx context.Context, dbtx repositories.DBTX, order *models.Order, items []models.OrderItem) error
	FindActiveUserOrdersFunc func(ctx context.Context, dbtx repositories.DBTX, userID int) ([]repositories.OrderWithDetailsDB, error)
	FindItemsByOrderIDsFunc  func(ctx context.Context, dbtx repositories.DBTX, orderIDs []int) (map[int][]models.ItemDetail, error)
	FindOrderByIDAndUserFunc func(ctx context.Context, dbtx repositories.DBTX, orderID int, userID int) (*models.Order, error)
	CountWaitingOrdersFunc   func(ctx context.Context, dbtx repositories.DBTX, shopID int, orderDate time.Time) (int, error)
}

func NewOrderRepositoryMockForOrder() *OrderRepositoryMockForOrder {
	return &OrderRepositoryMockForOrder{}
}

func (m *OrderRepositoryMockForOrder) CreateOrder(ctx context.Context, dbtx repositories.DBTX, order *models.Order, items []models.OrderItem) error {
	if m.CreateOrderFunc != nil {
		return m.CreateOrderFunc(ctx, dbtx, order, items)
	}
	panic("not implemented")
}

func (m *OrderRepositoryMockForOrder) UpdateUserIDByGuestToken(ctx context.Context, dbtx repositories.DBTX, guestToken string, userID int) error {
	panic("not implemented")
}

func (m *OrderRepositoryMockForOrder) FindActiveUserOrders(ctx context.Context, dbtx repositories.DBTX, userID int) ([]repositories.OrderWithDetailsDB, error) {
	if m.FindActiveUserOrdersFunc != nil {
		return m.FindActiveUserOrdersFunc(ctx, dbtx, userID)
	}
	panic("not implemented")
}

func (m *OrderRepositoryMockForOrder) FindItemsByOrderIDs(ctx context.Context, dbtx repositories.DBTX, orderIDs []int) (map[int][]models.ItemDetail, error) {
	if m.FindItemsByOrderIDsFunc != nil {
		return m.FindItemsByOrderIDsFunc(ctx, dbtx, orderIDs)
	}
	panic("not implemented")
}

func (m *OrderRepositoryMockForOrder) FindOrderByIDAndUser(ctx context.Context, dbtx repositories.DBTX, orderID int, userID int) (*models.Order, error) {
	if m.FindOrderByIDAndUserFunc != nil {
		return m.FindOrderByIDAndUserFunc(ctx, dbtx, orderID, userID)
	}
	panic("not implemented")
}

func (m *OrderRepositoryMockForOrder) CountWaitingOrders(ctx context.Context, dbtx repositories.DBTX, shopID int, orderDate time.Time) (int, error) {
	if m.CountWaitingOrdersFunc != nil {
		return m.CountWaitingOrdersFunc(ctx, dbtx, shopID, orderDate)
	}
	panic("not implemented")
}

func (m *OrderRepositoryMockForOrder) FindShopOrdersByStatuses(ctx context.Context, dbtx repositories.DBTX, shopID int, statuses []models.OrderStatus) ([]repositories.AdminOrderDBResult, error) {
	panic("not implemented")
}

func (m *OrderRepositoryMockForOrder) FindOrderByIDAndShopID(ctx context.Context, dbtx repositories.DBTX, orderID int, shopID int) (*models.Order, error) {
	panic("not implemented")
}

func (m *OrderRepositoryMockForOrder) UpdateOrderStatus(ctx context.Context, dbtx repositories.DBTX, orderID int, shopID int, newStatus models.OrderStatus) error {
	panic("not implemented")
}

func (m *OrderRepositoryMockForOrder) DeleteOrderByIDAndShopID(ctx context.Context, dbtx repositories.DBTX, orderID int, shopID int) error {
	panic("not implemented")
}

// テスト定数
const (
	testOrderUserID = 1
	testOrderShopID = 1
	testOrderID     = 1
)

func TestOrderService_GetUserOrders(t *testing.T) {
	tests := []struct {
		name            string
		userID          int
		setupOrderRepo  func(*OrderRepositoryMockForOrder)
		wantOrders      []models.OrderListResponse
		expectedErrCode apperrors.ErrCode
	}{
		{
			name:   "正常系: ユーザーの注文一覧取得",
			userID: testOrderUserID,
			setupOrderRepo: func(m *OrderRepositoryMockForOrder) {
				m.FindActiveUserOrdersFunc = func(ctx context.Context, dbtx repositories.DBTX, userID int) ([]repositories.OrderWithDetailsDB, error) {
					return []repositories.OrderWithDetailsDB{
						{
							OrderID:      testOrderID,
							ShopName:     "テストショップ",
							Location:     "テスト場所",
							OrderDate:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
							TotalAmount:  300,
							Status:       models.Cooking,
							WaitingCount: 2,
						},
					}, nil
				}
				m.FindItemsByOrderIDsFunc = func(ctx context.Context, dbtx repositories.DBTX, orderIDs []int) (map[int][]models.ItemDetail, error) {
					return map[int][]models.ItemDetail{
						testOrderID: {
							{ItemName: "商品1", Quantity: 2},
							{ItemName: "商品2", Quantity: 1},
						},
					}, nil
				}
			},
			wantOrders: []models.OrderListResponse{
				{
					OrderID:      testOrderID,
					ShopName:     "テストショップ",
					Location:     "テスト場所",
					OrderDate:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
					TotalAmount:  300,
					Status:       models.Cooking.String(),
					WaitingCount: 2,
					Items: []models.ItemDetail{
						{ItemName: "商品1", Quantity: 2},
						{ItemName: "商品2", Quantity: 1},
					},
				},
			},
			expectedErrCode: "",
		},
		{
			name:   "正常系: 注文が0件の場合",
			userID: testOrderUserID,
			setupOrderRepo: func(m *OrderRepositoryMockForOrder) {
				m.FindActiveUserOrdersFunc = func(ctx context.Context, dbtx repositories.DBTX, userID int) ([]repositories.OrderWithDetailsDB, error) {
					return []repositories.OrderWithDetailsDB{}, nil
				}
			},
			wantOrders:      []models.OrderListResponse{},
			expectedErrCode: "",
		},
		{
			name:   "異常系: FindActiveUserOrdersでエラー",
			userID: testOrderUserID,
			setupOrderRepo: func(m *OrderRepositoryMockForOrder) {
				m.FindActiveUserOrdersFunc = func(ctx context.Context, dbtx repositories.DBTX, userID int) ([]repositories.OrderWithDetailsDB, error) {
					return nil, apperrors.Unknown.Wrap(nil, "データベースエラー")
				}
			},
			wantOrders:      nil,
			expectedErrCode: apperrors.Unknown,
		},
		{
			name:   "異常系: FindItemsByOrderIDsでエラー",
			userID: testOrderUserID,
			setupOrderRepo: func(m *OrderRepositoryMockForOrder) {
				m.FindActiveUserOrdersFunc = func(ctx context.Context, dbtx repositories.DBTX, userID int) ([]repositories.OrderWithDetailsDB, error) {
					return []repositories.OrderWithDetailsDB{
						{OrderID: testOrderID, ShopName: "テストショップ", Status: models.Cooking},
					}, nil
				}
				m.FindItemsByOrderIDsFunc = func(ctx context.Context, dbtx repositories.DBTX, orderIDs []int) (map[int][]models.ItemDetail, error) {
					return nil, apperrors.Unknown.Wrap(nil, "アイテム取得エラー")
				}
			},
			wantOrders:      nil,
			expectedErrCode: apperrors.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックセットアップ
			orderRepo := NewOrderRepositoryMockForOrder()
			tt.setupOrderRepo(orderRepo)

			itemRepo := NewItemRepositoryMockForOrder()

			// データベースの設定（DBTX対応）
			mockDB := &sqlx.DB{}

			// サービス初期化（DBTX対応 - NewOrderServiceForTestを使わずに直接NewOrderServiceを使用）
			orderService := services.NewOrderService(orderRepo, itemRepo, mockDB)

			// テスト実行
			gotOrders, err := orderService.GetUserOrders(context.Background(), tt.userID)

			// エラーアサーション
			if tt.expectedErrCode == "" {
				testhelpers.AssertNoError(t, err)
			} else {
				testhelpers.AssertAppError(t, err, tt.expectedErrCode)
			}

			// 正常系の場合のアサーション
			if tt.expectedErrCode == "" {
				if diff := cmp.Diff(tt.wantOrders, gotOrders); diff != "" {
					t.Errorf("%s: orders mismatch (-want +got):\n%s", tt.name, diff)
				}
			}
		})
	}
}

func TestOrderService_GetOrderStatus(t *testing.T) {
	tests := []struct {
		name            string
		userID          int
		orderID         int
		setupOrderRepo  func(*OrderRepositoryMockForOrder)
		wantStatus      *models.OrderStatusResponse
		expectedErrCode apperrors.ErrCode
	}{
		{
			name:    "正常系: 調理中注文のステータス取得",
			userID:  testOrderUserID,
			orderID: testOrderID,
			setupOrderRepo: func(m *OrderRepositoryMockForOrder) {
				m.FindOrderByIDAndUserFunc = func(ctx context.Context, dbtx repositories.DBTX, orderID int, userID int) (*models.Order, error) {
					return &models.Order{
						OrderID:   testOrderID,
						ShopID:    testOrderShopID,
						Status:    models.Cooking,
						OrderDate: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
					}, nil
				}
				m.CountWaitingOrdersFunc = func(ctx context.Context, dbtx repositories.DBTX, shopID int, orderDate time.Time) (int, error) {
					return 3, nil
				}
			},
			wantStatus: &models.OrderStatusResponse{
				OrderID:      testOrderID,
				Status:       models.Cooking.String(),
				WaitingCount: 3,
			},
			expectedErrCode: "",
		},
		{
			name:    "正常系: 調理完了注文のステータス取得（待ち人数0）",
			userID:  testOrderUserID,
			orderID: testOrderID,
			setupOrderRepo: func(m *OrderRepositoryMockForOrder) {
				m.FindOrderByIDAndUserFunc = func(ctx context.Context, dbtx repositories.DBTX, orderID int, userID int) (*models.Order, error) {
					return &models.Order{
						OrderID:   testOrderID,
						ShopID:    testOrderShopID,
						Status:    models.Completed,
						OrderDate: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
					}, nil
				}
			},
			wantStatus: &models.OrderStatusResponse{
				OrderID:      testOrderID,
				Status:       models.Completed.String(),
				WaitingCount: 0,
			},
			expectedErrCode: "",
		},
		{
			name:    "異常系: 注文が見つからない",
			userID:  testOrderUserID,
			orderID: testOrderID,
			setupOrderRepo: func(m *OrderRepositoryMockForOrder) {
				m.FindOrderByIDAndUserFunc = func(ctx context.Context, dbtx repositories.DBTX, orderID int, userID int) (*models.Order, error) {
					return nil, apperrors.NoData.Wrap(nil, "注文が見つかりません")
				}
			},
			wantStatus:      nil,
			expectedErrCode: apperrors.NoData,
		},
		{
			name:    "異常系: CountWaitingOrdersでエラー",
			userID:  testOrderUserID,
			orderID: testOrderID,
			setupOrderRepo: func(m *OrderRepositoryMockForOrder) {
				m.FindOrderByIDAndUserFunc = func(ctx context.Context, dbtx repositories.DBTX, orderID int, userID int) (*models.Order, error) {
					return &models.Order{
						OrderID:   testOrderID,
						ShopID:    testOrderShopID,
						Status:    models.Cooking,
						OrderDate: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
					}, nil
				}
				m.CountWaitingOrdersFunc = func(ctx context.Context, dbtx repositories.DBTX, shopID int, orderDate time.Time) (int, error) {
					return 0, apperrors.Unknown.Wrap(nil, "待ち人数取得エラー")
				}
			},
			wantStatus:      nil,
			expectedErrCode: apperrors.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックセットアップ
			orderRepo := NewOrderRepositoryMockForOrder()
			tt.setupOrderRepo(orderRepo)

			itemRepo := NewItemRepositoryMockForOrder()

			// データベースの設定（DBTX対応）
			mockDB := &sqlx.DB{}

			// サービス初期化（DBTX対応）
			orderService := services.NewOrderService(orderRepo, itemRepo, mockDB)

			// テスト実行
			gotStatus, err := orderService.GetOrderStatus(context.Background(), tt.userID, tt.orderID)

			// エラーアサーション
			if tt.expectedErrCode == "" {
				testhelpers.AssertNoError(t, err)
			} else {
				testhelpers.AssertAppError(t, err, tt.expectedErrCode)
			}

			// 正常系の場合のアサーション
			if tt.expectedErrCode == "" {
				if diff := cmp.Diff(tt.wantStatus, gotStatus); diff != "" {
					t.Errorf("%s: status mismatch (-want +got):\n%s", tt.name, diff)
				}
			}
		})
	}
}

// NOTE: CreateOrderとCreateAuthenticatedOrderのテストは
// トランザクションを使用するため、order_service_integration_test.goに移動しました
