package main

import "fmt"
import "strings"

type Player struct {
	FirstName string
	LastName  string
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

func printHello(name string) {

	pl := Player{}

	pl.splitName(name)

	//pl.printNames()

	fmt.Printf(
		"Hello, %s%s\n",
		pl.FirstName,
		pl.LastName,
	)
}
