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
func CheckDuplicate(ctx context.Context, conn *pgxpool.Pool, originalURL string) error {
	query := "SELECT COUNT(*) FROM short_urls WHERE original_url = $1"
	var count int
	err := conn.QueryRow(ctx, query, originalURL).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("данный url уже существует в БД")
	}

	return nil
}

func DeleteAllRecords(conn *pgxpool.Pool) error {
	// SQL-запрос для удаления всех записей из таблицы
	query := fmt.Sprintf("DELETE FROM short_urls")

	// Непосредственно выполняем SQL-запрос
	var _, err = conn.Exec(context.Background(), query)
	if err != nil {
		return err
	}

	fmt.Printf("Все записи из таблицы удалены успешно.\n")
	return nil
}
