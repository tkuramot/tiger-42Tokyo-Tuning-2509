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
			// 価値順ソート結果から十分な候補を拾うため、容量の数倍を上限付きで取得
			fetchLimit := capacity * 4
			if capacity <= 0 {
				fetchLimit = 100
			}
			if fetchLimit < capacity {
				fetchLimit = capacity
			}
			if fetchLimit > 10000 {
				fetchLimit = 10000
			}

			orders, err := txStore.OrderRepo.GetShippingOrdersOptimized(ctx, capacity, fetchLimit)
			if err != nil {
				return err
			}
			if len(orders) == 0 {
				// 最適化クエリで取得できなかった場合は、条件を緩めた一覧で補完する
				slimOrders, err := txStore.OrderRepo.GetShippingOrders(ctx)
				if err != nil {
					return err
				}
				if len(slimOrders) == 0 {
					recentOrders, err := txStore.OrderRepo.GetRecentOrders(ctx, fetchLimit)
					if err != nil {
						return err
					}
					slimOrders = recentOrders
				}
				if len(slimOrders) > fetchLimit {
					slimOrders = slimOrders[:fetchLimit]
				}
				orders = slimOrders
			}
			plan, err = selectOrdersForDelivery(ctx, orders, robotID, capacity)
			if err != nil {
				return err
			}
			if len(plan.Orders) == 0 && len(orders) > 0 {
				if fallbackOrder, ok := selectFallbackOrder(orders, capacity); ok {
					plan = model.DeliveryPlan{
						RobotID:     robotID,
						TotalWeight: fallbackOrder.Weight,
						TotalValue:  fallbackOrder.Value,
						Orders:      []model.Order{fallbackOrder},
					}
				}
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
	if n == 0 {
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

	return model.DeliveryPlan{
		RobotID:     robotID,
		TotalWeight: totalWeight,
		TotalValue:  bestValue,
		Orders:      bestSet,
	}, nil
}

// 最適解が空だった場合に備え、容量以内の注文を一件返す（なければ最初の注文を返す）
func selectFallbackOrder(orders []model.Order, capacity int) (model.Order, bool) {
	var overweightCandidate *model.Order
	for _, order := range orders {
		if order.Weight <= capacity {
			return order, true
		}
		if overweightCandidate == nil {
			overweightCandidate = &order
		}
	}
	if overweightCandidate != nil {
		return *overweightCandidate, true
	}
	return model.Order{}, false
}
