package models

import "centralsim"

type Message struct {
	MsgType string
	event   centralsim.Event
}
