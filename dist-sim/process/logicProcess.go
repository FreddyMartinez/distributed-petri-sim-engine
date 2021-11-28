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
func CreateLogicProcess(pid int, network []models.ProcessInfo, netFileName string, transitions []models.TransitionMap, killChan chan bool) *LogicProcess {
	lefs, err := centralsim.Load(netFileName)
	if err != nil {
		println("Couldn't load the Petri Net file !")
	}

	sendEventCh := make(chan centralsim.Event)              // Canal para enviar eventos
	incomingEventCh := make(chan centralsim.IncommingEvent) // Canal para recibir eventos
	requestLookAheadCh := make(chan int)                    // Canal para enviar solicitud de LA
	receiveLACh := make(chan centralsim.LookAhead)          // Canal para recibir LA solicitado a otro proceso
	receiveLAReqCh := make(chan int)                        // Canal para recibir solicitud de LA
	sendLookAheadCh := make(chan centralsim.LookAhead)      // Canal para enviar LA a otro proceso
	maxLookAhead := centralsim.TypeClock(transitions[pid].MinTime)

	partnersLookAheads := make(map[int]centralsim.TypeClock)
	for _, a := range transitions[pid].Ancestors { // crea el mapa de LookAheads solo con los predecesores
		partnersLookAheads[a] = centralsim.TypeClock(0) // LookAheads se inicializan en cero
	}

	logger := centralsim.CreateLogger(strconv.Itoa(pid))
	simEngine := centralsim.MakeSimulationEngine(
		lefs,
		logger,
		sendEventCh,
		incomingEventCh,
		requestLookAheadCh,
		receiveLACh,
		receiveLAReqCh,
		sendLookAheadCh,
		partnersLookAheads,
		maxLookAhead)
	comMod := CreateCommunicationModule(
		pid,
		network,
		transitions,
		logger,
		sendEventCh,
		incomingEventCh,
		requestLookAheadCh,
		receiveLACh,
		receiveLAReqCh,
		sendLookAheadCh,
		killChan)
	lp := LogicProcess{
		simEngine:        simEngine,
		communicationMod: comMod,
	}
	return &lp
}

// Here we run the local simulation
func (LP *LogicProcess) RunSimulation(numberOfCycles int) {
	LP.simEngine.SimularPeriodo(0, centralsim.TypeClock(numberOfCycles))
	LP.communicationMod.killProcesses() // al terminar avisa a los demás procesos
}
