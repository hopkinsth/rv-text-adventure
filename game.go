package main

import "github.com/chzyer/readline"
import "io"
import "fmt"

type Actions map[int]string

type game interface {
	play() error
}

type localGame struct {
	player  *Player
	actions Actions
	rl      *readline.Instance
}

func NewLocalGame(rl *readline.Instance) *localGame {
	fmt.Println("welcome to our game!")
	fmt.Println("What is your name?")

	line, err := rl.Readline()

	if err != nil {
		panic(err)
	}

	pl := &Player{}
	pl.splitName(line)

	// pre-fill with actions
	r, err := db.Query(`
		select 
			ActionID,
			Action
		from 
			TextAdventure.Actions
	`)
lolwut:
	if err != nil {
		fmt.Println("Failed getting actions!")
		panic(err)
	}

	actions := Actions{}
	for r.Next() {
		var id int
		var actionText string

		err = r.Scan(&id, &actionText)
		if err != nil {
			r.Close()
			goto lolwut
		}

		actions[id] = actionText
	}

	return &localGame{
		player:  pl,
		actions: actions,
		rl:      rl,
	}
}

func (g *localGame) play() {
	for {
		line, err := g.rl.Readline()

		if err == io.EOF {
			return
		}

		if err != nil {
			panic(err)
		}

		g.parse(line)
	}
}

// this needs to be something we can share
// between local and telnet games
func (g *localGame) parse(_ string) {
	return
}
