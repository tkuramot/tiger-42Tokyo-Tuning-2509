package repository

import (
	"backend/internal/model"
	"context"
	"fmt"
	"strings"
)

// DB へのアクセスをまとめて面倒を見る層。UseCase からはこのパッケージを経由して DB とやり取りする。
type ProductRepository struct {
	db DBTX
}

// NewProductRepository はリポジトリを初期化し、呼び出し側から渡された DB インターフェースを保持する。
func NewProductRepository(db DBTX) *ProductRepository {
	return &ProductRepository{db: db}
}

// 条件やページ番号を受け取り、商品一覧と件数を返す
func (r *ProductRepository) ListProducts(ctx context.Context, userID int, req model.ListRequest) ([]model.Product, int, error) {
	var products []model.Product

	whereClause, whereArgs := buildProductSearchClause(req.Search)

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	sortableColumns := map[string]string{
		"product_id": "product_id",
		"name":       "name",
		"value":      "value",
		"weight":     "weight",
	}
	orderColumn, ok := sortableColumns[req.SortField]
	if !ok {
		orderColumn = "product_id"
	}
	sortOrder := "ASC"
	if strings.EqualFold(req.SortOrder, "DESC") {
		sortOrder = "DESC"
	}

	query := "SELECT product_id, name, value, weight, image, description FROM products"
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	query += fmt.Sprintf(" ORDER BY %s %s, product_id ASC LIMIT ? OFFSET ?", orderColumn, sortOrder)

	args := append([]any{}, whereArgs...)
	args = append(args, pageSize, offset)

	if err := r.db.SelectContext(ctx, &products, query, args...); err != nil {
		return nil, 0, err
	}

	countQuery := "SELECT COUNT(*) FROM products"
	if whereClause != "" {
		countQuery += " WHERE " + whereClause
	}

	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, whereArgs...); err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func buildProductSearchClause(search string) (string, []any) {
	search = strings.TrimSpace(search)
	if search == "" {
		return "", nil
	}

	terms := strings.Fields(search)
	if len(terms) == 0 {
		return "", nil
	}
	for i, term := range terms {
		terms[i] = term + "*"
	}

	return "MATCH(name, description) AGAINST(? IN BOOLEAN MODE)", []any{strings.Join(terms, " ")}
}
