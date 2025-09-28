package repository

import (
    "backend/internal/model"
    "context"
    "fmt"
    "sort"
    "strings"
    "sync"
    "unicode"
)

// ProductRepository mediates DB access for product data.
type ProductRepository struct {
    db              DBTX
    cacheMu         sync.RWMutex
    cachedProducts  []model.Product
    cachedHaystacks []string
}

// NewProductRepository initialises repository with provided DB handle.
func NewProductRepository(db DBTX) *ProductRepository {
    return &ProductRepository{db: db}
}

// ListProducts returns products and total count according to request params.
func (r *ProductRepository) ListProducts(ctx context.Context, userID int, req model.ListRequest) ([]model.Product, int, error) {
    whereClause, whereArgs, terms := buildProductSearchClause(req.Search)
    orderColumn, sortOrder := resolveSort(req.SortField, req.SortOrder)

    hasNonASCII := containsNonASCII(req.Search)

    var (
        fallbackTried   bool
        fallbackProducts []model.Product
        fallbackTotal    int
    )

    if len(terms) > 0 && hasNonASCII {
        var fallbackErr error
        fallbackProducts, fallbackTotal, fallbackErr = r.fallbackSearchProducts(ctx, req, orderColumn, sortOrder, terms)
        if fallbackErr != nil {
            return nil, 0, fallbackErr
        }
        fallbackTried = true
        if fallbackTotal > 0 {
            return fallbackProducts, fallbackTotal, nil
        }
    }

    products, total, err := r.fetchProducts(ctx, req, orderColumn, sortOrder, whereClause, whereArgs)
    if err != nil {
        return nil, 0, err
    }

    if len(terms) > 0 && hasNonASCII && total == 0 {
        if !fallbackTried {
            var fallbackErr error
            fallbackProducts, fallbackTotal, fallbackErr = r.fallbackSearchProducts(ctx, req, orderColumn, sortOrder, terms)
            if fallbackErr != nil {
                return nil, 0, fallbackErr
            }
        }
        if fallbackTotal > 0 {
            return fallbackProducts, fallbackTotal, nil
        }
    }

    return products, total, nil
}

func buildProductSearchClause(search string) (string, []any, []string) {
    trimmed := strings.TrimSpace(search)
    if trimmed == "" {
        return "", nil, nil
    }

    terms := strings.Fields(trimmed)
    if len(terms) == 0 {
        terms = []string{trimmed}
    }

    fulltextTerms := make([]string, len(terms))
    for i, term := range terms {
        fulltextTerms[i] = term + "*"
    }

    return "MATCH(name, description) AGAINST(? IN BOOLEAN MODE)", []any{strings.Join(fulltextTerms, " ")}, terms
}

