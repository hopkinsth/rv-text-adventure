package main

import "fmt"
import "strings"

type Actions map[int]string

// maps an item to its quantity
//type Inventory map[Item]int

// <VERB> [<object> | <direction>]
// 	 [<preposition> <p. obj.>]
//
// where prep. is one of
// TO, WITH, AT, ON, UNDER
//
//
// get badge
// look
// hug <person>
// throw bowling ball [ at pins | at basket ball hoop ]
// use shower
// put lettuce on sandwich
//
// show inventory
// inventory
// show me the goods
//
// HELP
// 		WTF? try do

type item struct {
	ItemID       int
	Item         int
	ValidActions []int
}

type InventoryItem struct {
	quantity int
	item
}

type Player struct {
	FirstName string
	LastName  string
	Inventory []InventoryItem
}

func (p *Player) printNames() {
	fmt.Printf(
		"fName: %s ; lName: %s\n",
		p.FirstName,
		p.LastName,
	)
}

func (p *Player) splitName(name string) (string, string) {
	parts := strings.Split(name, " ")

	var fName string
	var lName string

	if len(parts) == 0 {
		panic("invalid name")
	} else if len(parts) == 1 {
		fName = parts[0]
	} else {
		fName = parts[0]

		for i, v := range parts {
			if i > 0 {
				lName += " " + v
			}
		}
	}

	p.FirstName = fName
	p.LastName = strings.TrimLeft(lName, " ")

	return p.FirstName, p.LastName
}
