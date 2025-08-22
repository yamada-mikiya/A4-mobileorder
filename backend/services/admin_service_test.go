package services_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/A4-dev-team/mobileorder.git/services"
	"github.com/google/go-cmp/cmp"
)

// OrderRepositoryMock - OrderRepositoryのモック実装
type OrderRepositoryMock struct {
	FindShopOrdersByStatusesFunc func(ctx context.Context, shopID int, statuses []models.OrderStatus) ([]repositories.AdminOrderDBResult, error)
	FindItemsByOrderIDsFunc      func(ctx context.Context, orderIDs []int) (map[int][]models.ItemDetail, error)
	FindOrderByIDAndShopIDFunc   func(ctx context.Context, orderID int, shopID int) (*models.Order, error)
	UpdateOrderStatusFunc        func(ctx context.Context, orderID int, shopID int, newStatus models.OrderStatus) error
	DeleteOrderByIDAndShopIDFunc func(ctx context.Context, orderID int, shopID int) error
}

func NewOrderRepositoryMock() *OrderRepositoryMock {
	return &OrderRepositoryMock{}
}

// インターフェース実装
func (m *OrderRepositoryMock) CreateOrder(ctx context.Context, order *models.Order, items []models.OrderItem) error {
	panic("not implemented")
}

func (m *OrderRepositoryMock) UpdateUserIDByGuestToken(ctx context.Context, guestToken string, userID int) error {
	panic("not implemented")
}

func (m *OrderRepositoryMock) FindActiveUserOrders(ctx context.Context, userID int) ([]repositories.OrderWithDetailsDB, error) {
	panic("not implemented")
}

func (m *OrderRepositoryMock) FindItemsByOrderIDs(ctx context.Context, orderIDs []int) (map[int][]models.ItemDetail, error) {
	return m.FindItemsByOrderIDsFunc(ctx, orderIDs)
}

func (m *OrderRepositoryMock) FindOrderByIDAndUser(ctx context.Context, orderID int, userID int) (*models.Order, error) {
	panic("not implemented")
}

func (m *OrderRepositoryMock) CountWaitingOrders(ctx context.Context, shopID int, orderDate time.Time) (int, error) {
	panic("not implemented")
}

func (m *OrderRepositoryMock) FindShopOrdersByStatuses(ctx context.Context, shopID int, statuses []models.OrderStatus) ([]repositories.AdminOrderDBResult, error) {
	return m.FindShopOrdersByStatusesFunc(ctx, shopID, statuses)
}

func (m *OrderRepositoryMock) FindOrderByIDAndShopID(ctx context.Context, orderID int, shopID int) (*models.Order, error) {
	return m.FindOrderByIDAndShopIDFunc(ctx, orderID, shopID)
}

func (m *OrderRepositoryMock) UpdateOrderStatus(ctx context.Context, orderID int, shopID int, newStatus models.OrderStatus) error {
	return m.UpdateOrderStatusFunc(ctx, orderID, shopID, newStatus)
}

func (m *OrderRepositoryMock) DeleteOrderByIDAndShopID(ctx context.Context, orderID int, shopID int) error {
	return m.DeleteOrderByIDAndShopIDFunc(ctx, orderID, shopID)
}

// テスト用データ生成関数
func createTestAdminOrderDBResult(orderID int, email string, totalAmount float64, status models.OrderStatus) repositories.AdminOrderDBResult {
	var customerEmail sql.NullString
	if email != "" {
		customerEmail = sql.NullString{String: email, Valid: true}
	}

	return repositories.AdminOrderDBResult{
		OrderID:       orderID,
		CustomerEmail: customerEmail,
		OrderDate:     time.Date(2025, 8, 16, 10, 0, 0, 0, time.UTC),
		TotalAmount:   totalAmount,
		Status:        status,
	}
}

func createTestItemDetails() map[int][]models.ItemDetail {
	return map[int][]models.ItemDetail{
		1: {
			{ItemName: "ハンバーガー", Quantity: 2},
			{ItemName: "ポテト", Quantity: 1},
		},
		2: {
			{ItemName: "チキンサンド", Quantity: 1},
		},
	}
}

