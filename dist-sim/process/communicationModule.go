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
	outgoingEventCh       chan centralsim.Event // Canal para envío de Eventos
	incomingEventCh       chan centralsim.Event // Canal para recibir eventos generados en otros procesos
	reqLookAheadCh        chan bool             // Canal para solicitar LookAhead a proceso precedente
	receiveLookAheadCh    chan int              // Canal para recibir LookAhead de proceso precedente
	receiveLookAheadReqCh chan int              // Recibe solicitud de LookAhead de proceso posterior
	sendLookAheadCh       chan centralsim.Event // Envía LookAhead propio a proceso posterior
}

func CreateCommunicationModule(
	pid int,
	network []models.ProcessInfo,
	transitions []models.TransitionMap,
	logger *centralsim.Logger,
	sendEventCh chan centralsim.Event,
	incomingEventCh chan centralsim.Event,
	reqLookAheadCh chan bool,
	receiveLACh chan int,
	receiveLAReqCh chan int,
	sendLookAheadCh chan centralsim.Event,
) CommunicationModule {

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
	return cm
}

// Rutina encargada de los mensajes que entran
func (comMod *CommunicationModule) receiver() {
	for {
		data := new(models.Message)
		err := helpers.Receive(data, &comMod.listener)
		if err != nil {
			panic(err)
		}

		comMod.logger.Event.Println(
			fmt.Sprintf("EVENTO ENTRANTE DESDE PL%v, TRANSICIÓN: %v CTE: %v, TIEMPO: %v", data.Sender, data.Event.IiTransicion, data.Event.IiCte, data.Event.IiTiempo))

		switch data.MsgType {
		case models.MsgEvent: // Evento generado en otro proceso
			comMod.incomingEventCh <- data.Event
		case models.MsgLookAheadRequest: // Otro proceso solicita Lookhead
			comMod.receiveLookAheadReqCh <- data.Sender
		}
	}
}

// Rutina encargada de enviar mensajes a los otros procesos
func (comMod *CommunicationModule) sender() {
	for {
		select {
		case event := <-comMod.outgoingEventCh: // Recibe Evento que se debe propagar
			msg := models.Message{MsgType: models.MsgEvent, Event: event}
			processId := comMod.findProcessId(&event)

			if processId == -1 {
				panic("process not found")
			}

			proc := comMod.networkInfo[processId]
			comMod.logger.Event.Println(
				fmt.Sprintf("ENVIAR EVENTO A PL%v, TRANSICIÓN: %v CTE: %v, TIEMPO: %v", processId, event.IiTransicion, event.IiCte, event.IiTiempo))
			helpers.Send(msg, proc.Ip+":"+proc.Port)

		case <-comMod.reqLookAheadCh: // El simulador solicita un LookAhead a otro(s) proceso(s)
			comMod.logger.Mark.Println("SOLICITA LOOKAHEAD")
			msg := models.Message{MsgType: models.MsgLookAheadRequest, Sender: comMod.pId}
			comMod.logger.NoFmtLog.Println(msg)
		}
	}
}

func (comMod *CommunicationModule) findProcessId(event *centralsim.Event) int {
	globalId := -1 * (int(event.IiTransicion) + 1)
	event.IiTransicion = centralsim.IndLocalTrans(globalId)
	for i, p := range comMod.transitionsMap {
		for _, t := range p {
			if t == globalId {
				return i
			}
		}
	}
	return -1
}
