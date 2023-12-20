package models

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type sqlxnDB struct {
	*sqlx.DB
}

func (db sqlxnDB) namedGet(dest any, query string, arg any) error {
	preparedQuery, err := db.PrepareNamed(query)
	if err != nil {
		return err
	}
	err = preparedQuery.Get(dest, arg)
	return err
}

func (db sqlxnDB) namedSelect(dest any, query string, arg any) error {
	preparedQuery, err := db.PrepareNamed(query)
	if err != nil {
		return err
	}
	err = preparedQuery.Select(dest, arg)
	return err
}

func (db sqlxnDB) namedGetContext(ctx context.Context, dest any, query string, arg any) error {
	preparedQuery, err := db.PrepareNamed(query)
	if err != nil {
		return err
	}
	err = preparedQuery.GetContext(ctx, dest, arg)
	return err
}

func (db sqlxnDB) namedSelectContext(ctx context.Context, dest any, query string, arg any) error {
	preparedQuery, err := db.PrepareNamed(query)
	if err != nil {
		return err
	}
	err = preparedQuery.SelectContext(ctx, dest, arg)
	return err
}
