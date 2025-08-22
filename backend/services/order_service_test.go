package services_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/A4-dev-team/mobileorder.git/services"
	"github.com/google/go-cmp/cmp"
)

// ItemRepositoryMock - ItemRepositoryのモック実装
type ItemRepositoryMock struct {
	ValidateAndGetItemsForShopFunc func(ctx context.Context, shopID int, itemIDs []int) (map[int]models.Item, error)
	GetItemListFunc                func(shopID int) ([]models.ItemListResponse, error)
}

func NewItemRepositoryMock() *ItemRepositoryMock {
	return &ItemRepositoryMock{}
}

// インターフェース実装
func (m *ItemRepositoryMock) ValidateAndGetItemsForShop(ctx context.Context, shopID int, itemIDs []int) (map[int]models.Item, error) {
	if m.ValidateAndGetItemsForShopFunc != nil {
		return m.ValidateAndGetItemsForShopFunc(ctx, shopID, itemIDs)
	}
	panic("not implemented")
}

func (m *ItemRepositoryMock) GetItemList(shopID int) ([]models.ItemListResponse, error) {
	if m.GetItemListFunc != nil {
		return m.GetItemListFunc(shopID)
	}
	panic("not implemented")
}

// OrderRepositoryMockForOrder - OrderService用のOrderRepositoryモック
type OrderRepositoryMockForOrder struct {
	CreateOrderFunc          func(ctx context.Context, order *models.Order, items []models.OrderItem) error
	FindActiveUserOrdersFunc func(ctx context.Context, userID int) ([]repositories.OrderWithDetailsDB, error)
	FindItemsByOrderIDsFunc  func(ctx context.Context, orderIDs []int) (map[int][]models.ItemDetail, error)
	FindOrderByIDAndUserFunc func(ctx context.Context, orderID int, userID int) (*models.Order, error)
	CountWaitingOrdersFunc   func(ctx context.Context, shopID int, orderDate time.Time) (int, error)
}

func NewOrderRepositoryMockForOrder() *OrderRepositoryMockForOrder {
	return &OrderRepositoryMockForOrder{}
}

// インターフェース実装
func (m *OrderRepositoryMockForOrder) CreateOrder(ctx context.Context, order *models.Order, items []models.OrderItem) error {
	if m.CreateOrderFunc != nil {
		return m.CreateOrderFunc(ctx, order, items)
	}
	panic("not implemented")
}

func (m *OrderRepositoryMockForOrder) UpdateUserIDByGuestToken(ctx context.Context, guestToken string, userID int) error {
	panic("not implemented")
}

func (m *OrderRepositoryMockForOrder) FindActiveUserOrders(ctx context.Context, userID int) ([]repositories.OrderWithDetailsDB, error) {
	if m.FindActiveUserOrdersFunc != nil {
		return m.FindActiveUserOrdersFunc(ctx, userID)
	}
	panic("not implemented")
}

func (m *OrderRepositoryMockForOrder) FindItemsByOrderIDs(ctx context.Context, orderIDs []int) (map[int][]models.ItemDetail, error) {
	if m.FindItemsByOrderIDsFunc != nil {
		return m.FindItemsByOrderIDsFunc(ctx, orderIDs)
	}
	panic("not implemented")
}

func (m *OrderRepositoryMockForOrder) FindOrderByIDAndUser(ctx context.Context, orderID int, userID int) (*models.Order, error) {
	if m.FindOrderByIDAndUserFunc != nil {
		return m.FindOrderByIDAndUserFunc(ctx, orderID, userID)
	}
	panic("not implemented")
}

func (m *OrderRepositoryMockForOrder) CountWaitingOrders(ctx context.Context, shopID int, orderDate time.Time) (int, error) {
	if m.CountWaitingOrdersFunc != nil {
		return m.CountWaitingOrdersFunc(ctx, shopID, orderDate)
	}
	panic("not implemented")
}

func (m *OrderRepositoryMockForOrder) FindShopOrdersByStatuses(ctx context.Context, shopID int, statuses []models.OrderStatus) ([]repositories.AdminOrderDBResult, error) {
	panic("not implemented")
}

