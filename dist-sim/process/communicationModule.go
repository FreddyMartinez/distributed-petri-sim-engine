package process

import (
	"centralsim"
	"fmt"
	"net"
	"petrisim/helpers"
	"petrisim/models"
)

type CommunicationModule struct {
	pId                   int                    // Id del proceso
	networkInfo           []models.ProcessInfo   // Información de toda la red de procesos
	transitionsMap        []models.TransitionMap // Información de las transiciones en cada nodo
	listener              net.Listener
	logger                *centralsim.Logger
	outgoingEventCh       chan centralsim.Event          // Canal para envío de Eventos
	incomingEventCh       chan centralsim.IncommingEvent // Canal para recibir eventos generados en otros procesos
	reqLookAheadCh        chan int                       // Canal para solicitar LookAhead a proceso precedente
	receiveLookAheadCh    chan centralsim.LookAhead      // Canal para recibir LookAhead de proceso precedente
	receiveLookAheadReqCh chan int                       // Recibe solicitud de LookAhead de proceso posterior
	sendLookAheadCh       chan centralsim.LookAhead      // Envía LookAhead propio a proceso posterior
	killChan              chan bool
}

func CreateCommunicationModule(
	pid int,
	network []models.ProcessInfo,
	transitions []models.TransitionMap,
	logger *centralsim.Logger,
	sendEventCh chan centralsim.Event,
	incomingEventCh chan centralsim.IncommingEvent,
	reqLookAheadCh chan int,
	receiveLACh chan centralsim.LookAhead,
	receiveLAReqCh chan int,
	sendLookAheadCh chan centralsim.LookAhead,
	killChan chan bool,
) *CommunicationModule {

	process := network[pid]
	listener, err := net.Listen("tcp", ":"+process.Port)
	if err != nil {
		panic(fmt.Sprintf("Server listen error %v", err))
	}

	cm := CommunicationModule{
		pId:                   pid,
		networkInfo:           network,
		transitionsMap:        transitions,
		listener:              listener,
		logger:                logger,
		outgoingEventCh:       sendEventCh,
		incomingEventCh:       incomingEventCh,
		reqLookAheadCh:        reqLookAheadCh,
		receiveLookAheadCh:    receiveLACh,
		receiveLookAheadReqCh: receiveLAReqCh,
		sendLookAheadCh:       sendLookAheadCh,
		killChan:              killChan,
	}

	// Se lanzan rutinas para enviar y recibir mensajes
	go cm.sender()
	go cm.receiver()
	return &cm
}

// Rutina encargada de los mensajes que entran
func (comMod *CommunicationModule) receiver() {
	for {
		data := new(models.Message)
		err := helpers.Receive(data, &comMod.listener, comMod.logger)
		if err != nil {
			panic(err)
		}

		switch data.MsgType {
		case models.MsgEvent: // Evento generado en otro proceso
			comMod.logger.Event.Println(
				fmt.Sprintf("EVENTO ENTRANTE DESDE PL%v, TRANSICIÓN: %v CTE: %v, TIEMPO: %v", data.Sender, data.Event.IiTransicion, data.Event.IiCte, data.Event.IiTiempo))
			comMod.logger.GoVectLog(
				fmt.Sprintf("EVENTO ENTRANTE DESDE PL%v, TRANSICIÓN: %v CTE: %v, TIEMPO: %v", data.Sender, data.Event.IiTransicion, data.Event.IiCte, data.Event.IiTiempo))
			incommingEv := centralsim.IncommingEvent{Event: data.Event, ProcessId: data.Sender}
			comMod.incomingEventCh <- incommingEv

		case models.MsgLookAheadRequest: // Otro proceso solicita Lookhead
			comMod.logger.Mark.Println(fmt.Sprintf("PL%v SOLICITA LOOKAHEAD", data.Sender))
			comMod.logger.GoVectLog(fmt.Sprintf("PL%v Solicita LookAhead", data.Sender))
			comMod.receiveLookAheadReqCh <- data.Sender

		case models.MsgLookAhead: // Recibe el LookAhead de otro proceso
			comMod.logger.GoVectLog(
				fmt.Sprintf("Recibe LookAhead de PL%v, TIEMPO: %v", data.Sender, data.Time))
			comMod.logger.Mark.Println(
				fmt.Sprintf("Recibe LookAhead de PL%v, TIEMPO: %v", data.Sender, data.Time))
			comMod.receiveLookAheadCh <- centralsim.LookAhead{Process: data.Sender, Time: data.Time}

		case models.MsgKill: // Mensaje de que la simulación ha terminado
			comMod.killChan <- true
		}
	}
}

