package types

import "fmt"

/*
This protocol is super simple and json based because that's what everyone uses

example message to an other user

{
	"type":"message" ---  message is the default type for user messages
	"from":"User"
	"to":"User2"
	"contents":"Hello from across the web!"
	"time":"2025-10-23 00:45" -- whatever formating you choose to implement i chose ANSIC for the format
}

Example Server command -- these are commands users can ask the server ex: who's active on the current server
{
	"type":"command"
	"from":"User" -- this can be void as the handle_conn holds the current users connnection
	"contents":"show active" -- the command is stored with in the contents portion with the command and it's arguments split with spaces
	"time":"2025-10-23 00:45"

}
*/

type Message struct {
	Type     string `json:"type"`
	From     string `json:"from"`
	To       string `json:"to"`
	Contents string `json:"contents"`
	Time     string `json:"Time"`
}

type User struct {
	Name     string   `json:"name"`
	Password [32]byte `json:"password"`
}

func (self Message) Display_Message() {
	fmt.Println(self.Type)
	fmt.Println(self.From)
	fmt.Println(self.Contents)
	fmt.Println(self.Time)
}