// TestAdminService_GetCookingOrders - GetCookingOrdersメソッドのテスト
func TestAdminService_GetCookingOrders(t *testing.T) {
	tests := []struct {
		name                    string
		shopID                  int
		mockFindShopOrdersFunc  func(ctx context.Context, shopID int, statuses []models.OrderStatus) ([]repositories.AdminOrderDBResult, error)
		mockFindItemsFunc       func(ctx context.Context, orderIDs []int) (map[int][]models.ItemDetail, error)
		expectedOrdersCount     int
		expectedFirstOrderEmail *string
		expectedErrCode         apperrors.ErrCode
	}{
		{
			name:   "正常系: 調理中の注文が複数件取得できる",
			shopID: 1,
			mockFindShopOrdersFunc: func(ctx context.Context, shopID int, statuses []models.OrderStatus) ([]repositories.AdminOrderDBResult, error) {
				// 期待するstatusesの値をチェック
				expectedStatuses := []models.OrderStatus{models.Cooking}
				if len(statuses) != 1 || statuses[0] != expectedStatuses[0] {
					t.Errorf("Expected statuses %v, got %v", expectedStatuses, statuses)
				}

				return []repositories.AdminOrderDBResult{
					createTestAdminOrderDBResult(1, "customer1@test.com", 1500.0, models.Cooking),
					createTestAdminOrderDBResult(2, "", 800.0, models.Cooking),
				}, nil
			},
			mockFindItemsFunc: func(ctx context.Context, orderIDs []int) (map[int][]models.ItemDetail, error) {
				return createTestItemDetails(), nil
			},
			expectedOrdersCount: 2,
			expectedFirstOrderEmail: func() *string {
				email := "customer1@test.com"
				return &email
			}(),
			expectedErrCode: "",
		},
		{
			name:   "正常系: 調理中の注文が0件の場合は空のスライスを返す",
			shopID: 1,
			mockFindShopOrdersFunc: func(ctx context.Context, shopID int, statuses []models.OrderStatus) ([]repositories.AdminOrderDBResult, error) {
				return []repositories.AdminOrderDBResult{}, nil
			},
			mockFindItemsFunc:       nil,
			expectedOrdersCount:     0,
			expectedFirstOrderEmail: nil,
			expectedErrCode:         "",
		},
		{
			name:   "異常系: FindShopOrdersByStatusesでデータベースエラーが発生",
			shopID: 1,
			mockFindShopOrdersFunc: func(ctx context.Context, shopID int, statuses []models.OrderStatus) ([]repositories.AdminOrderDBResult, error) {
				return nil, apperrors.GetDataFailed.Wrap(errors.New("database connection failed"), "データベース接続エラー")
			},
			mockFindItemsFunc:       nil,
			expectedOrdersCount:     0,
			expectedFirstOrderEmail: nil,
			expectedErrCode:         apperrors.GetDataFailed,
		},
		{
			name:   "異常系: FindItemsByOrderIDsでエラーが発生",
			shopID: 1,
			mockFindShopOrdersFunc: func(ctx context.Context, shopID int, statuses []models.OrderStatus) ([]repositories.AdminOrderDBResult, error) {
				return []repositories.AdminOrderDBResult{
					createTestAdminOrderDBResult(1, "customer1@test.com", 1500.0, models.Cooking),
				}, nil
			},
			mockFindItemsFunc: func(ctx context.Context, orderIDs []int) (map[int][]models.ItemDetail, error) {
				return nil, apperrors.GetDataFailed.Wrap(errors.New("items query failed"), "アイテム取得エラー")
			},
			expectedOrdersCount:     0,
			expectedFirstOrderEmail: nil,
			expectedErrCode:         apperrors.GetDataFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの設定
			mockRepo := NewOrderRepositoryMock()
			mockRepo.FindShopOrdersByStatusesFunc = tt.mockFindShopOrdersFunc
			if tt.mockFindItemsFunc != nil {
				mockRepo.FindItemsByOrderIDsFunc = tt.mockFindItemsFunc
			}

			// サービス初期化
			mockTxm := services.NewMockTransactionManager(mockRepo)
			adminService := services.NewAdminServiceForTest(mockRepo, mockTxm)

			// テスト実行
			ctx := context.Background()
			result, err := adminService.GetCookingOrders(ctx, tt.shopID)

			// エラーの検証
			if tt.expectedErrCode != "" {
				assertAppError(t, err, tt.expectedErrCode)
				return
			}

			// 正常系の検証
			assertNoError(t, err)

			if len(result) != tt.expectedOrdersCount {
				t.Errorf("期待される注文数 = %d, 実際 = %d", tt.expectedOrdersCount, len(result))
			}

			if tt.expectedOrdersCount > 0 {
				// 最初の注文のメールアドレスをチェック
				if tt.expectedFirstOrderEmail == nil && result[0].CustomerEmail != nil {
					t.Errorf("CustomerEmailはnilが期待されましたが、%v が返されました", *result[0].CustomerEmail)
				} else if tt.expectedFirstOrderEmail != nil {
					if result[0].CustomerEmail == nil {
						t.Errorf("CustomerEmailは%vが期待されましたが、nilが返されました", *tt.expectedFirstOrderEmail)
					} else if *result[0].CustomerEmail != *tt.expectedFirstOrderEmail {
						t.Errorf("期待されるCustomerEmail = %v, 実際 = %v", *tt.expectedFirstOrderEmail, *result[0].CustomerEmail)
					}
				}

				// ステータスの文字列変換をチェック
				if result[0].Status != "cooking" {
					t.Errorf("期待されるStatus = cooking, 実際 = %v", result[0].Status)
				}
			}
		})
	}
}

