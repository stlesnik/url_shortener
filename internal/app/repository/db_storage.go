package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stlesnik/url_shortener/internal/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DataBase struct {
	db *sql.DB
}

func NewDataBase(dsn string) (*DataBase, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("got error while connecting to db: %w", err)
	}
	err = warmupDB(db)
	if err != nil {
		return nil, err
	}
	return &DataBase{db: db}, nil
}

func warmupDB(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), "CREATE TABLE IF NOT EXISTS url(id serial primary key, short_url varchar not null unique,long_url varchar not null)")
	if err != nil {
		logger.Sugaarz.Errorf("error while warming db up: %w", err)
		return fmt.Errorf("error while warming db up: %w", err)
	}
	return nil
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

func (d *DataBase) Save(short string, long string) error {
	_, err := d.db.ExecContext(context.Background(), "INSERT INTO url (short_url, long_url) VALUES ($1, $2)", short, long)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			logger.Sugaarz.Infow("this short url already exists", "short", short, "long", long)
			return nil
		}
		logger.Sugaarz.Errorf("error while saving url: %w", err)
		return fmt.Errorf("error while saving url: %w", err)
	}
	return nil
}

func (d *DataBase) Get(short string) (string, bool) {
	var longURL string
	err := d.db.QueryRowContext(context.Background(), "SELECT long_url FROM url WHERE short_url = $1", short).Scan(&longURL)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false
	}
	if err != nil {
		panic(err)
	}
	logger.Sugaarz.Infow("Got short url from db", "short", short, "long", longURL)
	return longURL, true
}

func (d *DataBase) Close() error {
	return d.db.Close()
}
