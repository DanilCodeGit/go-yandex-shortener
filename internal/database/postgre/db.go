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

func DBConn() error {
	//
	//db, err := sql.Open("pgx", *cfg.FlagDataBaseDSN)
	//if err != nil {
	//	panic(err)
	//}
	//
	//defer db.Close()
	//err = db.Ping()
	//if err != nil {
	//	log.Fatal("Неудачное подключение")
	//}
	//log.Println("Успешное подключение")
	//return err
	conn, err := pgx.Connect(context.Background(), *cfg.FlagDataBaseDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())
	log.Println("Успешное подключение")
	return err
}
