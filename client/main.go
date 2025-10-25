package main

import (
	"log"

	"github.com/chzyer/readline"
)

func main() {
	rl, err := readline.New("termsg>>>")
	if err != nil {
		log.Fatal(err)
	}
	defer rl.Close()
	for {
		user_in, err := rl.Readline()
		if err != nil {
			break
		}
	}
}
