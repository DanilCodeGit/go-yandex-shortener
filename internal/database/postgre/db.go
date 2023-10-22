package postgre

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var GlobalConn *DB

type DB struct {
	Conn *pgxpool.Pool
}

type Database interface {
	CreateTable() error
	SaveShortenedURL(originalURL, shortURL string) (string, error)
	Close()
}

func NewDataBase(ctx context.Context, dsn string) (*DB, error) {
	conn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v", err)
		return nil, err
	}
	log.Println("Успешное подключение")

	return &DB{Conn: conn}, nil

}

func (db *DB) CreateTable() error {
	createTable := `CREATE TABLE IF NOT EXISTS short_urls (
	  	original_url varchar(255) NOT NULL constraint original_url_key unique ,
	  	short_url VARCHAR(255) NOT NULL

)`

	_, err := db.Conn.Exec(context.Background(), createTable)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (db *DB) SaveShortenedURL(originalURL, shortURL string) (string, error) {
	_, err := db.Conn.Exec(context.Background(),
		`INSERT INTO 
			short_urls (original_url, short_url) 
			VALUES 
			($1, $2)`,
		originalURL, shortURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			fmt.Println(pgErr.Message)
			fmt.Println(pgErr.Code)
			return pgErr.Code, nil
		}
	}

	return "", err
}

func (db *DB) Close() {
	defer db.Conn.Close()
}

//func DBConn(ctx context.Context) (err error) {
//	conn, err := pgxpool.New(ctx, *cfg.FlagDataBaseDSN)
//	if err != nil {
//		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v", err)
//		return err
//	}
//	Conn = conn
//	log.Println("Успешное подключение")
//
//	return err
//}
//
//func CreateTable(conn *pgxpool.Pool) error {
//	createTable := `CREATE TABLE IF NOT EXISTS short_urls (
//	  	original_url varchar(255) NOT NULL constraint original_url_key unique ,
//	  	short_url VARCHAR(255) NOT NULL
//
//)`
//
//	_, err := conn.Exec(context.Background(), createTable)
//	if err != nil {
//		log.Println(err)
//	}
//	return err
//}
//
//func SaveShortenedURL(conn *pgxpool.Pool, originalURL, shortURL string) (code string) {
//	_, err := conn.Exec(context.Background(),
//		`INSERT INTO
//			short_urls (original_url, short_url)
//			VALUES
//			($1, $2)`,
//		originalURL, shortURL)
//	if err != nil {
//		var pgErr *pgconn.PgError
//		if errors.As(err, &pgErr) {
//			fmt.Println(pgErr.Message)
//			fmt.Println(pgErr.Code)
//			return pgErr.Code
//		}
//	}
//
//	return ""
//}
