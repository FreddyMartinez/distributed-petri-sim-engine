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
	outgoingEventCh       chan centralsim.Event            // Canal para envío de Eventos
	incomingEventCh       chan centralsim.Event            // Canal para recibir eventos generados en otros procesos
	reqLookAheadCh        chan bool                        // Canal para solicitar LookAhead a proceso precedente
	receiveLookAheadCh    chan centralsim.Event            // Canal para recibir LookAhead de proceso precedente
	receiveLookAheadReqCh chan centralsim.LookAheadRequest // Recibe solicitud de LookAhead de proceso posterior
	sendLookAheadCh       chan centralsim.LookAhead        // Envía LookAhead propio a proceso posterior
}

func CreateCommunicationModule(
	pid int,
	network []models.ProcessInfo,
	transitions []models.TransitionMap,
	logger *centralsim.Logger,
	sendEventCh chan centralsim.Event,
	incomingEventCh chan centralsim.Event,
	reqLookAheadCh chan bool,
	receiveLACh chan centralsim.Event,
	receiveLAReqCh chan centralsim.LookAheadRequest,
	sendLookAheadCh chan centralsim.LookAhead,
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
		err := helpers.Receive(data, &comMod.listener)
		if err != nil {
			panic(err)
		}

		switch data.MsgType {
		case models.MsgEvent: // Evento generado en otro proceso
			comMod.logger.Mark.Println(
				fmt.Sprintf("EVENTO ENTRANTE DESDE PL%v, TRANSICIÓN: %v CTE: %v, TIEMPO: %v", data.Sender, data.Event.IiTransicion, data.Event.IiCte, data.Event.IiTiempo))
			comMod.incomingEventCh <- data.Event
		case models.MsgLookAheadRequest: // Otro proceso solicita Lookhead
			comMod.logger.Mark.Println(fmt.Sprintf("PL%v SOLICITA LOOKAHEAD", data.Sender))
			inputTransition := comMod.transitionsMap[data.Sender].InputTrans
			la := centralsim.LookAheadRequest{Process: data.Sender, InputTransition: inputTransition}
			comMod.receiveLookAheadReqCh <- la
		case models.MsgLookAhead: // Recibe el LookAhead de otro proceso
			comMod.logger.Mark.Println(
				fmt.Sprintf("LOOKAHEAD ENVIADO POR PL%v, TRANSICIÓN: %v CTE: %v, TIEMPO: %v", data.Sender, data.Event.IiTransicion, data.Event.IiCte, data.Event.IiTiempo))
			comMod.receiveLookAheadCh <- data.Event
		}
	}
}

// Rutina encargada de enviar mensajes a los otros procesos
func (comMod *CommunicationModule) sender() {
	for {
		select {
		case event := <-comMod.outgoingEventCh: // Recibe Evento que se debe propagar
			processId := comMod.findProcessId(&event)
			msg := models.Message{MsgType: models.MsgEvent, Event: event}

			if processId == -1 {
				panic("process not found")
			}

			proc := comMod.networkInfo[processId]
			comMod.logger.Event.Println(
				fmt.Sprintf("ENVIAR EVENTO A PL%v, TRANSICIÓN: %v CTE: %v, TIEMPO: %v", processId, event.IiTransicion, event.IiCte, event.IiTiempo))
			helpers.Send(msg, proc.Ip+":"+proc.Port)

		case <-comMod.reqLookAheadCh: // El simulador solicita un LookAhead a otro(s) proceso(s)
			comMod.logger.Mark.Println("SOLICITA LOOKAHEAD A ", comMod.transitionsMap[comMod.pId].Ancestors)
			msg := models.Message{MsgType: models.MsgLookAheadRequest, Sender: comMod.pId}

			for _, ancestorId := range comMod.transitionsMap[comMod.pId].Ancestors { // Enviar solicitud a todos los procesos precedentes
				proc := comMod.networkInfo[ancestorId]
				helpers.Send(msg, proc.Ip+":"+proc.Port)
			}

		case la := <-comMod.sendLookAheadCh: // El proceso envía LookAhead calculado al proceso que lo solicita
			msg := models.Message{MsgType: models.MsgLookAhead, Event: la.ExpectedEvent, Sender: comMod.pId}
			proc := comMod.networkInfo[la.Process]
			helpers.Send(msg, proc.Ip+":"+proc.Port)
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
