package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Store struct {
	db          DBTX
	UserRepo    *UserRepository
	SessionRepo *SessionRepository
	ProductRepo *ProductRepository
	OrderRepo   *OrderRepository
}

func NewStore(db DBTX) *Store {
	return &Store{
		db:          db,
		UserRepo:    NewUserRepository(db),
		SessionRepo: NewSessionRepository(db),
		ProductRepo: NewProductRepository(db),
		OrderRepo:   NewOrderRepository(db),
	}
}

func (s *Store) ExecTx(ctx context.Context, fn func(txStore *Store) error) error {
	db, ok := s.db.(*sqlx.DB)
	if !ok {
		return fn(s)
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txStore := NewStore(tx)
	if err := fn(txStore); err != nil {
		return err
	}

	return tx.Commit()
}
