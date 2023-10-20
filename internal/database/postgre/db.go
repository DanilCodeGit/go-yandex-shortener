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

func DBConn(ctx context.Context) (conn *pgxpool.Pool, err error) {
	conn, err = pgxpool.New(context.Background(), *cfg.FlagDataBaseDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v", err)
		return conn, err
	}

	log.Println("Успешное подключение")
	return conn, err
}

func CreateTable(conn *pgxpool.Pool) error {
	createTable := `CREATE TABLE IF NOT EXISTS short_urls (
   	original_url varchar(255) NOT NULL constraint original_url_key unique ,
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

//func CheckDuplicate(conn *pgxpool.Pool, originalURL, shortURL string) error {
//	query := `
//        INSERT INTO short_urls (original_url, short_url)
//        VALUES ($1, $2)
//        ON CONFLICT (original_url) DO NOTHING
//    `
//	err := conn.QueryRow(context.Background(), query, originalURL, shortURL).Scan(&shortURL)
//	if err != nil {
//		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
//			// Возвращаем ошибку, указывая на конфликт ключа
//			return fmt.Errorf("Конфликт ключа: %v", err)
//		} else {
//			// Возвращаем другую ошибку
//			return fmt.Errorf("Произошла ошибка при выполнении запроса: %v", err)
//		}
//	}
//
//	// Если нет ошибок, возвращаем nil
//	return nil
//}