func (r *ProductRepository) fetchProducts(
    ctx context.Context,
    req model.ListRequest,
    orderColumn, sortOrder, whereClause string,
    whereArgs []any,
) ([]model.Product, int, error) {
    pageSize := req.PageSize
    if pageSize <= 0 {
        pageSize = 20
    }
    offset := req.Offset
    if offset < 0 {
        offset = 0
    }

    query := "SELECT product_id, name, value, weight, image, description FROM products"
    if whereClause != "" {
        query += " WHERE " + whereClause
    }
    query += fmt.Sprintf(" ORDER BY %s %s, product_id ASC LIMIT ? OFFSET ?", orderColumn, sortOrder)

    args := append([]any{}, whereArgs...)
    args = append(args, pageSize, offset)

    var products []model.Product
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

func (r *ProductRepository) fallbackSearchProducts(
    ctx context.Context,
    req model.ListRequest,
    orderColumn, sortOrder string,
    terms []string,
) ([]model.Product, int, error) {
    products, haystacks, err := r.ensureCache(ctx)
    if err != nil {
        return nil, 0, err
    }

    lowerTerms := make([]string, 0, len(terms))
    for _, term := range terms {
        trimmed := strings.TrimSpace(term)
        if trimmed == "" {
            continue
        }
        lowerTerms = append(lowerTerms, strings.ToLower(trimmed))
    }
    if len(lowerTerms) == 0 {
        return []model.Product{}, 0, nil
    }

    filtered := make([]model.Product, 0)
    for i, product := range products {
        hay := haystacks[i]
        matched := true
        for _, term := range lowerTerms {
            if !strings.Contains(hay, term) {
                matched = false
                break
            }
        }
        if matched {
            filtered = append(filtered, product)
        }
    }

    if len(filtered) == 0 {
        return []model.Product{}, 0, nil
    }

    sortProducts(filtered, orderColumn, sortOrder)

    total := len(filtered)
    pageSize := req.PageSize
    if pageSize <= 0 {
        pageSize = 20
    }
    offset := req.Offset
    if offset < 0 {
        offset = 0
    }
    if offset >= total {
        return []model.Product{}, total, nil
    }

    end := offset + pageSize
    if end > total {
        end = total
    }

    result := make([]model.Product, end-offset)
    copy(result, filtered[offset:end])

    return result, total, nil
}

func resolveSort(sortField, sortOrder string) (string, string) {
    sortableColumns := map[string]string{
        "product_id": "product_id",
        "name":       "name",
        "value":      "value",
        "weight":     "weight",
    }
    orderColumn, ok := sortableColumns[sortField]
    if !ok {
        orderColumn = "product_id"
    }
    order := "ASC"
    if strings.EqualFold(sortOrder, "DESC") {
        order = "DESC"
    }
    return orderColumn, order
}

func (r *ProductRepository) ensureCache(ctx context.Context) ([]model.Product, []string, error) {
    r.cacheMu.RLock()
    if r.cachedProducts != nil && r.cachedHaystacks != nil {
        products := r.cachedProducts
        haystacks := r.cachedHaystacks
        r.cacheMu.RUnlock()
        return products, haystacks, nil
    }
    r.cacheMu.RUnlock()

    r.cacheMu.Lock()
    defer r.cacheMu.Unlock()
    if r.cachedProducts != nil && r.cachedHaystacks != nil {
        return r.cachedProducts, r.cachedHaystacks, nil
    }

    var products []model.Product
    const query = "SELECT product_id, name, value, weight, image, description FROM products ORDER BY product_id ASC"
    if err := r.db.SelectContext(ctx, &products, query); err != nil {
        return nil, nil, err
    }

    haystacks := make([]string, len(products))
    for i, product := range products {
        haystacks[i] = strings.ToLower(product.Name + " " + product.Description)
    }

    r.cachedProducts = products
    r.cachedHaystacks = haystacks

    return r.cachedProducts, r.cachedHaystacks, nil
}

func sortProducts(products []model.Product, column, order string) {
    if len(products) <= 1 {
        return
    }

    sort.SliceStable(products, func(i, j int) bool {
        less, equal := compareProductField(products[i], products[j], column)
        if equal {
            return products[i].ProductID < products[j].ProductID
        }
        if strings.EqualFold(order, "DESC") {
            return !less
        }
        return less
    })
}

func compareProductField(a, b model.Product, column string) (bool, bool) {
    switch column {
    case "name":
        if a.Name == b.Name {
            return false, true
        }
        return a.Name < b.Name, false
    case "value":
        if a.Value == b.Value {
            return false, true
        }
        return a.Value < b.Value, false
    case "weight":
        if a.Weight == b.Weight {
            return false, true
        }
        return a.Weight < b.Weight, false
    case "product_id":
        if a.ProductID == b.ProductID {
            return false, true
        }
        return a.ProductID < b.ProductID, false
    default:
        if a.ProductID == b.ProductID {
            return false, true
        }
        return a.ProductID < b.ProductID, false
    }
}

func containsNonASCII(s string) bool {
    for _, r := range s {
        if r > unicode.MaxASCII {
            return true
        }
    }
    return false
}
