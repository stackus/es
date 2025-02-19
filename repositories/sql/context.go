package sql

import (
	"context"
	"database/sql"
)

type contextKey struct{}

var txKey = contextKey{}

type dbContext struct {
	*sql.DB
}

func (c dbContext) withTx(ctx context.Context, fn func(ctx context.Context, tx *sql.Tx) error) (err error) {
	// if there is an existing transaction, use it
	if tx, ok := ctx.Value(txKey).(*sql.Tx); ok {
		return fn(ctx, tx)
	}
	var tx *sql.Tx
	tx, err = c.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// defer func() {
	// 	if p := recover(); p != nil {
	// 		if rbErr := tx.Rollback(); rbErr != nil {
	// 			err = errors.Join(err, rbErr)
	// 		}
	// 		panic(p)
	// 	}
	// }()
	qCtx := context.WithValue(ctx, txKey, tx)
	err = fn(qCtx, tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
