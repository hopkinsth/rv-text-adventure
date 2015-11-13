package main

import "github.com/chzyer/readline"
import "github.com/tj/go-debug"
import "io"
import "fmt"
import "strings"
import "database/sql"
import "math/rand"
import "regexp"

var validTokens *regexp.Regexp
var sentenceRegexp *regexp.Regexp
var parseDebug = debug.Debug("game.parse")
var allActions, allItems, allRooms string
var preps = "on|at|under|to"

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

	// TODO: this needs to go into a
	// more generic constructor func

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

	if validTokens == nil {

		r, err = db.Query(
			`
				select 
				  (select group_concat(Action separator '|') from TextAdventure.Actions) actions,
				  (select group_concat(Item separator '|') from TextAdventure.Items) item,
				  (select group_concat(Room separator '|') from TextAdventure.Rooms) rooms
			`,
		)

		var rstr string
		//var actions, items, rooms string
		for r.Next() {

			err = r.Scan(&allActions, &allItems, &allRooms)
			if err != nil {
				r.Close()
				goto lolwut
			}
		}

		rstr += allActions + "|" + allItems + "|" + allRooms
		rstr += "|" + preps
		r.Close()

		validTokens = regexp.MustCompile(
			fmt.Sprintf(
				"(?i)((%s)\\s{0,1})+",
				rstr,
			),
		)

		sentenceRegexp = regexp.MustCompile(
			// (verb) (prep.) (object) (prep.) (p. obj)
			fmt.Sprintf(
				"(?i)(%s)\\s?(%s)?\\s?(%s)?\\s?(%s)?\\s?(%s)?",
				allActions,
				preps,
				allItems,
				preps,
				allItems,
			),
		)
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

var failedCmds = [...]string{
	"What was that?",
	"Try again, please!",
	"First time using a keyboard?",
}

func cmdParseFail() {
	d := rand.Intn(len(failedCmds))
	fmt.Println(failedCmds[d])
}

// this needs to be something we can share
// between local and telnet games
func (g *localGame) parse(line string) {
	parseDebug("beginning line parse")
	// split this line
	allWords := strings.Split(line, " ")
	if len(allWords) == 1 && allWords[0] == "" {
		cmdParseFail()
		return
	}

	tokens := validTokens.FindAllString(line, -1)
	line = strings.Join(tokens, " ")

	if !sentenceRegexp.MatchString(line) {
		parseDebug("line prob. wasn't valid: \n%s\n", line)
		cmdParseFail()
		return
	}

	parts := sentenceRegexp.FindAllString(line, -1)

	// (verb) (prep.) (object) (prep.) (p. obj)
	verb := parts[0]

	var object, prep string

	for _, word := range parts[1:] {
		switch {
		case strings.Contains(preps, word):
			prep = word
		case strings.Contains(allItems, word):
			object = word
		}

		if object != "" && prep == "" {

		}
	}

}

func canUseItem(name string) {

}

var canUseActionQuery *sql.Stmt

func init() {
	var err error
	canUseActionQuery, err = db.Prepare(
		`
			select 
				IF(COUNT(ita.ItemTypeID) > 0), 
			from
				TextAdventure.Items i 
				join TextAdventure.Items_Actions ia using(ItemID)
				left join TextAdventure.ItemTypes it on it.ItemTypeID = i.ItemType
				left join TextAdventure.ItemTypes_Action ita on ita.ItemTypeID = it.ItemTypeID


		`,
	)
}
