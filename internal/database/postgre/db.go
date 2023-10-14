package postgre

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/DanilCodeGit/go-yandex-shortener/internal/cfg"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func DBConn() (conn *pgxpool.Pool, err error) {
	conn, err = pgxpool.New(context.Background(), *cfg.FlagDataBaseDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v", err)
		return conn, err
	}
	//defer conn.Close(context.Background())
	log.Println("Успешное подключение")
	return conn, err
}

func CreateTable(conn *pgxpool.Pool) error {
	createTable := `CREATE TABLE IF NOT EXISTS short_urls (
   original_url TEXT NOT NULL,
   short_url VARCHAR(255) NOT NULL
)`
	_, err := conn.Exec(context.Background(), createTable)
	if err != nil {
		log.Println(err)
	}
	return err
}

func SaveShortenedURL(conn *pgxpool.Pool, originalURL, shortURL string) error {
	_, err := conn.Exec(context.Background(), "INSERT INTO short_urls (original_url, short_url) VALUES ($1, $2)", originalURL, shortURL)
	return err
}

//func GetShortenedURL(conn *pgx.Conn, id int) (string, error) {
//	var shortURL string
//	err := conn.QueryRow(context.Background(), "SELECT short_url FROM short_urls WHERE id = $1", id).Scan(&shortURL)
//	if err != nil {
//		return "", err
//	}
//	return shortURL, nil
//}
