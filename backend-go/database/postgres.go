package database

import (
	"database/sql"
	"fmt"
	"log"

	"fake-review-ai2/config"

	_ "github.com/lib/pq"
)

type Store struct {
	DB *sql.DB
}

var Global *Store

func Init(cfg *config.Config) error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.DBName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}
	Global = &Store{DB: db}
	log.Println("Connected to PostgreSQL")
	return nil
}

func Close() error {
	if Global != nil && Global.DB != nil {
		return Global.DB.Close()
	}
	return nil
}
