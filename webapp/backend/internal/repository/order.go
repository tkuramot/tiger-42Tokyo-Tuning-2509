package repository

import (
	"backend/internal/model"
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type OrderRepository struct {
	db DBTX
}

func NewOrderRepository(db DBTX) *OrderRepository {
	return &OrderRepository{db: db}
}

// 注文を作成し、生成された注文IDを返す
func (r *OrderRepository) Create(ctx context.Context, order *model.Order) (string, error) {
	query := `INSERT INTO orders (user_id, product_id, shipped_status, created_at) VALUES (?, ?, 'shipping', NOW())`
	result, err := r.db.ExecContext(ctx, query, order.UserID, order.ProductID)
	if err != nil {
		return "", err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", id), nil
}

// 複数の注文IDのステータスを一括で更新
// 主に配送ロボットが注文を引き受けた際に一括更新をするために使用
func (r *OrderRepository) UpdateStatuses(ctx context.Context, orderIDs []int64, newStatus string) error {
	if len(orderIDs) == 0 {
		return nil
	}
	query, args, err := sqlx.In("UPDATE orders SET shipped_status = ? WHERE order_id IN (?)", newStatus, orderIDs)
	if err != nil {
		return err
	}
	query = r.db.Rebind(query)
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

// 配送中(shipped_status:shipping)の注文一覧を取得
func (r *OrderRepository) GetShippingOrders(ctx context.Context) ([]model.Order, error) {
	var orders []model.Order
	query := `
        SELECT
            o.order_id,
            p.weight,
            p.value
        FROM orders o
        JOIN products p ON o.product_id = p.product_id
        WHERE o.shipped_status = 'shipping'
    `
	err := r.db.SelectContext(ctx, &orders, query)
	return orders, err
}

// 配送中(shipped_status:shipping)の注文一覧を効率的に取得
// maxWeight: ロボットの積載容量
// limit: 取得する最大件数
func (r *OrderRepository) GetShippingOrdersOptimized(ctx context.Context, maxWeight int, limit int) ([]model.Order, error) {
	var orders []model.Order
	query := `
        SELECT
            o.order_id,
            p.weight,
            p.value
        FROM orders o
        JOIN products p ON o.product_id = p.product_id
        WHERE o.shipped_status = 'shipping' AND p.weight <= ?
        ORDER BY p.value DESC, o.created_at ASC
        LIMIT ?
    `
	err := r.db.SelectContext(ctx, &orders, query, maxWeight, limit)
	return orders, err
}

// 注文履歴一覧を取得
func (r *OrderRepository) ListOrders(ctx context.Context, userID int, req model.ListRequest) ([]model.Order, int, error) {
	var whereConditions []string
	var args []interface{}

	whereConditions = append(whereConditions, "o.user_id = ?")
	args = append(args, userID)

	if req.Search != "" {
		if req.Type == "prefix" {
			whereConditions = append(whereConditions, "p.name LIKE ?")
			args = append(args, req.Search+"%")
		} else {
			whereConditions = append(whereConditions, "p.name LIKE ?")
			args = append(args, "%"+req.Search+"%")
		}
	}

	whereClause := strings.Join(whereConditions, " AND ")

	countQuery := fmt.Sprintf(`
        SELECT COUNT(*)
        FROM orders o
        JOIN products p ON o.product_id = p.product_id
        WHERE %s
    `, whereClause)

	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	var orderByClause string
	sortOrder := "ASC"
	if strings.ToUpper(req.SortOrder) == "DESC" {
		sortOrder = "DESC"
	}

	switch req.SortField {
	case "product_name":
		orderByClause = fmt.Sprintf("p.name %s", sortOrder)
	case "created_at":
		orderByClause = fmt.Sprintf("o.created_at %s", sortOrder)
	case "shipped_status":
		orderByClause = fmt.Sprintf("o.shipped_status %s", sortOrder)
	case "arrived_at":
		orderByClause = fmt.Sprintf("o.arrived_at %s", sortOrder)
	case "order_id":
		fallthrough
	default:
		orderByClause = fmt.Sprintf("o.order_id %s", sortOrder)
	}

	query := fmt.Sprintf(`
        SELECT o.order_id, o.product_id, p.name as product_name, o.shipped_status, o.created_at, o.arrived_at
        FROM orders o
        JOIN products p ON o.product_id = p.product_id
        WHERE %s
        ORDER BY %s
        LIMIT ? OFFSET ?
    `, whereClause, orderByClause)

	args = append(args, req.PageSize, req.Offset)

	type orderRow struct {
		OrderID       int          `db:"order_id"`
		ProductID     int          `db:"product_id"`
		ProductName   string       `db:"product_name"`
		ShippedStatus string       `db:"shipped_status"`
		CreatedAt     sql.NullTime `db:"created_at"`
		ArrivedAt     sql.NullTime `db:"arrived_at"`
	}
	var ordersRaw []orderRow
	if err := r.db.SelectContext(ctx, &ordersRaw, query, args...); err != nil {
		return nil, 0, err
	}

	var orders []model.Order
	for _, o := range ordersRaw {
		orders = append(orders, model.Order{
			OrderID:       int64(o.OrderID),
			ProductID:     o.ProductID,
			ProductName:   o.ProductName,
			ShippedStatus: o.ShippedStatus,
			CreatedAt:     o.CreatedAt.Time,
			ArrivedAt:     o.ArrivedAt,
		})
	}

	return orders, total, nil
}
