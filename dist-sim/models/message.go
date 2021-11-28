package models

import "centralsim"

const MsgLookAheadRequest = "LookAheadReq"
const MsgLookAhead = "LookAhead"
const MsgEvent = "Event"
const MsgKill = "Kill"

type Message struct {
	MsgType string
	Sender  int
	Event   centralsim.Event
	Time    centralsim.TypeClock
}
