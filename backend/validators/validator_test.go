package validators

import (
	"testing"

	"github.com/A4-dev-team/mobileorder.git/models"
)

func TestValidator(t *testing.T) {
	t.Run("CreateOrderRequest", func(t *testing.T) {
		tests := []struct {
			name      string
			testData  models.CreateOrderRequest
			wantError bool
		}{
			{
				name: "正常系: 有効な注文リクエスト",
				testData: models.CreateOrderRequest{
					Items: []models.OrderItemRequest{
						{ItemID: 1, Quantity: 2},
						{ItemID: 2, Quantity: 1},
					},
				},
				wantError: false,
			},
			{
				name: "異常系: 商品リストが空",
				testData: models.CreateOrderRequest{
					Items: []models.OrderItemRequest{},
				},
				wantError: true,
			},
			{
				name: "異常系: 商品リストがnil",
				testData: models.CreateOrderRequest{
					Items: nil,
				},
				wantError: true,
			},
			{
				name: "異常系: 商品IDが無効（0以下）",
				testData: models.CreateOrderRequest{
					Items: []models.OrderItemRequest{
						{ItemID: 0, Quantity: 1},
					},
				},
				wantError: true,
			},
			{
				name: "異常系: 数量が無効（0以下）",
				testData: models.CreateOrderRequest{
					Items: []models.OrderItemRequest{
						{ItemID: 1, Quantity: 0},
					},
				},
				wantError: true,
			},
			{
				name: "異常系: 負の数量",
				testData: models.CreateOrderRequest{
					Items: []models.OrderItemRequest{
						{ItemID: 1, Quantity: -1},
					},
				},
				wantError: true,
			},
			{
				name: "異常系: 複数の不正な商品アイテム",
				testData: models.CreateOrderRequest{
					Items: []models.OrderItemRequest{
						{ItemID: 0, Quantity: 1},  // 無効なItemID
						{ItemID: 1, Quantity: -1}, // 無効なQuantity
						{ItemID: -1, Quantity: 0}, // 両方無効
					},
				},
				wantError: true,
			},
		}

		validator := NewValidator[models.CreateOrderRequest]()

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validator.Validate(tt.testData)
				if (err != nil) != tt.wantError {
					t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
				}
			})
		}
	})

	t.Run("OrderItemRequest", func(t *testing.T) {
		tests := []struct {
			name      string
			testData  models.OrderItemRequest
			wantError bool
		}{
			{
				name: "正常系: 有効な商品アイテム",
				testData: models.OrderItemRequest{
					ItemID:   1,
					Quantity: 2,
				},
				wantError: false,
			},
			{
				name: "正常系: 最小値での有効な商品アイテム",
				testData: models.OrderItemRequest{
					ItemID:   1,
					Quantity: 1,
				},
				wantError: false,
			},
			{
				name: "正常系: 大きな値での有効な商品アイテム",
				testData: models.OrderItemRequest{
					ItemID:   999999,
					Quantity: 100,
				},
				wantError: false,
			},
			{
				name: "異常系: ItemIDが0",
				testData: models.OrderItemRequest{
					ItemID:   0,
					Quantity: 1,
				},
				wantError: true,
			},
			{
				name: "異常系: ItemIDが負の値",
				testData: models.OrderItemRequest{
					ItemID:   -1,
					Quantity: 1,
				},
				wantError: true,
			},
			{
				name: "異常系: Quantityが0",
				testData: models.OrderItemRequest{
					ItemID:   1,
					Quantity: 0,
				},
				wantError: true,
			},
			{
				name: "異常系: Quantityが負の値",
				testData: models.OrderItemRequest{
					ItemID:   1,
					Quantity: -1,
				},
				wantError: true,
			},
			{
				name: "異常系: 両方のフィールドが無効",
				testData: models.OrderItemRequest{
					ItemID:   -1,
					Quantity: -1,
				},
				wantError: true,
			},
		}

		validator := NewValidator[models.OrderItemRequest]()

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validator.Validate(tt.testData)
				if (err != nil) != tt.wantError {
					t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
				}
			})
		}
	})

	t.Run("AuthenticateRequest", func(t *testing.T) {
		tests := []struct {
			name      string
			testData  models.AuthenticateRequest
			wantError bool
		}{
			{
				name: "正常系: 有効なメールアドレスのみ",
				testData: models.AuthenticateRequest{
					Email:           "user@example.com",
					GuestOrderToken: "",
				},
				wantError: false,
			},
			{
				name: "正常系: 有効なメールアドレスとゲストトークン",
				testData: models.AuthenticateRequest{
					Email:           "user@example.com",
					GuestOrderToken: "15ff4999-2cfd-41f3-b744-926e7c5c7a0e",
				},
				wantError: false,
			},
			{
				name: "正常系: ゲストトークンが空文字（omitempty）",
				testData: models.AuthenticateRequest{
					Email:           "test.user@domain.co.jp",
					GuestOrderToken: "",
				},
				wantError: false,
			},
			{
				name: "異常系: メールアドレスが空",
				testData: models.AuthenticateRequest{
					Email:           "",
					GuestOrderToken: "",
				},
				wantError: true,
			},
			{
				name: "異常系: 無効なメールアドレス（@なし）",
				testData: models.AuthenticateRequest{
					Email:           "invalid-email",
					GuestOrderToken: "",
				},
				wantError: true,
			},
			{
				name: "異常系: 無効なメールアドレス（ドメインなし）",
				testData: models.AuthenticateRequest{
					Email:           "user@",
					GuestOrderToken: "",
				},
				wantError: true,
			},
			{
				name: "異常系: 無効なUUID形式のゲストトークン",
				testData: models.AuthenticateRequest{
					Email:           "user@example.com",
					GuestOrderToken: "invalid-uuid",
				},
				wantError: true,
			},
			{
				name: "異常系: 不完全なUUID形式のゲストトークン",
				testData: models.AuthenticateRequest{
					Email:           "user@example.com",
					GuestOrderToken: "15ff4999-2cfd-41f3-b744",
				},
				wantError: true,
			},
			{
				name: "異常系: UUID v4以外のバージョン",
				testData: models.AuthenticateRequest{
					Email:           "user@example.com",
					GuestOrderToken: "15ff4999-2cfd-31f3-b744-926e7c5c7a0e", // v3 UUID
				},
				wantError: true,
			},
			{
				name: "異常系: 複数フィールドが無効",
				testData: models.AuthenticateRequest{
					Email:           "invalid-email",
					GuestOrderToken: "invalid-uuid",
				},
				wantError: true,
			},
		}

		validator := NewValidator[models.AuthenticateRequest]()

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validator.Validate(tt.testData)
				if (err != nil) != tt.wantError {
					t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
				}
			})
		}
	})
}
