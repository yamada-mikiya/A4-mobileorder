package services

import (
	"context"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/jmoiron/sqlx"
)

type AdminServicer interface {
	GetCookingOrders(ctx context.Context, shopID int) ([]models.AdminOrderResponse, error)
	GetCompletedOrders(ctx context.Context, shopID int) ([]models.AdminOrderResponse, error)
	UpdateOrderStatus(ctx context.Context, adminShopID int, targetOrderID int) error
	UpdateItemAvailability(ctx context.Context, itemID int, isAvailable bool) error
	DeleteOrder(ctx context.Context, adminShopID int, targetOrderID int) error
}

type adminService struct {
	orr repositories.OrderRepository
	itr repositories.ItemRepository
	db  *sqlx.DB
}

func NewAdminService(orr repositories.OrderRepository, itr repositories.ItemRepository, db *sqlx.DB) AdminServicer {
	return &adminService{
		orr: orr,
		itr: itr,
		db:  db,
	}
}

func (s *adminService) GetCookingOrders(ctx context.Context, shopID int) ([]models.AdminOrderResponse, error) {

	cookingOrders, err := s.orr.FindShopOrdersByStatuses(ctx, s.db, shopID, []models.OrderStatus{models.Cooking})
	if err != nil {
		return nil, err
	}
	return s.assembleAdminOrderResponses(ctx, cookingOrders)
}

func (s *adminService) GetCompletedOrders(ctx context.Context, shopID int) ([]models.AdminOrderResponse, error) {
	completedOrders, err := s.orr.FindShopOrdersByStatuses(ctx, s.db, shopID, []models.OrderStatus{models.Completed})
	if err != nil {
		return nil, err
	}
	return s.assembleAdminOrderResponses(ctx, completedOrders)
}

func (s *adminService) assembleAdminOrderResponses(ctx context.Context, dbOrders []repositories.AdminOrderDBResult) ([]models.AdminOrderResponse, error) {
	if len(dbOrders) == 0 {
		return []models.AdminOrderResponse{}, nil
	}

	orderIDs := make([]int, len(dbOrders))
	for i, o := range dbOrders {
		orderIDs[i] = o.OrderID
	}
	itemsMap, err := s.orr.FindItemsByOrderIDs(ctx, s.db, orderIDs)
	if err != nil {
		return nil, err
	}

	responses := make([]models.AdminOrderResponse, len(dbOrders))
	for i, dbOrder := range dbOrders {
		var emailPtr *string
		if dbOrder.CustomerEmail.Valid {
			emailPtr = &dbOrder.CustomerEmail.String
		}

		responses[i] = models.AdminOrderResponse{
			OrderID:       dbOrder.OrderID,
			CustomerEmail: emailPtr,
			OrderDate:     dbOrder.OrderDate,
			TotalAmount:   dbOrder.TotalAmount,
			Status:        dbOrder.Status.String(),
			Items:         itemsMap[dbOrder.OrderID],
		}
	}
	return responses, nil
}

func (s *adminService) UpdateOrderStatus(ctx context.Context, adminShopID int, targetOrderID int) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return apperrors.Unknown.Wrap(err, "トランザクションの開始に失敗しました。")
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil {
				err = apperrors.Unknown.Wrap(err, "トランザクションのコミットに失敗しました。")
			}
		}
	}()

	// 注文の確認とステータス更新を同一トランザクション内で実行
	currentOrder, err := s.orr.FindOrderByIDAndShopID(ctx, tx, targetOrderID, adminShopID)
	if err != nil {
		return err
	}

	var nextStatus models.OrderStatus
	switch currentOrder.Status {
	case models.Cooking:
		nextStatus = models.Completed
	case models.Completed:
		nextStatus = models.Handed
	default:
		return apperrors.Conflict.Wrapf(nil, "ステータスが'%s'の注文はこれ以上進められません。", currentOrder.Status.String())
	}

	return s.orr.UpdateOrderStatus(ctx, tx, targetOrderID, adminShopID, nextStatus)
}

func (s *adminService) DeleteOrder(ctx context.Context, adminShopID int, targetOrderID int) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return apperrors.Unknown.Wrap(err, "トランザクションの開始に失敗しました。")
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil {
				err = apperrors.Unknown.Wrap(err, "トランザクションのコミットに失敗しました。")
			}
		}
	}()

	// 注文の存在確認と削除を同一トランザクション内で実行
	_, err = s.orr.FindOrderByIDAndShopID(ctx, tx, targetOrderID, adminShopID)
	if err != nil {
		return err
	}

	return s.orr.DeleteOrderByIDAndShopID(ctx, tx, targetOrderID, adminShopID)
}

// UpdateItemAvailability は商品の販売状態を更新します
func (s *adminService) UpdateItemAvailability(ctx context.Context, itemID int, isAvailable bool) error {
	return s.itr.UpdateItemAvailability(ctx, s.db, itemID, isAvailable)
}
