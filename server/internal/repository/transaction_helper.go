package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Tx interface {
	Commit() error
	Rollback() error
}

type TxWrapper struct {
	tx *sqlx.Tx
}

func (t *TxWrapper) Commit() error   { return t.tx.Commit() }
func (t *TxWrapper) Rollback() error { return t.tx.Rollback() }

type TransactionManager interface {
	BeginTx(ctx context.Context) (Tx, error)
}

type transactionManager struct {
	db *sqlx.DB
}

func NewTransactionManager(db *sqlx.DB) TransactionManager {
	return &transactionManager{db: db}
}

func (tm *transactionManager) BeginTx(ctx context.Context) (Tx, error) {
	tx, err := tm.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &TxWrapper{tx: tx}, nil
}

func UnwrapTx(tx Tx) *sqlx.Tx {
	if wrapper, ok := tx.(*TxWrapper); ok {
		return wrapper.tx
	}
	return nil
}
