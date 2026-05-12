package database

import (
	"database/sql"
	"fmt"
	"log"

	"fake-review-ai2/config"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init(cfg *config.Config) error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.DBName)
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	if err = DB.Ping(); err != nil {
		return err
	}
	log.Println("Connected to PostgreSQL")
	return nil
}
