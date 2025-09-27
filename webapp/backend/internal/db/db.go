package db

import (
	"backend/internal/telemetry"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func InitDBConnection() (*sqlx.DB, error) {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		dbUrl = "user:password@tcp(db:4306)/42Tokyo2508-db"
	}
	dsn := fmt.Sprintf("%s?charset=utf8mb4&parseTime=True&loc=Local", dbUrl)
	log.Printf(dsn)

	driverName := telemetry.WrapSQLDriver("mysql")
	dbConn, err := sqlx.Open(driverName, dsn)
	if err != nil {
		log.Printf("Failed to open database connection: %v", err)
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = dbConn.PingContext(ctx)
	if err != nil {
		dbConn.Close()
		log.Printf("Failed to connect to database: %v", err)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Println("Successfully connected to MySQL!")

	// 高負荷時に柔軟に調整できるよう、接続プールは環境変数で上書き可能にする
	maxOpen := getEnvInt("DB_MAX_OPEN_CONNS", 100)
	maxIdle := getEnvInt("DB_MAX_IDLE_CONNS", 25)
	connLifetime := getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute)

	dbConn.SetMaxOpenConns(maxOpen)
	dbConn.SetMaxIdleConns(maxIdle)
	dbConn.SetConnMaxLifetime(connLifetime)

	return dbConn, nil
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
		return parsed
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	dur, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return dur
}