// TestAdminService_GetCompletedOrders - GetCompletedOrdersメソッドのテスト
func TestAdminService_GetCompletedOrders(t *testing.T) {
	tests := []struct {
		name                   string
		shopID                 int
		mockFindShopOrdersFunc func(ctx context.Context, shopID int, statuses []models.OrderStatus) ([]repositories.AdminOrderDBResult, error)
		mockFindItemsFunc      func(ctx context.Context, orderIDs []int) (map[int][]models.ItemDetail, error)
		expectedOrdersCount    int
		expectedErrCode        apperrors.ErrCode
	}{
		{
			name:   "正常系: 調理完了の注文が取得できる",
			shopID: 1,
			mockFindShopOrdersFunc: func(ctx context.Context, shopID int, statuses []models.OrderStatus) ([]repositories.AdminOrderDBResult, error) {
				// 期待するstatusesの値をチェック
				expectedStatuses := []models.OrderStatus{models.Completed}
				if len(statuses) != 1 || statuses[0] != expectedStatuses[0] {
					t.Errorf("Expected statuses %v, got %v", expectedStatuses, statuses)
				}

				return []repositories.AdminOrderDBResult{
					createTestAdminOrderDBResult(3, "customer3@test.com", 2000.0, models.Completed),
				}, nil
			},
			mockFindItemsFunc: func(ctx context.Context, orderIDs []int) (map[int][]models.ItemDetail, error) {
				return map[int][]models.ItemDetail{
					3: {{ItemName: "ピザ", Quantity: 1}},
				}, nil
			},
			expectedOrdersCount: 1,
			expectedErrCode:     "",
		},
		{
			name:   "正常系: 調理完了の注文が0件",
			shopID: 1,
			mockFindShopOrdersFunc: func(ctx context.Context, shopID int, statuses []models.OrderStatus) ([]repositories.AdminOrderDBResult, error) {
				return []repositories.AdminOrderDBResult{}, nil
			},
			mockFindItemsFunc:   nil,
			expectedOrdersCount: 0,
			expectedErrCode:     "",
		},
		{
			name:   "異常系: リポジトリでエラーが発生",
			shopID: 1,
			mockFindShopOrdersFunc: func(ctx context.Context, shopID int, statuses []models.OrderStatus) ([]repositories.AdminOrderDBResult, error) {
				return nil, apperrors.Unknown.Wrap(errors.New("unexpected error"), "予期しないエラー")
			},
			mockFindItemsFunc:   nil,
			expectedOrdersCount: 0,
			expectedErrCode:     apperrors.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの設定
			mockRepo := NewOrderRepositoryMock()
			mockRepo.FindShopOrdersByStatusesFunc = tt.mockFindShopOrdersFunc
			if tt.mockFindItemsFunc != nil {
				mockRepo.FindItemsByOrderIDsFunc = tt.mockFindItemsFunc
			}

			// サービス初期化
			mockTxm := services.NewMockTransactionManager(mockRepo)
			adminService := services.NewAdminServiceForTest(mockRepo, mockTxm)

			// テスト実行
			ctx := context.Background()
			result, err := adminService.GetCompletedOrders(ctx, tt.shopID)

			// エラーの検証
			if tt.expectedErrCode != "" {
				assertAppError(t, err, tt.expectedErrCode)
				return
			}

			// 正常系の検証
			assertNoError(t, err)

			if len(result) != tt.expectedOrdersCount {
				t.Errorf("期待される注文数 = %d, 実際 = %d", tt.expectedOrdersCount, len(result))
			}

			if tt.expectedOrdersCount > 0 {
				// ステータスの文字列変換をチェック
				if result[0].Status != "completed" {
					t.Errorf("期待されるStatus = completed, 実際 = %v", result[0].Status)
				}
			}
		})
	}
}

