package models

import "centralsim"

const MsgLookAheadRequest = "LookAheadReq"
const MsgLookAhead = "LookAhead"
const MsgEvent = "Event"

type Message struct {
	MsgType string
	Sender  int
	Event   centralsim.Event
}
