// Package db contains the models for the ghenga database.
package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	// import the database driver
	_ "github.com/lib/pq"

	"github.com/jmoiron/modl"
	"github.com/jmoiron/sqlx"
	"github.com/rubenv/modl-migrate"
)

const dialect = "postgres"

func init() {
	// use our own name mapper
	sqlx.NameMapper = ToSnakeCase
}

func findMigrationsDir() (dir string, err error) {
	subdir := filepath.Join("migrations", dialect)

	dir, err = os.Getwd()
	if err != nil {
		return "", err
	}

	for dir != "" {
		d := filepath.Join(dir, subdir)
		if fi, err := os.Stat(d); err == nil && fi.Mode().IsDir() {
			return d, nil
		}

		dir = filepath.Dir(dir)
	}

	return "", fmt.Errorf("directory %q not found", subdir)
}

// Database is a persistent storage database for ghenga.
type Database struct {
	dbmap *modl.DbMap
}

// configDBMap creates a new mapping on the given database and creates the
// tables (if necessary).
func configDBMap(db *sql.DB) (*modl.DbMap, error) {
	dbmap := modl.NewDbMap(db, modl.PostgresDialect{})
	dbmap.AddTableWithName(Person{}, "people").SetKeys(true, "id")
	dbmap.AddTableWithName(PhoneNumber{}, "phone_numbers").SetKeys(true, "id")
	dbmap.AddTableWithName(User{}, "users").SetKeys(true, "id")
	dbmap.AddTableWithName(Session{}, "sessions").SetKeys(false, "token")

	return dbmap, nil
}

// Open opens a PostgreSQL database. When the environment variable `DBTRACE`
// is set to 1, all queries are written to stderr.
func Open(dataSource string) (*Database, error) {
	db, err := sql.Open(dialect, dataSource)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	dbmap, err := configDBMap(db)
	if err != nil {
		return nil, err
	}

	if os.Getenv("DBTRACE") != "" {
		l := log.New(os.Stderr, "", log.LstdFlags)
		l.Printf("tracing database queries, data source is %q", dataSource)
		dbmap.TraceOn("DB: ", l)
	}

	if err = migrateDB(dbmap); err != nil {
		return nil, err
	}

	return &Database{dbmap}, nil
}

// migrateDB applies migrations according to the files in the subdir
// "migrations/".
func migrateDB(db *modl.DbMap) error {
	dir, err := findMigrationsDir()
	if err != nil {
		return err
	}

	src := &migrate.FileMigrationSource{Dir: dir}

	_, err = migrate.Exec(db.Db, dialect, src, migrate.Up)
	return err
}

// Close closes the connection to the underlying database.
func (db *Database) Close() error {
	return db.dbmap.Db.Close()
}
