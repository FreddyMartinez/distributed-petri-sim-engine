package process

import (
	"centralsim"
	"petrisim/models"
	"strconv"
)

// This is the main structure
type LogicProcess struct {
	simEngine        *centralsim.SimulationEngine
	communicationMod *CommunicationModule
}

// Crea el contenedor del simulador y el módulo de comunicación
func CreateLogicProcess(pid int, network []models.ProcessInfo, netFileName string, transitions []models.TransitionMap) *LogicProcess {
	lefs, err := centralsim.Load(netFileName)
	if err != nil {
		println("Couldn't load the Petri Net file !")
	}

	sendEventCh := make(chan centralsim.Event)               // Canal para enviar eventos
	incomingEventCh := make(chan centralsim.Event)           // Canal para recibir eventos
	requestLookAheadCh := make(chan bool)                    // Canal para enviar solicitud de LA
	receiveLACh := make(chan centralsim.Event)               // Canal para recibir LA solicitado a otro proceso
	receiveLAReqCh := make(chan centralsim.LookAheadRequest) // Canal para recibir solicitud de LA
	sendLookAheadCh := make(chan centralsim.LookAhead)       // Canal para enviar LA a otro proceso

	logger := centralsim.CreateLogger(strconv.Itoa(pid))
	input := transitions[pid].InputTrans
	simEngine := centralsim.MakeSimulationEngine(lefs, logger, input, sendEventCh, incomingEventCh, requestLookAheadCh, receiveLACh, receiveLAReqCh, sendLookAheadCh)
	comMod := CreateCommunicationModule(pid, network, transitions, logger, sendEventCh, incomingEventCh, requestLookAheadCh, receiveLACh, receiveLAReqCh, sendLookAheadCh)
	lp := LogicProcess{
		simEngine:        simEngine,
		communicationMod: comMod,
	}
	return &lp
}

// Here we run the local simulation
func (LP *LogicProcess) RunSimulation(numberOfCycles int) {
	LP.simEngine.SimularPeriodo(0, centralsim.TypeClock(numberOfCycles))
}
