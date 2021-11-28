package centralsim

import (
	"errors"
	"fmt"
)

type LookAhead struct {
	Process int
	Time    TypeClock
}

// Estructura utilitaria, usada para enviar el Evento con el id del proceso que lo generó
type IncommingEvent struct {
	Event     Event
	ProcessId int
}

// Rutina encargada de añadir eventos entrantes a la lista
func (se *SimulationEngine) manageIncommingEvents() {
	for {
		select {
		case event := <-se.incomEventsCh:
			se.mux.Lock()
			se.IlEventos.inserta(event.Event)
			if se.lookAheads[event.ProcessId] < event.Event.IiTiempo {
				se.lookAheads[event.ProcessId] = event.Event.IiTiempo
			}

			if se.isWaitingEvent {
				se.isWaitingEvent = false
				se.waitForEvent <- true
			}
			se.mux.Unlock()
		}
	}
}

// Rutina encargada de calcular el valor del LookAhead cuando lo requiere otro proceso
func (se *SimulationEngine) CalculateLookAhead() {
	for {
		select {
		case lar := <-se.receiveLookAheadReqCh:
			se.mux.Lock() // Bloquea recursos para leer estado local

			la := LookAhead{Process: lar}
			// si no hay transiciones sensibilizadas, da el tiempo futuro máximo
			if se.ilMislefs.IsTransSensib.isEmpty() {
				minLookAhead := TypeClock(15)
				for _, l := range se.lookAheads {
					// Aquí se podrían pedir LookAheads a precedentes para ampliar más el tiempo
					if l < minLookAhead {
						minLookAhead = l
					}
				}
				se.Log.NoFmtLog.Println("MinLook", minLookAhead)
				// El máximo es el LookAhead mínimo más el tiempo que le tome a un token atravesar la red
				la.Time = minLookAhead + se.maxLookAhead
			} else {
				la.Time = se.iiRelojlocal + 1 // asume que el tiempo mínimo en que puede generar un evento externo es 1
			}
			se.mux.Unlock()
			se.Log.Mark.Println(fmt.Sprintf("Envía LookAhead a P%v, con tiempo %v", la.Process, la.Time))
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

func (se *SimulationEngine) getLookAhead(processId int) {
	// solicita Look Ahead
	se.reqLookAheadCh <- processId
	// espera Look Ahead
	ev := <-se.receiveLookAheadCh
	// Si el evento es de un tiempo igual o mayor se inserta, si no, se vuelve a pedir
	if ev.Time >= se.iiRelojlocal {
		se.mux.Lock()
		se.lookAheads[ev.Process] = ev.Time
		se.mux.Unlock()
	} else {
		se.getLookAhead(processId)
	}
}
