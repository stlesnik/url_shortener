package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/stlesnik/url_shortener/internal/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DataBase struct {
	db *sqlx.DB
}

func NewDataBase(dsn string) (*DataBase, error) {
	db := sqlx.MustOpen("pgx", dsn)
	err := warmupDB(db)
	if err != nil {
		logger.Sugaarz.Errorf("error while warming db up: %w", err)
		return nil, fmt.Errorf("error while warming db up: %w", err)
	}
	return &DataBase{db: db}, nil
}

func warmupDB(db *sqlx.DB) error {
	_ = db.MustExecContext(context.Background(), "CREATE TABLE IF NOT EXISTS url("+
		"id serial primary key,"+
		"short_url varchar not null,"+
		"long_url varchar not null unique)")
	return nil
}

func (d *DataBase) Ping(ctx context.Context) error {
	if d.db == nil {
		return fmt.Errorf("database does not exist")
	}
	if err := d.db.PingContext(ctx); err != nil {
		return fmt.Errorf("error while ping to db: %w", err)
	}
	return nil
}

func (d *DataBase) Save(ctx context.Context, short string, long string) (isDouble bool, err error) {
	_, dbErr := d.db.ExecContext(ctx, "INSERT INTO url (short_url, long_url) VALUES ($1, $2)", short, long)
	if dbErr != nil {
		var pgErr *pgconn.PgError
		if errors.As(dbErr, &pgErr) && pgErr.Code == "23505" {
			logger.Sugaarz.Infow("this short url already exists", "short", short, "long", long)
			return true, nil
		}
		logger.Sugaarz.Errorf("error while saving url: %w", dbErr)
		return false, fmt.Errorf("error while saving url: %w", dbErr)
	}
	return false, nil
}

type URLPair struct {
	URLHash string
	LongURL string
}

func (d *DataBase) SaveBatch(ctx context.Context, batch []URLPair) error {
	tx := d.db.MustBegin()

	for _, pair := range batch {
		_, err := tx.ExecContext(ctx, ""+
			"INSERT INTO url (short_url, long_url) "+
			"VALUES ($1, $2) "+
			"ON CONFLICT (long_url) DO NOTHING", pair.URLHash, pair.LongURL)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("error while creating SQL statement in transaction: %w", err)
		}
	}

	return tx.Commit()
}

func (d *DataBase) Get(ctx context.Context, short string) (string, error) {
	var longURL string
	err := d.db.GetContext(ctx, &longURL, "SELECT long_url FROM url WHERE short_url = $1", short)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrURLNotFound
	}
	if err != nil {
		return "", fmt.Errorf("error while getting short url: %w", err)
	}
	logger.Sugaarz.Infow("Got short url from db", "short", short, "long", longURL)
	return longURL, nil
}

func (d *DataBase) Close() error {
	return d.db.Close()
}
