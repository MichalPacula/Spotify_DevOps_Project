package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func TestHttpSerrver(t *testing.T) {
	go SetupHttpServer()

	time.Sleep(5 * time.Second)
	resp, err := http.Get("http://127.0.0.1:8080")
	if err != nil {
		t.Errorf("Couldn't access http://127.0.0.1:8080: %s", err)
	}
	defer resp.Body.Close()
}

func TestDbConnection(t *testing.T) {
	db, err = sql.Open("mysql", fmt.Sprintf("root:%s@tcp(db:3306)/spotify_data_db", os.Getenv("MYSQL_ROOT_PASSWORD")))
	if err != nil {
		t.Errorf("Error connecting to db: %s", err)
	}
	defer db.Close()

	time.Sleep(5 * time.Second)
	_, err := db.Query("SHOW TABLES")
	if err != nil {
		t.Errorf("Couldn't retrieve tables from database: %s", err)
	}
}

func TestRedisConnection(t *testing.T) {
	SetupRedis()

	ctx = context.Background()

	_, err := redisdb.Keys(ctx, "*").Result()
	if err != nil {
		t.Errorf("Couldn't retrieve keys from redis: %s", err)
	}
}
