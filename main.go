package main

import "fmt"
import _ "github.com/go-sql-driver/mysql"
import "database/sql"
import "encoding/json"
import "io/ioutil"

type DbConfig struct {
	DbHost string
	DbUser string
	DbPass string
}

func main() {
	bstr, err := ioutil.ReadFile("db.json")

	if err != nil {
		fmt.Println("ERROR: couldn't read config")
		return
	}

	var cfg DbConfig

	err = json.Unmarshal(bstr, &cfg)
	if err != nil {
		fmt.Println("Error decoding config", err)
		return
	}

	db, err := sql.Open(
		"mysql",
		fmt.Sprintf(
			"%s:%s@tcp(%s:3306)/",
			cfg.DbUser,
			cfg.DbPass,
			cfg.DbHost,
		),
	)

	if err != nil {
		panic(err)
	}

	err = db.Ping()

	if err != nil {
		panic(err)
	}

	fmt.Println("connected to database")
}
