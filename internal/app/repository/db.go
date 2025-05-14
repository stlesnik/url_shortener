package repository

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB interface {
	Ping() error
	Close() error
}

// интерфейс реализуется явно
var _ DB = (*DataBase)(nil)

type DataBase struct {
	db *sql.DB
}

func NewDataBase(dsn string) (*DataBase, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("got error while connecting to db: %w", err)
	}
	return &DataBase{db: db}, nil
}
func (d *DataBase) Ping() error {
	if d.db == nil {
		return fmt.Errorf("database does not exist")
	}
	if err := d.db.PingContext(context.TODO()); err != nil {
		return fmt.Errorf("error while ping to db: %w", err)
	}
	return nil
}

func (d *DataBase) Close() error {
	return d.db.Close()
}
