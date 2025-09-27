package service

import (
	"context"
	"log"

	"backend/internal/model"
	"backend/internal/repository"
)

type ProductService struct {
	store *repository.Store
}

func NewProductService(store *repository.Store) *ProductService {
	return &ProductService{store: store}
}

func (s *ProductService) CreateOrders(ctx context.Context, userID int, items []model.RequestItem) ([]string, error) {
	var insertedOrderIDs []string

		err := s.store.ExecTx(ctx, func(txStore *repository.Store) error {
		itemsToProcess := make(map[int]int)
		for _, item := range items {
			if item.Quantity > 0 {
				itemsToProcess[item.ProductID] += item.Quantity
			}
		}
		if len(itemsToProcess) == 0 {
			return nil
		}

		totalQuantity := 0 // 先に容量を見積もってスライス再確保を避ける
		for _, quantity := range itemsToProcess {
			totalQuantity += quantity
		}

		ordersToInsert := make([]model.Order, 0, totalQuantity)
		for productID, quantity := range itemsToProcess {
			for i := 0; i < quantity; i++ {
				ordersToInsert = append(ordersToInsert, model.Order{
					UserID:    userID,
					ProductID: productID,
				})
			}
		}

		orderIDs, err := txStore.OrderRepo.CreateBulk(ctx, ordersToInsert)
		if err != nil {
			return err
		}
		insertedOrderIDs = append(insertedOrderIDs, orderIDs...)
		return nil
	})

	if err != nil {
		return nil, err
	}
	log.Printf("Created %d orders for user %d", len(insertedOrderIDs), userID)
	return insertedOrderIDs, nil
}

func (s *ProductService) FetchProducts(ctx context.Context, userID int, req model.ListRequest) ([]model.Product, int, error) {
	products, total, err := s.store.ProductRepo.ListProducts(ctx, userID, req)
	return products, total, err
}
