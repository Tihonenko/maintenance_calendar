package database

import (
	"belaz-calendar-server/internal/pkg/logger"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func ConnectDB() *sqlx.DB {

	if err := godotenv.Load(); err != nil {
		logger.Log.Fatal(".env не загружен", zap.Error(err))
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	db, err := sqlx.Connect("postgres", dsn)

	if err != nil {
		logger.Log.Fatal("Ошибка подключение к бд", zap.Error(err))
	}

	return db

}
