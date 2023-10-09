package postgre

import (
	"database/sql"
	"log"

	"github.com/DanilCodeGit/go-yandex-shortener/internal/cfg"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func DBConn() error {

	db, err := sql.Open("pgx", *cfg.FlagDataBaseDSN)
	if err != nil {
		panic(err)
	}

	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatal("Неудачное подключение")
	}
	log.Println("Успешное подключение")
	return err
}
