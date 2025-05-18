package migrator

import (
	"errors"
	"os"

	"github.com/stlesnik/url_shortener/internal/logger"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const migrationsDir = "./internal/migrator/migrations"

func Run(dsn string) {
	m, err := migrate.New("file://"+migrationsDir, dsn)
	if err != nil {
		logger.Sugaarz.Errorw("migrate.New failed", "error", err)
		os.Exit(1)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Sugaarz.Errorw("migrations failed", "error", err)
		os.Exit(1)
	}
	logger.Sugaarz.Infow("migrations applied successfully")
}
