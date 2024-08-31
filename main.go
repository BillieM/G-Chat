package main

import gchat "g-chat/src"

/*
to implement

message sending
	- don't create 2 instances of a message sent by me
		- whether sent from web client or habbo itself
	- ability to choose whisper/ say/ shout
		- default to shout

shout/ whisper/ say received differentiation

avatars
	- clicking on avatar allows you to pick

mutexes to ensure concurrency stability

ability to view badges/ mottos in chat client

ability to assign a colour to a user from chat client
	- opens when clicking avatar
	- colour picker section
	- requests all available colours from backend
		- can then select the one you want

colour scheme creator
	- separate page, displays all colours
	- ability to add new, remove existing, or edit existing
	- ui:
		- (-) [rose-500] background: [colour slider] text: [colour slider] -> how it looks: [example message]
		- [Save]

*/

func main() {
	gchat.InitExt()
}
