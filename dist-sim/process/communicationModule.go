package process

import (
	"fmt"
	"net"
	"petrisim/helpers"
	"petrisim/models"
)

type CommunicationModule struct {
	NetworkInfo []models.ProcessInfo
	Listener    net.Listener
}

func CreateCommunicationModule(pid int, network []models.ProcessInfo) CommunicationModule {
	process := network[pid]
	listener, err := net.Listen("tcp", ":"+process.Port)
	if err != nil {
		panic(fmt.Sprintf("Server listen error %v", err))
	}

	cm := CommunicationModule{
		Listener: listener,
	}
	return cm
}

// This method should wait for messages from other processes
func (comMod *CommunicationModule) receiver() {
	for {
		data := new(models.Message)
		err := helpers.Receive(data, &comMod.Listener)
		if err != nil {
			panic(err)
		}

		// TODO: do something with the incoming message
	}
}
