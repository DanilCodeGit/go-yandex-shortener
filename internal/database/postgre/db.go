package postgre

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	//"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DataBase interface {
	CreateTable() error
	SaveShortenedURL(originalURL, shortURL string) (string, error)
	Close()
}

type DB struct {
	Conn *pgxpool.Pool
	mu   sync.RWMutex
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
func (db *DB) CreateTable(ctx context.Context) error {
	createTable := `CREATE TABLE  short_urls (
	  	original_url varchar(255) NOT NULL constraint original_url_key unique ,
	  	short_url VARCHAR(255) NOT NULL

)`
	_, err := db.Conn.Exec(ctx, createTable)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (db *DB) SaveShortenedURL(ctx context.Context, originalURL, shortURL string) (string, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	_, err := db.Conn.Exec(ctx,
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

func (db *DB) Close(ctx context.Context) {
	defer db.Conn.Close()
}

func (db *DB) SaveBatch(ctx context.Context, originalURL, shortURL string) (string, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	_, err := db.Conn.Exec(ctx,
		`INSERT INTO 
			short_urls (original_url, short_url) 
			VALUES 
			($1, $2)`,
		originalURL, shortURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return pgErr.Code, nil
		}
	}

	return "", err
}
