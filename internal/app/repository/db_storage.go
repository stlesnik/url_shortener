package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/stlesnik/url_shortener/internal/app/models"
	"github.com/stlesnik/url_shortener/internal/logger"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	ErrCodeUniqueViolation = "23505" // unique_violation
)

type DataBase struct {
	db *sqlx.DB
}

func NewDataBase(dsn string) (*DataBase, error) {
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		logger.Sugaarz.Errorf("error while opening db: %w: %v", ErrOpenDB, err)
		return nil, fmt.Errorf("error while opening db: %w: %v", ErrOpenDB, err)
	}
	return &DataBase{db: db}, nil
}

func (d *DataBase) Ping(ctx context.Context) error {
	if err := d.db.PingContext(ctx); err != nil {
		return fmt.Errorf("error while ping to db: %w: %v", ErrPingDB, err)
	}
	return nil
}

func (d *DataBase) SaveURL(ctx context.Context, short string, long string, userID string) (isDouble bool, err error) {
	_, dbErr := d.db.ExecContext(ctx, "INSERT INTO url (short_url, original_url, user_id) VALUES ($1, $2, $3)", short, long, userID)
	if dbErr != nil {
		var pgErr *pgconn.PgError
		if errors.As(dbErr, &pgErr) && pgErr.Code == ErrCodeUniqueViolation {
			logger.Sugaarz.Infow("this short url already exists", "short", short, "long", long)
			return true, nil
		}
		logger.Sugaarz.Errorf("%w: %v", ErrSaveURL, dbErr)
		return false, fmt.Errorf("%w: %v", ErrSaveURL, dbErr)
	}
	return false, nil
}

type URLPair struct {
	URLHash string
	LongURL string
}

func (d *DataBase) SaveBatchURL(ctx context.Context, batch []URLPair) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("error while beginning transaction: %w: %v", ErrBeginTransaction, err)
	}

	for _, pair := range batch {
		_, err := tx.ExecContext(ctx, ""+
			"INSERT INTO url (short_url, original_url) "+
			"VALUES ($1, $2) "+
			"ON CONFLICT (original_url) DO NOTHING", pair.URLHash, pair.LongURL)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("error while creating SQL statement in transaction: %w", err)
		}
	}

	return tx.Commit()
}

func (d *DataBase) GetURL(ctx context.Context, short string) (models.GetURLDTO, error) {
	var urlDTO models.GetURLDTO
	err := d.db.GetContext(ctx, &urlDTO, "SELECT original_url, is_deleted FROM url WHERE short_url = $1", short)
	if errors.Is(err, sql.ErrNoRows) {
		return models.GetURLDTO{}, ErrURLNotFound
	}
	if err != nil {
		return models.GetURLDTO{}, fmt.Errorf("error while getting short url: %w: %v", ErrGetURL, err)
	}
	logger.Sugaarz.Infow("Got short url from db", "short", short, "urlDTO", urlDTO)
	return urlDTO, nil
}

func (d *DataBase) GetURLList(ctx context.Context, userID string) (data []models.BaseURLDTO, err error) {
	err = d.db.SelectContext(ctx, &data, "SELECT original_url, short_url FROM url WHERE user_id = $1", userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrURLNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGetURLList, err)
	}
	logger.Sugaarz.Infow("Got urls list from db", "data", data, "userID", userID)
	return
}

func (d *DataBase) DeleteURLList(values []interface{}, placeholders []string) (int64, error) {
	query := fmt.Sprintf(`
		UPDATE url 
		SET is_deleted = TRUE 
		WHERE (user_id,short_url) in (%s)
	`, strings.Join(placeholders, ", "))

	result, err := d.db.Exec(query, values...)
	if err != nil {
		return 0, err
	}
	ra, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return ra, nil
}

func (d *DataBase) Close() error {
	return d.db.Close()
}
