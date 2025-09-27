package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service/utils"
	"context"
	"log"
)

type RobotService struct {
	store *repository.Store
}

func NewRobotService(store *repository.Store) *RobotService {
	return &RobotService{store: store}
}

func (s *RobotService) GenerateDeliveryPlan(ctx context.Context, robotID string, capacity int) (*model.DeliveryPlan, error) {
	var plan model.DeliveryPlan

	err := utils.WithTimeout(ctx, func(ctx context.Context) error {
		return s.store.ExecTx(ctx, func(txStore *repository.Store) error {
			orders, err := txStore.OrderRepo.GetShippingOrders(ctx)
			if err != nil {
				return err
			}
			plan, err = selectOrdersForDelivery(ctx, orders, robotID, capacity)
			if err != nil {
				return err
			}

			if len(plan.Orders) > 0 {
				orderIDs := make([]int64, len(plan.Orders))
				for i, order := range plan.Orders {
					orderIDs[i] = order.OrderID
				}

				if err := txStore.OrderRepo.UpdateStatuses(ctx, orderIDs, "delivering"); err != nil {
					return err
				}
				log.Printf("Updated status to 'delivering' for %d orders", len(orderIDs))
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (s *RobotService) UpdateOrderStatus(ctx context.Context, orderID int64, newStatus string) error {
	return utils.WithTimeout(ctx, func(ctx context.Context) error {
		return s.store.OrderRepo.UpdateStatuses(ctx, []int64{orderID}, newStatus)
	})
}

func selectOrdersForDelivery(_ context.Context, orders []model.Order, robotID string, robotCapacity int) (model.DeliveryPlan, error) {
	n := len(orders)
	log.Printf("selectOrdersForDelivery: robotID=%s, robotCapacity=%d, orders count=%d", robotID, robotCapacity, n)
	if n == 0 {
		log.Printf("No orders available for delivery")
		return model.DeliveryPlan{
			RobotID:     robotID,
			TotalWeight: 0,
			TotalValue:  0,
			Orders:      []model.Order{},
		}, nil
	}

	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, robotCapacity+1)
	}

	for i := 1; i <= n; i++ {
		order := orders[i-1]
		for w := 0; w <= robotCapacity; w++ {
			dp[i][w] = dp[i-1][w]
			if order.Weight <= w {
				if dp[i-1][w-order.Weight]+order.Value > dp[i][w] {
					dp[i][w] = dp[i-1][w-order.Weight] + order.Value
				}
			}
		}
	}

	bestValue := dp[n][robotCapacity]
	var bestSet []model.Order
	w := robotCapacity
	for i := n; i > 0; i-- {
		order := orders[i-1]
		if w >= order.Weight && dp[i][w] == dp[i-1][w-order.Weight]+order.Value {
			bestSet = append(bestSet, order)
			w -= order.Weight
		}
	}

	for i, j := 0, len(bestSet)-1; i < j; i, j = i+1, j-1 {
		bestSet[i], bestSet[j] = bestSet[j], bestSet[i]
	}

	var totalWeight int
	for _, o := range bestSet {
		totalWeight += o.Weight
	}

	log.Printf("Selected %d orders, totalWeight=%d, totalValue=%d", len(bestSet), totalWeight, bestValue)
	return model.DeliveryPlan{
		RobotID:     robotID,
		TotalWeight: totalWeight,
		TotalValue:  bestValue,
		Orders:      bestSet,
	}, nil
}
