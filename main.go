package main

import "fmt"
import _ "github.com/go-sql-driver/mysql"
import "database/sql"
import "encoding/json"
import "io/ioutil"
import "flag"
import "github.com/chzyer/readline"

type DbConfig struct {
	DbHost string
	DbUser string
	DbPass string
}

var server = flag.Bool("server", false, "starts text adventure as a telnet server")

var db *sql.DB

func main() {
	flag.Parse()
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

	db, err = sql.Open(
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

	if *server == true {
		panic("not implemented!")
	} else {
		startLocal()
		return
	}
}

func startLocal() {
	rl, err := readline.NewEx(
		&readline.Config{
			Prompt: "> ",
		},
	)

	if err != nil {
		panic(err)
	}

	g := NewLocalGame(rl)
	g.play()

}
