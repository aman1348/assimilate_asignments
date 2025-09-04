package config

import (
    // "database/sql"
    "fmt"
    "log"
    "os"

    "github.com/joho/godotenv"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
    _ = godotenv.Load()

    host := os.Getenv("DB_HOST")
    port := os.Getenv("DB_PORT")
    user := os.Getenv("DB_USER")
    // pass := os.Getenv("DB_PASSWORD")
    dbname := os.Getenv("DB_NAME")
    sslmode := os.Getenv("DB_SSLMODE")

    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s dbname=%s sslmode=%s",
        host, port, user, dbname, sslmode,
    )

    // sqlDB, err := sql.Open("postgres", dsn)
    // if err != nil {
    //     log.Fatalf("failed to open db: %v", err)
    // }
    // if err := sqlDB.Ping(); err != nil {
    //     log.Fatalf("failed to ping db: %v", err)
    // }

    // gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("failed to init gorm: %v", err)
    }

    DB = gormDB
}
