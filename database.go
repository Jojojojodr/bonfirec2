package bonfirec2

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Jojojojodr/bonfirec2/config"
	"github.com/Jojojojodr/bonfirec2/models"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var Data *Database

type Database struct {
	db *gorm.DB
}

func (d *Database) GetDB() *gorm.DB {
	return d.db
}

func (d *Database) Connect() error {
	var db *gorm.DB
	var err error
	
	switch config.AppConfig.Database.Type {
	case "sqlite":
		db, err = connectSQLite(config.AppConfig.Database.DSN)
		if err != nil {
			return err
		}
	case "postgres":
		db, err = connectPostgres(config.AppConfig.Database.DSN)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported database type: %s", config.AppConfig.Database.Type)
	}
	
	d.db = db
	Data = d
	
	err = d.db.AutoMigrate(
		&Listener{},
		&Grunt{},
		&models.Message{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate database: %v", err)
	}
	
	log.Println("Database connection established and auto-migration completed")
	return nil
}

func NewDatabase() *Database {
	return &Database{db: nil}
}

func connectSQLite(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		dsn = "./data/bonfire.db" // Default SQLite database file
	}
	dir := strings.Split(dsn, "/")
	if len(dir) > 1 {
		dirPath := strings.Join(dir[:len(dir)-1], "/")
		if _, err := os.Stat(dirPath); err != nil {
			if os.IsNotExist(err) {
				if mkErr := os.MkdirAll(dirPath, 0o755); mkErr != nil {
					return nil, fmt.Errorf("failed to create directory for SQLite database: %w", mkErr)
				}
			} else {
				return nil, fmt.Errorf("failed to access directory for SQLite database: %w", err)
			}
		}
	}
	return gorm.Open(sqlite.Open(dsn), &gorm.Config{})
}

func connectPostgres(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("PostgreSQL DSN is required\nExample: postgres://user:password@localhost:5432/bonfire")
	}
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
