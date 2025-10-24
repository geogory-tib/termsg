package main

import serverstate "server/server_state"

func main() {
	srvr := serverstate.Init()
	srvr.Server_Main()
}
