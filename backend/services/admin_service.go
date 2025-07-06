package services

import (
	"context"
	"fmt"

	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
)

type AdminServicer interface {
	GetAdminOrderPageData(ctx context.Context, shopID int) (*models.AdminOrderPageResponse, error)
	AdvanceOrderStatus(ctx context.Context, adminShopID int, targetOrderID int) error
	DeleteOrder(ctx context.Context, adminShopID int, targetOrderID int) error
}

type adminService struct {
	orr repositories.OrderRepository
}

func NewAdminService(orr repositories.OrderRepository) AdminServicer {
	return &adminService{orr}
}

func (s *adminService) GetAdminOrderPageData(ctx context.Context, shopID int) (*models.AdminOrderPageResponse, error) {

	dbOrders, err := s.orr.FindActiveOrderByShopID(ctx, shopID)
	if err != nil {
		return nil, err
	}
	if len(dbOrders) == 0 {
		return &models.AdminOrderPageResponse{
			CookingOrders:   []models.AdminOrderResponse{},
			CompletedOrders: []models.AdminOrderResponse{},
		}, nil
	}

	orderIDs := make([]int, len(dbOrders))
	for i, o := range dbOrders {
		orderIDs[i] = o.OrderID
	}
	productsMap, err := s.orr.FindProductsByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, err
	}

	pageData := &models.AdminOrderPageResponse{
		CookingOrders:   make([]models.AdminOrderResponse, 0),
		CompletedOrders: make([]models.AdminOrderResponse, 0),
	}

	for _, dbOrder := range dbOrders {
		var emailPtr *string
		if dbOrder.CustomerEmail.Valid {
			emailPtr = &dbOrder.CustomerEmail.String
		}

		orderResponse := models.AdminOrderResponse{
			OrderID:       dbOrder.OrderID,
			CustomerEmail: emailPtr,
			OrderDate:     dbOrder.OrderDate,
			TotalAmount:   dbOrder.TotalAmount,
			Status:        dbOrder.Status.String(),
			Items:         productsMap[dbOrder.OrderID],
		}

		switch dbOrder.Status {
		case models.Cooking:
			pageData.CookingOrders = append(pageData.CookingOrders, orderResponse)
		case models.Completed:
			pageData.CompletedOrders = append(pageData.CompletedOrders, orderResponse)
		}
	}

	return pageData, nil
}

func (s *adminService) AdvanceOrderStatus(ctx context.Context, adminShopID int, targetOrderID int) error {

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
