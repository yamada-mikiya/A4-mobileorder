package repositories

import (
	"context"
	"fmt"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
)

type ShopRepository interface {
	FindShopIDByAdminID(ctx context.Context, dbtx DBTX, userID int) (int, error)
}

type shopRepository struct{}

func NewShopRepository() ShopRepository {
	return &shopRepository{}
}

func (r *shopRepository) FindShopIDByAdminID(ctx context.Context, dbtx DBTX, userID int) (int, error) {
	var shopIDs []int
	query := `
		SELECT s.shop_id FROM shops s
		INNER JOIN shop_staff ss ON s.shop_id = ss.shop_id
		WHERE ss.user_id = $1
	`
	err := dbtx.SelectContext(ctx, &shopIDs, query, userID)
	if err != nil {
		return 0, apperrors.GetDataFailed.Wrap(err, "管理者所属店舗の取得に失敗しました。")
	}

	switch len(shopIDs) {
	case 0:
		return 0, apperrors.NoData.Wrap(nil, "この管理者アカウントに紐づく店舗が見つかりません。")
	case 1:
		return shopIDs[0], nil
	default:
		err := fmt.Errorf("data inconsistency: user_id %d is associated with multiple shops", userID)
		return 0, apperrors.Unknown.Wrap(err, "データ不整合が発生しました。")
	}
}
