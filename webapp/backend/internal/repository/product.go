package repository

import (
	"backend/internal/model"
	"context"
	"strings"
	"sync"
	"sync/atomic"
)

// DB へのアクセスをまとめて面倒を見る層。UseCase からはこのパッケージを経由して DB とやり取りする。
type ProductRepository struct {
	db DBTX
}

var (
	productCountCacheValue atomic.Int64
	productCountCacheReady atomic.Bool
	productCountCacheMu    sync.Mutex
)

func getProductCount(ctx context.Context, db DBTX) (int, error) {
	if productCountCacheReady.Load() {
		return int(productCountCacheValue.Load()), nil
	}

	productCountCacheMu.Lock()
	defer productCountCacheMu.Unlock()

	if productCountCacheReady.Load() {
		return int(productCountCacheValue.Load()), nil
	}

	var total int
	if err := db.GetContext(ctx, &total, "SELECT COUNT(*) FROM products"); err != nil {
		return 0, err
	}
	productCountCacheValue.Store(int64(total))
	productCountCacheReady.Store(true)

	return total, nil
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
		searchTerms := strings.Fields(req.Search)
		for i, term := range searchTerms {
			searchTerms[i] = term + "*"
		}
		args = append(args, strings.Join(searchTerms, " "))
	}

	baseQuery += " ORDER BY " + req.SortField + " " + req.SortOrder + ", product_id ASC LIMIT ? OFFSET ?"
	args = append(args, req.PageSize, req.Offset)

	countQuery := "SELECT COUNT(*) FROM products"
	countArgs := []interface{}{}
	if req.Search != "" {
		countQuery += " WHERE MATCH(name, description) AGAINST(? IN BOOLEAN MODE)"
		searchTerms := strings.Fields(req.Search)
		for i, term := range searchTerms {
			searchTerms[i] = term + "*"
		}
		countArgs = append(countArgs, strings.Join(searchTerms, " "))
	}

	if err := r.db.SelectContext(ctx, &products, baseQuery, args...); err != nil {
		return nil, 0, err
	}

	var total int
	if req.Search == "" {
		cachedTotal, err := getProductCount(ctx, r.db)
		if err != nil {
			return nil, 0, err
		}
		total = cachedTotal
	} else {
		if err := r.db.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
			return nil, 0, err
		}
	}

	return products, total, nil
}
