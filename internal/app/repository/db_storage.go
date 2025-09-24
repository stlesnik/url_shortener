package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/stlesnik/url_shortener/internal/app/models"
	"github.com/stlesnik/url_shortener/internal/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// ErrCodeUniqueViolation is the PostgreSQL error code for unique constraint violation.
const ErrCodeUniqueViolation = "23505"

// MaxOpenConns is the maximum number of open database connections.
const MaxOpenConns = 25

// MaxIdleConns is the maximum number of idle database connections.
const MaxIdleConns = 10

// MaxIdleTime is the maximum amount of time a connection may be idle.
const MaxIdleTime = 5 * time.Minute

// MaxConnLifetime is the maximum amount of time a connection may be reused.
const MaxConnLifetime = time.Hour

// DataBase represents a PostgreSQL database connection.
type DataBase struct {
	db *sqlx.DB
}

// NewDataBase creates a new DataBase instance with the given DSN.
func NewDataBase(dsn string) (*DataBase, error) {
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		logger.Sugaarz.Errorf("error while opening db: %w: %v", ErrOpenDB, err)
		return nil, fmt.Errorf("error while opening db: %w: %v", ErrOpenDB, err)
	}

	database := &DataBase{db: db}
	database.configureConnectionPool()

	return database, nil
}

// configureConnectionPool configures the database connection pool settings.
func (d *DataBase) configureConnectionPool() {
	d.db.SetMaxOpenConns(MaxOpenConns)
	d.db.SetMaxIdleConns(MaxIdleConns)
	d.db.SetConnMaxIdleTime(MaxIdleTime)
	d.db.SetConnMaxLifetime(MaxConnLifetime)
}

// Ping checks the database connection.
func (d *DataBase) Ping(ctx context.Context) error {
	if err := d.db.PingContext(ctx); err != nil {
		return fmt.Errorf("error while ping to db: %w: %v", ErrPingDB, err)
	}
	return nil
}

// SaveURL saves a URL mapping to the database.
func (d *DataBase) SaveURL(ctx context.Context, short string, long string, userID string) (bool, error) {
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

// URLPair represents a short and long URL pair.
type URLPair struct {
	URLHash string
	LongURL string
}

// SaveBatchURL saves a batch of URL pairs to the database.
func (d *DataBase) SaveBatchURL(ctx context.Context, batch []URLPair) error {
	tx, err := d.db.BeginTx(ctx, nil)
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

// GetURL retrieves a URL mapping from the database by short URL.
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

// GetURLList retrieves all URL mappings for a user from the database.
func (d *DataBase) GetURLList(ctx context.Context, userID string) ([]models.BaseURLDTO, error) {
	var data []models.BaseURLDTO
	err := d.db.SelectContext(ctx, &data, "SELECT original_url, short_url FROM url WHERE user_id = $1", userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrURLNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGetURLList, err)
	}
	logger.Sugaarz.Infow("Got urls list from db", "data", data, "userID", userID)
	return data, err
}

// DeleteURLList marks a list of URLs as deleted in the database.
func (d *DataBase) DeleteURLList(values []interface{}, placeholders []string) (int64, error) {
	query := fmt.Sprintf(`
       UPDATE url 
       SET is_deleted = TRUE 
       WHERE (user_id,short_url) in (%s)
    `, strings.Join(placeholders, ", "))

	result, err := d.db.Exec(query, values...)
	if err != nil {
		logger.Sugaarz.Errorf("\n\nerror while exec: %w\nquery=%v\nvalues=%v\n\n", err, query, strings.Trim(fmt.Sprintf("%v", values), "[]"))
		return 0, err
	}
	ra, err := result.RowsAffected()
	if err != nil {
		logger.Sugaarz.Errorf("\n\nerror while RowsAffected: %w\n\n", err)
		return 0, err
	}
	return ra, nil
}

// GetStats retrieves statistics about db.
func (d *DataBase) GetStats(ctx context.Context) (models.StatsDTO, error) {
	var stats models.StatsDTO
	err := d.db.GetContext(ctx, &stats.URLCount, `SELECT COUNT(*) FROM url GROUP BY short_url`)
	if err != nil {
		return stats, err
	}
	err = d.db.GetContext(ctx, &stats.UserCount, `SELECT COUNT(*) FROM url GROUP BY user_id`)
	if err != nil {
		return stats, err
	}
	return stats, nil
}

// Close closes the database connection.
func (d *DataBase) Close() error {
	return d.db.Close()
}
