package postgre

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/DanilCodeGit/go-yandex-shortener/internal/cfg"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func DBConn() (conn *pgx.Conn, err error) {
	conn, err = pgx.Connect(context.Background(), *cfg.FlagDataBaseDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		return conn, err
	}
	//defer conn.Close(context.Background())
	log.Println("Успешное подключение")
	return conn, err
}