// TestAdminService_AssembleResponse_Integration - レスポンス組み立ての統合テスト
func TestAdminService_AssembleResponse_Integration(t *testing.T) {
	// モックの設定
	mockRepo := NewOrderRepositoryMock()

	// 調理中の注文データ
	mockRepo.FindShopOrdersByStatusesFunc = func(ctx context.Context, shopID int, statuses []models.OrderStatus) ([]repositories.AdminOrderDBResult, error) {
		return []repositories.AdminOrderDBResult{
			{
				OrderID:       100,
				CustomerEmail: sql.NullString{String: "integration@test.com", Valid: true},
				OrderDate:     time.Date(2025, 8, 16, 14, 30, 0, 0, time.UTC),
				TotalAmount:   2500.0,
				Status:        models.Cooking,
			},
		}, nil
	}

	// アイテム詳細データ
	mockRepo.FindItemsByOrderIDsFunc = func(ctx context.Context, orderIDs []int) (map[int][]models.ItemDetail, error) {
		expectedOrderIDs := []int{100}
		if !cmp.Equal(orderIDs, expectedOrderIDs) {
			t.Errorf("期待されるorderIDs = %v, 実際 = %v", expectedOrderIDs, orderIDs)
		}

		return map[int][]models.ItemDetail{
			100: {
				{ItemName: "デラックスバーガー", Quantity: 2},
				{ItemName: "フライドポテト", Quantity: 2},
				{ItemName: "コーラ", Quantity: 2},
			},
		}, nil
	}

	// サービス初期化
	mockTxm := services.NewMockTransactionManager(mockRepo)
	adminService := services.NewAdminServiceForTest(mockRepo, mockTxm)

	// テスト実行
	ctx := context.Background()
	result, err := adminService.GetCookingOrders(ctx, 1)

	// 検証
	if err != nil {
		t.Fatalf("エラーが発生しました: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("期待される注文数 = 1, 実際 = %d", len(result))
	}

	order := result[0]

	// レスポンスの詳細検証
	if order.OrderID != 100 {
		t.Errorf("期待されるOrderID = 100, 実際 = %d", order.OrderID)
	}

	if order.CustomerEmail == nil || *order.CustomerEmail != "integration@test.com" {
		t.Errorf("期待されるCustomerEmail = integration@test.com, 実際 = %v", order.CustomerEmail)
	}

	if order.TotalAmount != 2500.0 {
		t.Errorf("期待されるTotalAmount = 2500.0, 実際 = %f", order.TotalAmount)
	}

	if order.Status != "cooking" {
		t.Errorf("期待されるStatus = cooking, 実際 = %s", order.Status)
	}

	expectedItems := []models.ItemDetail{
		{ItemName: "デラックスバーガー", Quantity: 2},
		{ItemName: "フライドポテト", Quantity: 2},
		{ItemName: "コーラ", Quantity: 2},
	}

	if !cmp.Equal(order.Items, expectedItems) {
		t.Errorf("期待されるItems = %+v, 実際 = %+v", expectedItems, order.Items)
	}

	// 日付の検証
	expectedDate := time.Date(2025, 8, 16, 14, 30, 0, 0, time.UTC)
	if !order.OrderDate.Equal(expectedDate) {
		t.Errorf("期待されるOrderDate = %v, 実際 = %v", expectedDate, order.OrderDate)
	}
}

// TestAdminService_UpdateOrderStatus - UpdateOrderStatusメソッドのテスト
func TestAdminService_UpdateOrderStatus(t *testing.T) {
	tests := []struct {
		name                      string
		adminShopID               int
		targetOrderID             int
		mockFindOrderFunc         func(ctx context.Context, orderID int, shopID int) (*models.Order, error)
		mockUpdateOrderStatusFunc func(ctx context.Context, orderID int, shopID int, newStatus models.OrderStatus) error
		expectedErrCode           apperrors.ErrCode
		expectedNextStatus        models.OrderStatus
	}{
		{
			name:          "正常系: 調理中→調理完了にステータス更新",
			adminShopID:   1,
			targetOrderID: 100,
			mockFindOrderFunc: func(ctx context.Context, orderID int, shopID int) (*models.Order, error) {
				return &models.Order{
					OrderID: 100,
					ShopID:  1,
					Status:  models.Cooking,
				}, nil
			},
			mockUpdateOrderStatusFunc: func(ctx context.Context, orderID int, shopID int, newStatus models.OrderStatus) error {
				// 期待される引数をチェック
				if orderID != 100 || shopID != 1 || newStatus != models.Completed {
					t.Errorf("UpdateOrderStatus called with unexpected args: orderID=%d, shopID=%d, status=%v", orderID, shopID, newStatus)
				}
				return nil
			},
			expectedErrCode:    "",
			expectedNextStatus: models.Completed,
		},
		{
			name:          "正常系: 調理完了→受け渡し完了にステータス更新",
			adminShopID:   1,
			targetOrderID: 101,
			mockFindOrderFunc: func(ctx context.Context, orderID int, shopID int) (*models.Order, error) {
				return &models.Order{
					OrderID: 101,
					ShopID:  1,
					Status:  models.Completed,
				}, nil
			},
			mockUpdateOrderStatusFunc: func(ctx context.Context, orderID int, shopID int, newStatus models.OrderStatus) error {
				// 期待される引数をチェック
				if orderID != 101 || shopID != 1 || newStatus != models.Handed {
					t.Errorf("UpdateOrderStatus called with unexpected args: orderID=%d, shopID=%d, status=%v", orderID, shopID, newStatus)
				}
				return nil
			},
			expectedErrCode:    "",
			expectedNextStatus: models.Handed,
		},
		{
			name:          "異常系: 注文が存在しない場合はFindOrderByIDAndShopIDでエラー",
			adminShopID:   1,
			targetOrderID: 999,
			mockFindOrderFunc: func(ctx context.Context, orderID int, shopID int) (*models.Order, error) {
				return nil, apperrors.NoData.Wrap(nil, "注文が見つかりません")
			},
			mockUpdateOrderStatusFunc: nil, // 呼ばれない
			expectedErrCode:           apperrors.NoData,
		},
		{
			name:          "異常系: 受け渡し完了済みの注文は更新できない",
			adminShopID:   1,
			targetOrderID: 102,
			mockFindOrderFunc: func(ctx context.Context, orderID int, shopID int) (*models.Order, error) {
				return &models.Order{
					OrderID: 102,
					ShopID:  1,
					Status:  models.Handed,
				}, nil
			},
			mockUpdateOrderStatusFunc: nil, // 呼ばれない
			expectedErrCode:           apperrors.Conflict,
		},
		{
			name:          "異常系: UpdateOrderStatusでデータベースエラーが発生",
			adminShopID:   1,
			targetOrderID: 103,
			mockFindOrderFunc: func(ctx context.Context, orderID int, shopID int) (*models.Order, error) {
				return &models.Order{
					OrderID: 103,
					ShopID:  1,
					Status:  models.Cooking,
				}, nil
			},
			mockUpdateOrderStatusFunc: func(ctx context.Context, orderID int, shopID int, newStatus models.OrderStatus) error {
				return apperrors.UpdateDataFailed.Wrap(errors.New("database error"), "ステータス更新に失敗しました")
			},
			expectedErrCode: apperrors.UpdateDataFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの設定
			mockRepo := NewOrderRepositoryMock()
			mockRepo.FindOrderByIDAndShopIDFunc = tt.mockFindOrderFunc
			if tt.mockUpdateOrderStatusFunc != nil {
				mockRepo.UpdateOrderStatusFunc = tt.mockUpdateOrderStatusFunc
			}

			// サービス初期化
			mockTxm := services.NewMockTransactionManager(mockRepo)
			adminService := services.NewAdminServiceForTest(mockRepo, mockTxm)

			// テスト実行
			ctx := context.Background()
			err := adminService.UpdateOrderStatus(ctx, tt.adminShopID, tt.targetOrderID)

			// エラーの検証
			if tt.expectedErrCode != "" {
				assertAppError(t, err, tt.expectedErrCode)
				return
			}

			// 正常系の検証
			assertNoError(t, err)
		})
	}
}