func (m *OrderRepositoryMockForOrder) FindOrderByIDAndShopID(ctx context.Context, orderID int, shopID int) (*models.Order, error) {
	panic("not implemented")
}

func (m *OrderRepositoryMockForOrder) UpdateOrderStatus(ctx context.Context, orderID int, shopID int, newStatus models.OrderStatus) error {
	panic("not implemented")
}

func (m *OrderRepositoryMockForOrder) DeleteOrderByIDAndShopID(ctx context.Context, orderID int, shopID int) error {
	panic("not implemented")
}

// MockTransactionManager テスト用のTransactionManagerモック
type MockTransactionManager struct {
	WithOrderTransactionFunc     func(ctx context.Context, fn func(repositories.OrderRepository) error) error
	WithUserOrderTransactionFunc func(ctx context.Context, fn func(repositories.UserRepository, repositories.OrderRepository) error) error
}

func (m *MockTransactionManager) WithOrderTransaction(ctx context.Context, fn func(repositories.OrderRepository) error) error {
	if m.WithOrderTransactionFunc != nil {
		return m.WithOrderTransactionFunc(ctx, fn)
	}
	panic("not implemented")
}

func (m *MockTransactionManager) WithUserOrderTransaction(ctx context.Context, fn func(repositories.UserRepository, repositories.OrderRepository) error) error {
	if m.WithUserOrderTransactionFunc != nil {
		return m.WithUserOrderTransactionFunc(ctx, fn)
	}
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
				m.FindActiveUserOrdersFunc = func(ctx context.Context, userID int) ([]repositories.OrderWithDetailsDB, error) {
					return []repositories.OrderWithDetailsDB{
						{
							OrderID:      testOrderID,
							ShopName:     "テストショップ",
							Location:     "テスト場所",
							OrderDate:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
							TotalAmount:  300.0,
							Status:       models.Cooking,
							WaitingCount: 2,
						},
					}, nil
				}
				m.FindItemsByOrderIDsFunc = func(ctx context.Context, orderIDs []int) (map[int][]models.ItemDetail, error) {
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
					TotalAmount:  300.0,
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
				m.FindActiveUserOrdersFunc = func(ctx context.Context, userID int) ([]repositories.OrderWithDetailsDB, error) {
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
				m.FindActiveUserOrdersFunc = func(ctx context.Context, userID int) ([]repositories.OrderWithDetailsDB, error) {
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
				m.FindActiveUserOrdersFunc = func(ctx context.Context, userID int) ([]repositories.OrderWithDetailsDB, error) {
					return []repositories.OrderWithDetailsDB{
						{OrderID: testOrderID, ShopName: "テストショップ", Status: models.Cooking},
					}, nil
				}
				m.FindItemsByOrderIDsFunc = func(ctx context.Context, orderIDs []int) (map[int][]models.ItemDetail, error) {
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

			itemRepo := NewItemRepositoryMock()

			// サービス作成 - NewOrderServiceForTestを使用してモックTransactionManagerを渡す
			mockTM := &MockTransactionManager{}
			mockTM.WithOrderTransactionFunc = func(ctx context.Context, fn func(repositories.OrderRepository) error) error {
				// モックorderRepoを直接渡す
				return fn(orderRepo)
			}
			orderService := services.NewOrderServiceForTest(orderRepo, itemRepo, mockTM)

			// テスト実行
			gotOrders, err := orderService.GetUserOrders(context.Background(), tt.userID)

			// エラーアサーション
			if tt.expectedErrCode == "" {
				assertNoError(t, err)
			} else {
				assertAppError(t, err, tt.expectedErrCode)
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
				m.FindOrderByIDAndUserFunc = func(ctx context.Context, orderID int, userID int) (*models.Order, error) {
					return &models.Order{
						OrderID:   testOrderID,
						ShopID:    testOrderShopID,
						Status:    models.Cooking,
						OrderDate: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
					}, nil
				}
				m.CountWaitingOrdersFunc = func(ctx context.Context, shopID int, orderDate time.Time) (int, error) {
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
				m.FindOrderByIDAndUserFunc = func(ctx context.Context, orderID int, userID int) (*models.Order, error) {
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
				m.FindOrderByIDAndUserFunc = func(ctx context.Context, orderID int, userID int) (*models.Order, error) {
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
				m.FindOrderByIDAndUserFunc = func(ctx context.Context, orderID int, userID int) (*models.Order, error) {
					return &models.Order{
						OrderID:   testOrderID,
						ShopID:    testOrderShopID,
						Status:    models.Cooking,
						OrderDate: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
					}, nil
				}
				m.CountWaitingOrdersFunc = func(ctx context.Context, shopID int, orderDate time.Time) (int, error) {
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

			itemRepo := NewItemRepositoryMock()

			// サービス作成 - NewOrderServiceForTestを使用してモックTransactionManagerを渡す
			mockTM := &MockTransactionManager{}
			mockTM.WithOrderTransactionFunc = func(ctx context.Context, fn func(repositories.OrderRepository) error) error {
				// モックorderRepoを直接渡す
				return fn(orderRepo)
			}
			orderService := services.NewOrderServiceForTest(orderRepo, itemRepo, mockTM)

			// テスト実行
			gotStatus, err := orderService.GetOrderStatus(context.Background(), tt.userID, tt.orderID)

			// エラーアサーション
			if tt.expectedErrCode == "" {
				assertNoError(t, err)
			} else {
				assertAppError(t, err, tt.expectedErrCode)
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

func TestOrderService_CreateOrder(t *testing.T) {
	tests := []struct {
		name            string
		shopID          int
		items           []models.OrderItemRequest
		setupOrderRepo  func(*OrderRepositoryMockForOrder)
		setupItemRepo   func(*ItemRepositoryMock)
		wantOrder       *models.Order
		expectedErrCode apperrors.ErrCode
	}{
		{
			name:   "正常系: ゲスト注文作成成功",
			shopID: testOrderShopID,
			items: []models.OrderItemRequest{
				{ItemID: 1, Quantity: 2},
				{ItemID: 2, Quantity: 1},
			},
			setupOrderRepo: func(m *OrderRepositoryMockForOrder) {
				m.CreateOrderFunc = func(ctx context.Context, order *models.Order, items []models.OrderItem) error {
					// CreateOrderが呼ばれた際にOrderIDを設定（DBから返される想定）
					order.OrderID = testOrderID
					return nil
				}
			},
			setupItemRepo: func(m *ItemRepositoryMock) {
				m.ValidateAndGetItemsForShopFunc = func(ctx context.Context, shopID int, itemIDs []int) (map[int]models.Item, error) {
					return map[int]models.Item{
						1: {ItemID: 1, ItemName: "商品1", Price: 100.0, IsAvailable: true},
						2: {ItemID: 2, ItemName: "商品2", Price: 200.0, IsAvailable: true},
					}, nil
				}
			},
			wantOrder: &models.Order{
				OrderID:     testOrderID,
				ShopID:      testOrderShopID,
				TotalAmount: 400.0, // 100*2 + 200*1
				Status:      models.Cooking,
				// GuestOrderTokenとUserIDは動的なので比較対象外
			},
			expectedErrCode: "",
		},
		{
			name:   "異常系: 商品検証エラー",
			shopID: testOrderShopID,
			items: []models.OrderItemRequest{
				{ItemID: 999, Quantity: 1}, // 存在しない商品
			},
			setupOrderRepo: func(m *OrderRepositoryMockForOrder) {
				// CreateOrderは呼ばれない想定
			},
			setupItemRepo: func(m *ItemRepositoryMock) {
				m.ValidateAndGetItemsForShopFunc = func(ctx context.Context, shopID int, itemIDs []int) (map[int]models.Item, error) {
					return nil, apperrors.NoData.Wrap(nil, "商品が見つかりません")
				}
			},
			wantOrder:       nil,
			expectedErrCode: apperrors.NoData,
		},
		{
			name:   "異常系: CreateOrderでエラー",
			shopID: testOrderShopID,
			items: []models.OrderItemRequest{
				{ItemID: 1, Quantity: 1},
			},
			setupOrderRepo: func(m *OrderRepositoryMockForOrder) {
				m.CreateOrderFunc = func(ctx context.Context, order *models.Order, items []models.OrderItem) error {
					return apperrors.Unknown.Wrap(nil, "注文作成エラー")
				}
			},
			setupItemRepo: func(m *ItemRepositoryMock) {
				m.ValidateAndGetItemsForShopFunc = func(ctx context.Context, shopID int, itemIDs []int) (map[int]models.Item, error) {
					return map[int]models.Item{
						1: {ItemID: 1, ItemName: "商品1", Price: 100.0, IsAvailable: true},
					}, nil
				}
			},
			wantOrder:       nil,
			expectedErrCode: apperrors.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックセットアップ
			orderRepo := NewOrderRepositoryMockForOrder()
			tt.setupOrderRepo(orderRepo)

			itemRepo := NewItemRepositoryMock()
			tt.setupItemRepo(itemRepo)

			// TransactionManagerモック
			mockTM := &MockTransactionManager{}
			mockTM.WithOrderTransactionFunc = func(ctx context.Context, fn func(repositories.OrderRepository) error) error {
				return fn(orderRepo)
			}

			orderService := services.NewOrderServiceForTest(orderRepo, itemRepo, mockTM)

			// テスト実行
			gotOrder, err := orderService.CreateOrder(context.Background(), tt.shopID, tt.items)

			// エラーアサーション
			if tt.expectedErrCode == "" {
				assertNoError(t, err)
			} else {
				assertAppError(t, err, tt.expectedErrCode)
			}

			// 正常系の場合のアサーション
			if tt.expectedErrCode == "" {
				if gotOrder.OrderID != tt.wantOrder.OrderID {
					t.Errorf("OrderID mismatch: want=%d, got=%d", tt.wantOrder.OrderID, gotOrder.OrderID)
				}
				if gotOrder.ShopID != tt.wantOrder.ShopID {
					t.Errorf("ShopID mismatch: want=%d, got=%d", tt.wantOrder.ShopID, gotOrder.ShopID)
				}
				if gotOrder.TotalAmount != tt.wantOrder.TotalAmount {
					t.Errorf("TotalAmount mismatch: want=%.2f, got=%.2f", tt.wantOrder.TotalAmount, gotOrder.TotalAmount)
				}
				if gotOrder.Status != tt.wantOrder.Status {
					t.Errorf("Status mismatch: want=%v, got=%v", tt.wantOrder.Status, gotOrder.Status)
				}
				// ゲスト注文の場合、UserIDはnullでGuestOrderTokenが設定されていることを確認
				if gotOrder.UserID.Valid {
					t.Error("ゲスト注文では UserID は null であるべきです")
				}
				if !gotOrder.GuestOrderToken.Valid || gotOrder.GuestOrderToken.String == "" {
					t.Error("ゲスト注文では GuestOrderToken が設定されているべきです")
				}
			}
		})
	}
}

func TestOrderService_CreateAuthenticatedOrder(t *testing.T) {
	tests := []struct {
		name            string
		userID          int
		shopID          int
		items           []models.OrderItemRequest
		setupOrderRepo  func(*OrderRepositoryMockForOrder)
		setupItemRepo   func(*ItemRepositoryMock)
		wantOrder       *models.Order
		expectedErrCode apperrors.ErrCode
	}{
		{
			name:   "正常系: 認証済みユーザー注文作成成功",
			userID: testOrderUserID,
			shopID: testOrderShopID,
			items: []models.OrderItemRequest{
				{ItemID: 1, Quantity: 1},
				{ItemID: 2, Quantity: 2},
			},
			setupOrderRepo: func(m *OrderRepositoryMockForOrder) {
				m.CreateOrderFunc = func(ctx context.Context, order *models.Order, items []models.OrderItem) error {
					// CreateOrderが呼ばれた際にOrderIDを設定（DBから返される想定）
					order.OrderID = testOrderID
					return nil
				}
			},
			setupItemRepo: func(m *ItemRepositoryMock) {
				m.ValidateAndGetItemsForShopFunc = func(ctx context.Context, shopID int, itemIDs []int) (map[int]models.Item, error) {
					return map[int]models.Item{
						1: {ItemID: 1, ItemName: "商品1", Price: 150.0, IsAvailable: true},
						2: {ItemID: 2, ItemName: "商品2", Price: 250.0, IsAvailable: true},
					}, nil
				}
			},
			wantOrder: &models.Order{
				OrderID:     testOrderID,
				UserID:      sql.NullInt64{Int64: int64(testOrderUserID), Valid: true},
				ShopID:      testOrderShopID,
				TotalAmount: 650.0, // 150*1 + 250*2
				Status:      models.Cooking,
			},
			expectedErrCode: "",
		},
		{
			name:   "異常系: 商品検証エラー",
			userID: testOrderUserID,
			shopID: testOrderShopID,
			items: []models.OrderItemRequest{
				{ItemID: 999, Quantity: 1}, // 存在しない商品
			},
			setupOrderRepo: func(m *OrderRepositoryMockForOrder) {
				// CreateOrderは呼ばれない想定
			},
			setupItemRepo: func(m *ItemRepositoryMock) {
				m.ValidateAndGetItemsForShopFunc = func(ctx context.Context, shopID int, itemIDs []int) (map[int]models.Item, error) {
					return nil, apperrors.NoData.Wrap(nil, "商品が見つかりません")
				}
			},
			wantOrder:       nil,
			expectedErrCode: apperrors.NoData,
		},
		{
			name:   "異常系: CreateOrderでエラー",
			userID: testOrderUserID,
			shopID: testOrderShopID,
			items: []models.OrderItemRequest{
				{ItemID: 1, Quantity: 1},
			},
			setupOrderRepo: func(m *OrderRepositoryMockForOrder) {
				m.CreateOrderFunc = func(ctx context.Context, order *models.Order, items []models.OrderItem) error {
					return apperrors.Unknown.Wrap(nil, "注文作成エラー")
				}
			},
			setupItemRepo: func(m *ItemRepositoryMock) {
				m.ValidateAndGetItemsForShopFunc = func(ctx context.Context, shopID int, itemIDs []int) (map[int]models.Item, error) {
					return map[int]models.Item{
						1: {ItemID: 1, ItemName: "商品1", Price: 100.0, IsAvailable: true},
					}, nil
				}
			},
			wantOrder:       nil,
			expectedErrCode: apperrors.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックセットアップ
			orderRepo := NewOrderRepositoryMockForOrder()
			tt.setupOrderRepo(orderRepo)

			itemRepo := NewItemRepositoryMock()
			tt.setupItemRepo(itemRepo)

			// TransactionManagerモック
			mockTM := &MockTransactionManager{}
			mockTM.WithOrderTransactionFunc = func(ctx context.Context, fn func(repositories.OrderRepository) error) error {
				return fn(orderRepo)
			}

			orderService := services.NewOrderServiceForTest(orderRepo, itemRepo, mockTM)

			// テスト実行
			gotOrder, err := orderService.CreateAuthenticatedOrder(context.Background(), tt.userID, tt.shopID, tt.items)

			// エラーアサーション
			if tt.expectedErrCode == "" {
				assertNoError(t, err)
			} else {
				assertAppError(t, err, tt.expectedErrCode)
			}

			// 正常系の場合のアサーション
			if tt.expectedErrCode == "" {
				if gotOrder.OrderID != tt.wantOrder.OrderID {
					t.Errorf("OrderID mismatch: want=%d, got=%d", tt.wantOrder.OrderID, gotOrder.OrderID)
				}
				if gotOrder.UserID != tt.wantOrder.UserID {
					t.Errorf("UserID mismatch: want=%v, got=%v", tt.wantOrder.UserID, gotOrder.UserID)
				}
				if gotOrder.ShopID != tt.wantOrder.ShopID {
					t.Errorf("ShopID mismatch: want=%d, got=%d", tt.wantOrder.ShopID, gotOrder.ShopID)
				}
				if gotOrder.TotalAmount != tt.wantOrder.TotalAmount {
					t.Errorf("TotalAmount mismatch: want=%.2f, got=%.2f", tt.wantOrder.TotalAmount, gotOrder.TotalAmount)
				}
				if gotOrder.Status != tt.wantOrder.Status {
					t.Errorf("Status mismatch: want=%v, got=%v", tt.wantOrder.Status, gotOrder.Status)
				}
				// 認証済み注文の場合、GuestOrderTokenはnullであることを確認
				if gotOrder.GuestOrderToken.Valid {
					t.Error("認証済み注文では GuestOrderToken は null であるべきです")
				}
			}
		})
	}
}