// Rutina encargada de enviar mensajes a los otros procesos
func (comMod *CommunicationModule) sender() {
	for {
		select {
		case event := <-comMod.outgoingEventCh: // Evento que se debe propagar a otro PL
			processId := comMod.findProcessId(&event)
			msg := models.Message{MsgType: models.MsgEvent, Event: event, Sender: comMod.pId}

			if processId == -1 {
				panic("process not found")
			}

			proc := comMod.networkInfo[processId]
			comMod.logger.Event.Println(
				fmt.Sprintf("ENVIAR EVENTO A PL%v, TRANSICIÓN: %v CTE: %v, TIEMPO: %v", processId, event.IiTransicion, event.IiCte, event.IiTiempo))
			comMod.logger.GoVectLog(
				fmt.Sprintf("ENVIAR EVENTO A PL%v, TRANSICIÓN: %v CTE: %v, TIEMPO: %v", processId, event.IiTransicion, event.IiCte, event.IiTiempo))
			helpers.Send(msg, proc.Ip+":"+proc.Port, comMod.logger)

		case id := <-comMod.reqLookAheadCh: // El simulador solicita un LookAhead a otro proceso
			comMod.logger.Mark.Println(fmt.Sprintf("Solicita LookAhead a P%v", id))
			comMod.logger.GoVectLog(fmt.Sprintf("Solicita LookAhead a P%v", id))
			msg := models.Message{MsgType: models.MsgLookAheadRequest, Sender: comMod.pId}

			// Enviar solicitud al proceso precedente
			proc := comMod.networkInfo[id]
			helpers.Send(msg, proc.Ip+":"+proc.Port, comMod.logger)

		case la := <-comMod.sendLookAheadCh: // El proceso envía LookAhead calculado al proceso que lo solicita
			comMod.logger.GoVectLog(fmt.Sprintf("Envía LookAhead a P%v, con tiempo %v", la.Process, la.Time))
			comMod.logger.Mark.Println(fmt.Sprintf("Envía LookAhead a P%v, con tiempo %v", la.Process, la.Time))
			msg := models.Message{MsgType: models.MsgLookAhead, Time: la.Time, Sender: comMod.pId}
			proc := comMod.networkInfo[la.Process]
			helpers.Send(msg, proc.Ip+":"+proc.Port, comMod.logger)
		}
	}
}

// Permite encontrar el proceso al que se debe enviar un evento
func (comMod *CommunicationModule) findProcessId(event *centralsim.Event) int {
	globalId := -1 * (int(event.IiTransicion) + 1)
	event.IiTransicion = centralsim.IndLocalTrans(globalId)
	for i, p := range comMod.transitionsMap {
		for _, t := range p.Transitions {
			if t == globalId {
				return i
			}
		}
	}
	return -1
}

// Envía mensaje a otros procesos para terminar la tarea
func (comMod *CommunicationModule) killProcesses() {
	for i, proc := range comMod.networkInfo {
		if i != comMod.pId {
			msg := models.Message{MsgType: models.MsgKill}
			helpers.Send(msg, proc.Ip+":"+proc.Port, comMod.logger)
		}
	}
	comMod.killChan <- true
}
