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
	roomID  int
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
			"(?i)(%s)\\s?(%s)?\\s?(%s)?\\s?(%s)?",
			allActions,
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
		roomID:  1,
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
	// use an array w/fixed size to guarantee that all parts will be there
	var matches [4]string

	for _, m := range parts {
		for i, v := range m {
			if i > 0 {
				parseDebug("%d %s", i, v)
				matches[i-1] = v
			}
		}
	}

	parseDebug("got all sentence parts: %s", parts)

	var object, prep, prepObj, verb string
	verb = matches[0]
	object = matches[1]
	prep = matches[2]
	prepObj = matches[3]

	parseDebug(
		"finished parsing, verb = %s, object = %s, prep = %s, pobj = %s",
		verb, object, prep, prepObj,
	)

	var canUse bool
	var resText, resAction string

	if verb != "" && object != "" {
		// if verb and object, we must check
		// that we can use the verb on the object
		canUse, resText, resAction = tryAction(object, verb, g.roomID)
		if !canUse {
			cmdParseFail()
			return
		}

		if resText != "" {
			fmt.Println(resText)
		}

		switch resAction {
		case "print-room-description":

		}
	}

}

// checks that you can perform an action on an item
// returns
// 	- whether you can use the item (bool)
// 	- what text to display (if any)
// 	- how to modify game state
func tryAction(action, item string, roomID int) (bool, string, string) {
	parseDebug("checking item usability")
	var err error
	row := db.QueryRow(
		`
			select 
				SUM(IFNULL(COALESCE(a1.ActionID, a2.ActionID), 0)) canUse,
				GROUP_CONCAT(distinct COALESCE(ria.ResponseText, ia.ResponseText, '')) ResponseText,
				GROUP_CONCAT(distinct COALESCE(ria.ResponseAction, ia.ResponseAction, a.ResponseAction, '')) ResponseAction
			from
				TextAdventure.Actions a 
				left join TextAdventure.Items_Actions ia on ia.ActionID = a.ActionID
				left join TextAdventure.Items i on i.ItemID = ia.ItemID and i.Item = ?

				left join TextAdventure.Items_ItemTypes iit on iit.ItemID = i.ItemID
				left join TextAdventure.Actions a1 on a1.ActionID = ia.ActionID and a1.Action = UPPER(?)
				left join TextAdventure.ItemTypes it on it.ItemTypeID = iit.ItemTypeID
				left join TextAdventure.ItemTypes_Actions ita on ita.ItemTypeID = it.ItemTypeID
				left join TextAdventure.Actions a2 on a2.ActionID = ita.ActionID and a2.Action = UPPER(?)

				left join TextAdventure.Rooms_Items_Actions ria on 
					ria.RoomID = ? and ria.ItemID = i.ItemID and ria.ActionID = coalesce(a1.ActionID, a2.ActionID, a.ActionID)
			where
				// (LOWER(i.Item) = ? or i.Item IS NULL)
				and a.Action = ?
			group by
				i.ItemID
		`,
		item,
		action,
		action,
		roomID,
		action,
	)

	if err != nil {
		fmt.Println("whoa, this query failed:")
		panic(err)
	}

	var canUse int
	var responseText, responseAction string
	err = row.Scan(&canUse, &responseText, &responseAction)

	parseDebug("ran can use check, got %d", canUse)

	if err != nil {
		fmt.Println("whoa, this query failed:")
		panic(err)
	}

	var ok bool
	if canUse > 0 {
		ok = true
	}

	return ok, responseText, responseAction
}

type room struct {
	name        string
	description string
	accessible  map[int]*room
	items       map[int]*item
}

func getRooms() []*room {
	allRooms := map[int]*room{}
	//allItems := map[int]*item{}
	res, err := db.Query(
		`
			select
				r.RoomID,
				r.Room,
				r.Description,
				i.Item,
				i.ItemID,
				rr.ToRoomID,
				tr.Room ToRoom,
				tr.Description ToDescription
			from 	
				TextAdventure.Rooms r 
				join TextAdventure.Rooms_Items ri using(RoomID)
				join TextAdventure.Room_Rooms rr on rr.FromRoomID = r.RoomID
				join TextAdventure.Rooms tr on tr.RoomID = rr.ToRoomID
		`,
	)

	if err != nil {
		fmt.Println("Error retrieving rooms")
		panic(err)
	}

	for res.Next() {
		var roomId, itemId, toRoomId int
		var name, description, itemName string
		res.Scan(&roomId, &name, &description, &itemName, &itemId, &toRoomId)

		if allRooms[roomId] == nil {
			allRooms[roomId] = &room{
				name:        name,
				description: description,
				accessible:  map[int]*room{},
				items:       map[int]*item{},
			}
		}

		room := allRooms[roomId]

		if itemId != 0 && room.items[itemId] == nil {
			room.items[itemId] = &item{
				ItemID: itemId,
				//Item:   item,
			}
		}

		if room.accessible[toRoomId] == nil {
			if allRooms[roomId] != nil {

			} else {

			}
		}
	}

	return nil
}
