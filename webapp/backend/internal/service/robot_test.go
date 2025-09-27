package service

import (
	"backend/internal/model"
	"context"
	"testing"
)

func TestSelectOrdersForDelivery(t *testing.T) {
	tests := []struct {
		name           string
		orders         []model.Order
		robotID        string
		robotCapacity  int
		wantTotalValue int
		wantOrderCount int
	}{
		{
			name:           "empty orders",
			orders:         []model.Order{},
			robotID:        "robot1",
			robotCapacity:  100,
			wantTotalValue: 0,
			wantOrderCount: 0,
		},
		{
			name: "single order fits",
			orders: []model.Order{
				{OrderID: 1, Weight: 50, Value: 100},
			},
			robotID:        "robot1",
			robotCapacity:  100,
			wantTotalValue: 100,
			wantOrderCount: 1,
		},
		{
			name: "single order too heavy",
			orders: []model.Order{
				{OrderID: 1, Weight: 150, Value: 100},
			},
			robotID:        "robot1",
			robotCapacity:  100,
			wantTotalValue: 0,
			wantOrderCount: 0,
		},
		{
			name: "multiple orders - select best value",
			orders: []model.Order{
				{OrderID: 1, Weight: 30, Value: 50},
				{OrderID: 2, Weight: 40, Value: 80},
				{OrderID: 3, Weight: 50, Value: 90},
			},
			robotID:        "robot1",
			robotCapacity:  100,
			wantTotalValue: 170,
			wantOrderCount: 2,
		},
		{
			name: "capacity constraint - choose optimal subset",
			orders: []model.Order{
				{OrderID: 1, Weight: 10, Value: 60},
				{OrderID: 2, Weight: 20, Value: 100},
				{OrderID: 3, Weight: 30, Value: 120},
			},
			robotID:        "robot1",
			robotCapacity:  50,
			wantTotalValue: 220,
			wantOrderCount: 2,
		},
		{
			name: "exact capacity",
			orders: []model.Order{
				{OrderID: 1, Weight: 25, Value: 50},
				{OrderID: 2, Weight: 25, Value: 50},
				{OrderID: 3, Weight: 25, Value: 50},
				{OrderID: 4, Weight: 25, Value: 50},
			},
			robotID:        "robot1",
			robotCapacity:  100,
			wantTotalValue: 200,
			wantOrderCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			plan, err := selectOrdersForDelivery(ctx, tt.orders, tt.robotID, tt.robotCapacity)
			if err != nil {
				t.Fatalf("selectOrdersForDelivery() error = %v", err)
			}

			if plan.TotalValue != tt.wantTotalValue {
				t.Errorf("TotalValue = %v, want %v", plan.TotalValue, tt.wantTotalValue)
			}

			if len(plan.Orders) != tt.wantOrderCount {
				t.Errorf("OrderCount = %v, want %v", len(plan.Orders), tt.wantOrderCount)
			}

			if plan.RobotID != tt.robotID {
				t.Errorf("RobotID = %v, want %v", plan.RobotID, tt.robotID)
			}

			if plan.TotalWeight > tt.robotCapacity {
				t.Errorf("TotalWeight %v exceeds capacity %v", plan.TotalWeight, tt.robotCapacity)
			}

			actualValue := 0
			actualWeight := 0
			for _, order := range plan.Orders {
				actualValue += order.Value
				actualWeight += order.Weight
			}

			if actualValue != plan.TotalValue {
				t.Errorf("Actual value sum %v != TotalValue %v", actualValue, plan.TotalValue)
			}

			if actualWeight != plan.TotalWeight {
				t.Errorf("Actual weight sum %v != TotalWeight %v", actualWeight, plan.TotalWeight)
			}
		})
	}
}