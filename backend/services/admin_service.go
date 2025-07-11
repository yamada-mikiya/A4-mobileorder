package services

import (
	"context"
	"fmt"

	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
)

type AdminServicer interface {
	GetCookingOrders(ctx context.Context, shopID int) ([]models.AdminOrderResponse, error)
	GetCompletedOrders(ctx context.Context, shopID int) ([]models.AdminOrderResponse, error)
	UpdateOrderStatus(ctx context.Context, adminShopID int, targetOrderID int) error
	DeleteOrder(ctx context.Context, adminShopID int, targetOrderID int) error
}

type adminService struct {
	orr repositories.OrderRepository
}

func NewAdminService(orr repositories.OrderRepository) AdminServicer {
	return &adminService{orr}
}

func (s *adminService) GetCookingOrders(ctx context.Context, shopID int) ([]models.AdminOrderResponse, error) {

	dbOrders, err := s.orr.FindShopOrdersByStatuses(ctx, shopID, []models.OrderStatus{models.Cooking})
	if err != nil {
		return nil, err
	}
	return s.assembleAdminOrderResponses(ctx, dbOrders)
}

func (s *adminService) GetCompletedOrders(ctx context.Context, shopID int) ([]models.AdminOrderResponse, error) {
	dbOrders, err := s.orr.FindShopOrdersByStatuses(ctx, shopID, []models.OrderStatus{models.Completed})
	if err != nil {
		return nil, err
	}
	return s.assembleAdminOrderResponses(ctx, dbOrders)
}

func (s *adminService) assembleAdminOrderResponses(ctx context.Context, dbOrders []repositories.AdminOrderDBResult) ([]models.AdminOrderResponse, error) {
	if len(dbOrders) == 0 {
		return []models.AdminOrderResponse{}, nil
	}

	orderIDs := make([]int, len(dbOrders))
	for i, o := range dbOrders {
		orderIDs[i] = o.OrderID
	}
	itemsMap, err := s.orr.FindItemsByOrderIDs(ctx, orderIDs)
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

	currentOrder, err := s.orr.FindOrderByIDAndShopID(ctx, targetOrderID, adminShopID)
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
		return fmt.Errorf("order with status '%s' cannot be advanced", currentOrder.Status.String())
	}

	return s.orr.UpdateOrderStatus(ctx, targetOrderID, adminShopID, nextStatus)
}

func (s *adminService) DeleteOrder(ctx context.Context, adminShopID int, targetOrderID int) error {

	_, err := s.orr.FindOrderByIDAndShopID(ctx, targetOrderID, adminShopID)
	if err != nil {
		return err
	}

	return s.orr.DeleteOrderByIDAndShopID(ctx, targetOrderID, adminShopID)
}
