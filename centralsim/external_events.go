package centralsim

import (
	"errors"
)

type LookAhead struct {
	Process       int
	ExpectedEvent Event
}

type LookAheadRequest struct {
	Process         int // id del proceso que solicita el LookAhead
	InputTransition int // Transición de entrada del proceso que solicita el lookAhead
}

// Rutina encargada de añadir eventos entrantes a la lista
func (se *SimulationEngine) manageIncommingEvents() {
	for {
		select {
		case event := <-se.incomEventsCh:
			se.mux.Lock()
			se.IlEventos.inserta(event)
			se.mux.Unlock()
		}
	}
}

// Rutina encargada de calcular el valor del LookAhead cuando lo requiere otro proceso
func (se *SimulationEngine) CalculateLookAhead() {
	for {
		select {
		case lar := <-se.receiveLookAheadReqCh:
			se.Log.Mark.Println("LookAhead Solicitado por", lar.Process)

			// Encuentra la transición de salida de la subred actual
			outboundTransition, err := se.findOutbound()
			if err != nil {
				panic(err)
			}

			se.mux.Lock()
			var currentTransition int
			if !se.ilMislefs.haySensibilizadas() { // Este caso se presenta si no ha comenzado la simulación, o si está esperando un LookAhead
				currentTransition = se.inputTr // Asume que está en la posición de entrada a la subred
			} else {
				enabledTransitions := se.ilMislefs.IsTransSensib
				currentTransition = int(enabledTransitions[0])
			}
			currentTransition -= int(se.ilMislefs.IaRed[0].IiIndLocal) // Normalizar el índice

			futureTime := se.iiRelojlocal + TypeClock(outboundTransition.TiempoHastaMarca.LiTiempos[currentTransition])

			// Encuentra la constante que se envía al proceso que solicita el LookAhead
			trCo, err := getFutureEvent(outboundTransition.TransConstPul, lar.InputTransition)

			if err != nil {
				panic(err)
			}

			ev := Event{IiTiempo: futureTime, IiTransicion: IndLocalTrans(lar.InputTransition), IiCte: TypeConst(trCo[1])}
			la := LookAhead{Process: lar.Process, ExpectedEvent: ev}
			se.mux.Unlock()
			se.sendLookAheadCh <- la
		}
	}
}

// Devuelve la transición de salida de la subred
func (se *SimulationEngine) findOutbound() (Transition, error) {
	for _, t := range se.ilMislefs.IaRed {
		if t.EsSalida {
			return t, nil
		}
	}
	return Transition{}, errors.New("Transición de salida no encontrada")
}

func getFutureEvent(constantsList [][2]int, inputTranId int) ([2]int, error) {
	for _, trCo := range constantsList { // Par transición / constante
		transitionId := -1 * (trCo[0] + 1) // Transforma el valor en un id global
		if transitionId == inputTranId {
			return trCo, nil
		}
	}
	return [2]int{}, errors.New("No se encontró la transición buscada")
}
