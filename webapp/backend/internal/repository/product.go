package repository

import (
	"backend/internal/model"
	"context"
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

	baseQuery := `
		SELECT product_id, name, value, weight, image, description
		FROM products
	`
	args := []interface{}{}

	if req.Search != "" {
		baseQuery += " WHERE MATCH(name, description) AGAINST(? IN BOOLEAN MODE)"
		args = append(args, req.Search+"*")
	}

	baseQuery += " ORDER BY " + req.SortField + " " + req.SortOrder + ", product_id ASC LIMIT ? OFFSET ?"
	args = append(args, req.PageSize, req.Offset)

	countQuery := "SELECT COUNT(*) FROM products"
	countArgs := []interface{}{}
	if req.Search != "" {
		countQuery += " WHERE MATCH(name, description) AGAINST(? IN BOOLEAN MODE)"
		countArgs = append(countArgs, req.Search+"*")
	}

	if err := r.db.SelectContext(ctx, &products, baseQuery, args...); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
		return nil, 0, err
	}

	return products, total, nil
}