// TestAdminService_DeleteOrder - DeleteOrderメソッドのテスト
func TestAdminService_DeleteOrder(t *testing.T) {
	tests := []struct {
		name                string
		adminShopID         int
		targetOrderID       int
		mockFindOrderFunc   func(ctx context.Context, orderID int, shopID int) (*models.Order, error)
		mockDeleteOrderFunc func(ctx context.Context, orderID int, shopID int) error
		expectedErrCode     apperrors.ErrCode
	}{
		{
			name:          "正常系: 注文を正常に削除できる",
			adminShopID:   1,
			targetOrderID: 100,
			mockFindOrderFunc: func(ctx context.Context, orderID int, shopID int) (*models.Order, error) {
				return &models.Order{
					OrderID: 100,
					ShopID:  1,
					Status:  models.Cooking,
				}, nil
			},
			mockDeleteOrderFunc: func(ctx context.Context, orderID int, shopID int) error {
				// 期待される引数をチェック
				if orderID != 100 || shopID != 1 {
					t.Errorf("DeleteOrderByIDAndShopID called with unexpected args: orderID=%d, shopID=%d", orderID, shopID)
				}
				return nil
			},
			expectedErrCode: "",
		},
		{
			name:          "異常系: 注文が存在しない場合はFindOrderByIDAndShopIDでエラー",
			adminShopID:   1,
			targetOrderID: 999,
			mockFindOrderFunc: func(ctx context.Context, orderID int, shopID int) (*models.Order, error) {
				return nil, apperrors.NoData.Wrap(nil, "注文が見つかりません")
			},
			mockDeleteOrderFunc: nil, // 呼ばれない
			expectedErrCode:     apperrors.NoData,
		},
		{
			name:          "異常系: 別の店舗の注文は削除できない",
			adminShopID:   1,
			targetOrderID: 100,
			mockFindOrderFunc: func(ctx context.Context, orderID int, shopID int) (*models.Order, error) {
				return nil, apperrors.NoData.Wrap(nil, "権限がありません")
			},
			mockDeleteOrderFunc: nil, // 呼ばれない
			expectedErrCode:     apperrors.NoData,
		},
		{
			name:          "異常系: DeleteOrderByIDAndShopIDでデータベースエラーが発生",
			adminShopID:   1,
			targetOrderID: 101,
			mockFindOrderFunc: func(ctx context.Context, orderID int, shopID int) (*models.Order, error) {
				return &models.Order{
					OrderID: 101,
					ShopID:  1,
					Status:  models.Completed,
				}, nil
			},
			mockDeleteOrderFunc: func(ctx context.Context, orderID int, shopID int) error {
				return apperrors.DeleteDataFailed.Wrap(errors.New("database error"), "注文の削除に失敗しました")
			},
			expectedErrCode: apperrors.DeleteDataFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの設定
			mockRepo := NewOrderRepositoryMock()
			mockRepo.FindOrderByIDAndShopIDFunc = tt.mockFindOrderFunc
			if tt.mockDeleteOrderFunc != nil {
				mockRepo.DeleteOrderByIDAndShopIDFunc = tt.mockDeleteOrderFunc
			}

			// サービス初期化
			mockTxm := services.NewMockTransactionManager(mockRepo)
			adminService := services.NewAdminServiceForTest(mockRepo, mockTxm)

			// テスト実行
			ctx := context.Background()
			err := adminService.DeleteOrder(ctx, tt.adminShopID, tt.targetOrderID)

			// エラーの検証
			if tt.expectedErrCode != "" {
				assertAppError(t, err, tt.expectedErrCode)
				return
			}

			// 正常系の検証
			assertNoError(t, err)
		})
	}
}
