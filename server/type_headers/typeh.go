package typeheaders

/*
This is the package that contains the type information of each JSON message sent to the server

example
{
	"type":"message" <- constants for this field
}
*/

const (
	LOGIN   = "login"
	MESSAGE = "message"
	COMMAND = "command"
)
