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

func loadProductCount(ctx context.Context, db DBTX, query string) (int, error) {
	return fetchProductCount(ctx, db, query, false)
}

func refreshProductCount(ctx context.Context, db DBTX, query string) (int, error) {
	return fetchProductCount(ctx, db, query, true)
}

func fetchProductCount(ctx context.Context, db DBTX, query string, force bool) (int, error) {
	if !force {
		if ready := productCountCacheReady.Load(); ready {
			return int(productCountCacheValue.Load()), nil
		}
	}

	productCountCacheMu.Lock()
	defer productCountCacheMu.Unlock()

	if !force {
		if ready := productCountCacheReady.Load(); ready {
			return int(productCountCacheValue.Load()), nil
		}
	}

	var total int
	if err := db.GetContext(ctx, &total, query); err != nil {
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
	if len(countArgs) == 0 {
		var err error
		if req.Offset == 0 {
			total, err = refreshProductCount(ctx, r.db, countQuery)
		} else {
			total, err = loadProductCount(ctx, r.db, countQuery)
			if err == nil && req.Offset+len(products) > total {
				total, err = refreshProductCount(ctx, r.db, countQuery)
			}
		}
		if err != nil {
			return nil, 0, err
		}
	}
	if req.Search != "" {
		if err := r.db.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
			return nil, 0, err
		}
		if req.Offset == 0 && !productCountCacheReady.Load() {
			refreshProductCount(ctx, r.db, "SELECT COUNT(*) FROM products")
		}
	}

	return products, total, nil
}
