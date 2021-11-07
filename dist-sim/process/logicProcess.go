package process

import (
	"centralsim"
	"petrisim/models"
)

// This is the main structure
type LogicProcess struct {
	simEngine        centralsim.SimulationEngine
	communicationMod CommunicationModule
}

func CreateLogicProcess(pid int, network []models.ProcessInfo, netFileName string) *LogicProcess {
	lefs, err := centralsim.Load(netFileName)
	if err != nil {
		println("Couldn't load the Petri Net file !")
	}

	simEngine := centralsim.MakeMotorSimulation(lefs)
	comMod := CreateCommunicationModule(pid, network)
	lp := LogicProcess{
		simEngine:        simEngine,
		communicationMod: comMod,
	}
	return &lp
}
