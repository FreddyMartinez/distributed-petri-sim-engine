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
			// Encuentra la transición de salida de la subred actual
			outboundTransition, err := se.findOutbound()
			if err != nil {
				panic(err)
			}

			var currentTransition int
			if se.lastEnabled == -1 { // Este caso se presenta si no ha comenzado la simulación
				se.Log.Mark.Println("Espera a que inicie la simulación")
				se.lastEnabled = <-se.simulationInit
			}
			se.mux.Lock() // Bloquea recursos para leer estado local

			currentTransition = se.lastEnabled
			se.Log.NoFmtLog.Println("Current", currentTransition)

			futureTime := se.iiRelojlocal + TypeClock(outboundTransition.TiempoHastaMarca.LiTiempos[currentTransition]-1)

			// Encuentra la constante que se envía al proceso que solicita el LookAhead
			trCo, err := getFutureEvent(outboundTransition.TransConstPul, lar.InputTransition)

			if err != nil {
				panic(err)
			}

			ev := Event{IiTiempo: futureTime, IiTransicion: IndLocalTrans(lar.InputTransition), IiCte: TypeConst(trCo[1])}
			la := LookAhead{Process: lar.Process, ExpectedEvent: ev}
			se.mux.Unlock()
			se.Log.NoFmtLog.Println("Envía LookAhead", ev)
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

func (se *SimulationEngine) getLookAhead() {
	// solicita Look Ahead
	se.reqLookAheadCh <- true
	// espera Look Ahead
	ev := <-se.receiveLookAheadCh
	// Si el evento es de un tiempo igual o mayor se inserta, si no, se vuelve a pedir
	if ev.getTiempo() >= se.iiRelojlocal {
		se.mux.Lock()
		se.IlEventos.inserta(ev)
		se.mux.Unlock()
	} else {
		se.getLookAhead()
	}
}
