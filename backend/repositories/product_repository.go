package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/jmoiron/sqlx"
)

type ProductRepository interface {
	ValidateAndGetProductsForShop(ctx context.Context, shopID int, productIDs []int) (map[int]models.Product, error)
	GetProductList(shopID int) ([]models.ProductListResponse, error)
}

type productRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) ProductRepository {
	return &productRepository{db}
}

func (r *productRepository) ValidateAndGetProductsForShop(ctx context.Context, shopID int, productIDs []int) (map[int]models.Product, error) {
	//商品IDで商品情報取得
	productMap := make(map[int]models.Product)

	if len(productIDs) == 0 {
		return productMap, nil
	}

	const baseQuery = `
		SELECT
			p.product_id,
			p.product_name,
			p.price,
			p.is_available
		FROM
			products p
		INNER JOIN
			shop_product sp ON p.product_id = sp.product_id
		WHERE
			sp.shop_id = ? AND p.product_id IN (?)
	`

	query, args, err := sqlx.In(baseQuery, shopID, productIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to build query for product validation: %w", err)
	}
	query = r.db.Rebind(query)

	var products []models.Product
	if err := r.db.SelectContext(ctx, &products, query, args...); err != nil {
		return nil, fmt.Errorf("failed to select products for shop: %w", err)
	}

	if len(products) != len(productIDs) {
		return nil, errors.New("one or more products do not belong to the specified shop")
	}

	for _, p := range products {
		productMap[p.ProductID] = p
	}

	return productMap, nil

}

func (r *productRepository) GetProductList(shopID int) ([]models.ProductListResponse, error) {
	//TODO
	return nil, nil
}