package main

import "github.com/chzyer/readline"
import "github.com/tj/go-debug"
import "io"
import "fmt"
import "strings"
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

		sentence := fmt.Sprintf(
			"(?i)(%s)\\s?(%s)?\\s?(%s)?\\s?(%s)?\\s?(%s)?",
			allActions,
			preps,
			allItems,
			preps,
			allItems,
		)

		parseDebug("sentence regexp:\n%s", sentence)

		sentenceRegexp = regexp.MustCompile(
			// (verb) (prep.) (object) (prep.) (p. obj)
			sentence,
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

	parts := sentenceRegexp.FindAllStringSubmatch(line, -1)

	// (verb) (prep.) (object) (prep.) (p. obj)

	parseDebug("got all sentence parts: %s", parts)

	var object, prep, prepObj string
	for _, parts := range parts {
		verb := parts[1]
		parseDebug("have %d words", len(parts))
		for _, word := range parts[2:] {
			word = strings.ToLower(word)
			//parseDebug("found us a word")
			switch {
			case strings.Contains(strings.ToLower(preps), word):
				parseDebug("got preposition")
				prep = word
			case strings.Contains(strings.ToLower(allItems), word):
				parseDebug("found an item/object")
				object = word
			case object != "" && prep != "":
				if !strings.Contains(strings.ToLower(allItems), word) {
					cmdParseFail()
					return
				} else {
					prepObj = word
				}
			}

			if object != "" && prep == "" {

				if !canUseItem(object, verb) {
					cmdParseFail()
					return
				}
			}
		}
	}

}

func canUseItem(item, action string) bool {
	parseDebug("checking item usability")
	var err error
	row := db.QueryRow(
		`
			select 
				SUM(IFNULL(COALESCE(a1.ActionID, a2.ActionID), 0)) canUse
			from
				TextAdventure.Items i
				left join TextAdventure.Items_Actions ia using(ItemID)
				left join TextAdventure.Items_ItemTypes iit on iit.ItemID = i.ItemID
				left join TextAdventure.Actions a1 on a1.ActionID = ia.ActionID and a1.Action = UPPER(?)
				left join TextAdventure.ItemTypes it on it.ItemTypeID = iit.ItemTypeID
				left join TextAdventure.ItemTypes_Actions ita on ita.ItemTypeID = it.ItemTypeID
				left join TextAdventure.Actions a2 on a2.ActionID = ita.ActionID and a2.Action = UPPER(?)
			where 
				LOWER(i.Item) = ?
			group by
				i.ItemID
		`,
		action,
		action,
		item,
	)

	if err != nil {
		fmt.Println("whoa, this query failed:")
		panic(err)
	}

	var canUse int
	err = row.Scan(&canUse)

	parseDebug("ran can use check, got %d", canUse)

	if err != nil {
		fmt.Println("whoa, this query failed:")
		panic(err)
	}

	if canUse > 0 {
		return true
	} else {
		return false
	}
}
